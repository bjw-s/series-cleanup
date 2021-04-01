package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/bjw-s/seriescleanup/internal/helpers"
	"github.com/bjw-s/seriescleanup/internal/mediafile"
	"github.com/bjw-s/seriescleanup/internal/trakt"
)

type config struct {
	General struct {
		DryRun           bool
		DeleteAfterHours int
	}
	Trakt struct {
		ClientID     string
		ClientSecret string
		User         string
	}
	Logging struct {
		Loglevel string
	}
	ScanFolders []string
	Overrides   []struct {
		Folder  string
		Mapping mediafile.ShowMapping
		Skip    bool
	}
}

const scriptName = "seriescleanup"

var configData config
var configFile string
var dataFolder string

func main() {
	flag.StringVar(&configFile, "c", "settings.json", "Specify config file to use")
	flag.StringVar(&dataFolder, "d", "/data", "Specify folder to store data")
	flag.Parse()

	log.WithFields(log.Fields{
		"configFile": configFile,
		"dataFolder": dataFolder,
	}).Info("Running...")

	// Check pre-requisites
	if !helpers.FolderExists(dataFolder) {
		log.WithFields(log.Fields{
			"folder": dataFolder,
		}).Fatal("Could not find data folder")
	}

	if !helpers.FileExists(configFile) {
		log.WithFields(log.Fields{
			"file": configFile,
		}).Fatal("Could not find configuration file")
	}

	// Read configuration file
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.WithFields(log.Fields{
			"file":    configFile,
			"message": err.Error(),
		}).Fatal("Could not read file")
	}
	configData = config{}

	err = json.Unmarshal([]byte(file), &configData)
	if err != nil {
		log.WithFields(log.Fields{
			"file":    configFile,
			"message": err.Error(),
		}).Fatal("Could not parse config file")
	}

	if configData.Trakt.ClientID == "" {
		traktClientID, exists := os.LookupEnv("TRAKT_CLIENT_ID")
		if exists {
			configData.Trakt.ClientID = traktClientID
		} else {
			log.Fatal("No Trakt Client ID set")
		}
	}

	if configData.Trakt.ClientSecret == "" {
		traktClientSecret, exists := os.LookupEnv("TRAKT_CLIENT_SECRET")
		if exists {
			configData.Trakt.ClientSecret = traktClientSecret
		} else {
			log.Fatal("No Trakt Client Secret set")
		}
	}

	log.SetLevelFromString(configData.Logging.Loglevel)

	// Initialize Trakt API
	var traktAPI = trakt.API{}
	traktAPI.ClientID = configData.Trakt.ClientID
	traktAPI.ClientSecret = configData.Trakt.ClientSecret
	traktAPI.DataPath = dataFolder

	err = traktAPI.Authenticate()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Could not authenticate with Trakt")
	}

	log.Info("Successfully authenticated with Trakt")

	// Get User data from Trakt
	var traktUser = trakt.User{}
	traktUser.Name = configData.Trakt.User

	traktUser.GetWatchedShows(traktAPI)

	for _, scanFolder := range configData.ScanFolders {
		log.WithFields(log.Fields{
			"folder": scanFolder,
		}).Info("Processing...")

		if !helpers.FolderExists(scanFolder) {
			log.WithFields(log.Fields{
				"folder": scanFolder,
			}).Fatal("Folder does not exist")
		}

		tvShowFiles, err := collectTvShowFiles(scanFolder)
		if err != nil {
			log.Fatal(err.Error())
		}

		tvShowFilesLength := len(tvShowFiles)
		var waitgroup sync.WaitGroup
		waitgroup.Add(tvShowFilesLength)

		for i := 0; i < tvShowFilesLength; i++ {
			go func(i int) {
				defer waitgroup.Done()
				err := processTvShowFile(tvShowFiles[i], &traktUser)
				if err != nil {
					log.Fatal(err.Error())
				}
			}(i)
		}

		waitgroup.Wait()
	}

	log.WithFields(log.Fields{
		"script": scriptName,
	}).Info("Finished...")
}

func collectTvShowFiles(scanFolder string) ([]*mediafile.TVShowFile, error) {
	var tvShowFiles []*mediafile.TVShowFile
	err := filepath.Walk(scanFolder, func(path string, info os.FileInfo, nestedErr error) error {
		if info.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)
		if strings.HasPrefix(fileName, ".") {
			return nil
		}

		if !mediafile.IsMediaFile(path) {
			return nil
		}

		file, err := mediafile.NewTVShowFile(path)
		if err != nil {
			return err
		}
		if file != nil {
			// Add mappings
			skipShow := false
			for _, item := range configData.Overrides {
				if strings.EqualFold(file.ShowDir, item.Folder) {
					file.Mappings = item.Mapping
					skipShow = item.Skip
					break
				}
			}
			if !skipShow {
				tvShowFiles = append(tvShowFiles, file)
			} else {
				log.WithFields(log.Fields{
					"show": file.ShowDir,
					"file": file.Filename,
				}).Debug("Show is configured to be skipped...")
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tvShowFiles, nil
}

func processTvShowFile(mediafile *mediafile.TVShowFile, user *trakt.User) error {
	log.WithFields(log.Fields{
		"dir":  mediafile.Dir,
		"file": mediafile.Filename,
	}).Debug("Processing tv show file")

	var watchedShow *trakt.WatchedShow
	if mediafile.Mappings.TVDBID != 0 {
		watchedShow = user.FindWatchedShowByTVDBID(mediafile.Mappings.TVDBID)
	} else if mediafile.Mappings.TraktName != "" {
		watchedShow = user.FindWatchedShowByName(mediafile.Mappings.TraktName)
	} else {
		watchedShow = user.FindWatchedShowByName(mediafile.ShowDir)
	}

	if watchedShow == nil {
		log.WithFields(log.Fields{
			"show": mediafile.ShowDir,
		}).Debug("Show is unwatched or could not be found, skipping...")
		return nil
	}

	season := watchedShow.FindSeason(mediafile.Season)
	if season == nil {
		log.WithFields(log.Fields{
			"show":   watchedShow.Show.Title,
			"season": mediafile.Season,
		}).Debug("Season is unwatched, skipping...")
		return nil
	}

	episode := season.FindEpisode(mediafile.Episode)
	if episode == nil {
		log.WithFields(log.Fields{
			"show":    watchedShow.Show.Title,
			"season":  mediafile.Season,
			"episode": mediafile.Episode,
		}).Debug("Episode is unwatched, skipping...")
		return nil
	}

	watchedBeforeTime := time.Now().Add(-time.Duration(int64(configData.General.DeleteAfterHours) * int64(time.Hour)))
	if episode.LastWatchedBefore(watchedBeforeTime) {
		if configData.General.DryRun {
			log.WithFields(log.Fields{
				"dir":  mediafile.Dir,
				"file": mediafile.Filename,
			}).Info("Tv show file would have been removed")
		} else {
			log.WithFields(log.Fields{
				"dir":  mediafile.Dir,
				"file": mediafile.Filename,
			}).Info("Removing tv show file")
			mediafile.DeleteWithSubtitleFiles()
		}
	}

	return nil
}
