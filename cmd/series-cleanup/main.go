// Package main implements all application logic for series-cleanup
package main

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bjw-s/series-cleanup/internal/config"
	"github.com/bjw-s/series-cleanup/internal/helpers"
	"github.com/bjw-s/series-cleanup/internal/logger"
	"github.com/bjw-s/series-cleanup/internal/mediafile"
	"github.com/bjw-s/series-cleanup/internal/trakt"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func main() {
	logger.SetLevel(config.Config.LogLevel)

	logger.Debug("Loaded configuration",
		zap.Any("configuration", config.Config),
	)

	// Initialize Trakt API
	var traktAPI = trakt.API{}
	traktAPI.ClientID = config.Config.Trakt.ClientID
	traktAPI.ClientSecret = string(config.Config.Trakt.ClientSecret)
	traktAPI.DataPath = config.Config.Trakt.CacheFolder

	if err := traktAPI.Authenticate(); err != nil {
		logger.Fatal("Could not authenticate with Trakt",
			zap.Error(err),
		)
	}

	logger.Info("Successfully authenticated with Trakt")

	// Get User data from Trakt
	var traktUser = trakt.User{}
	traktUser.Name = config.Config.Trakt.User

	if err := traktUser.GetWatchedShows(traktAPI); err != nil {
		logger.Fatal("Could not get watched shows from Trakt",
			zap.Error(err),
		)
	}

	for _, scanFolder := range config.Config.ScanFolders {
		logger.Info("Processing...",
			zap.String("folder", scanFolder),
		)

		if !helpers.FolderExists(scanFolder) {
			logger.Fatal("Folder does not exist",
				zap.String("folder", scanFolder),
			)
		}

		tvShowFiles, err := collectTvShowFiles(scanFolder)
		if err != nil {
			logger.Fatal("Could not collect TV show files",
				zap.Error(err),
			)
		}

		tvShowFilesLength := len(tvShowFiles)
		var waitgroup sync.WaitGroup
		waitgroup.Add(tvShowFilesLength)

		for i := 0; i < tvShowFilesLength; i++ {
			go func(i int) {
				defer waitgroup.Done()
				err := processTvShowFile(tvShowFiles[i], &traktUser)
				if err != nil {
					logger.Fatal("Could not process TV show file",
						zap.Error(err),
					)
				}
			}(i)
		}

		waitgroup.Wait()
	}

	logger.Info("Finished...")
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

		file, err := mediafile.NewTVShowFile(path, config.Config.FolderRegex)
		if err != nil {
			return err
		}
		if file != nil {
			// Add mappings
			skipShow := false
			var skipSeasons []int
			for _, item := range config.Config.Overrides {
				parentFolderName := filepath.Base(filepath.Dir(file.Dir))
				if strings.EqualFold(parentFolderName, item.Folder) {
					file.Mappings = item.Mapping
					skipShow = item.Skip
					skipSeasons = item.SkipSeasons
					break
				}
			}

			if skipShow {
				logger.Debug("Skipped",
					zap.String("show", file.Show),
					zap.String("file", file.Filename),
					zap.String("reason", "Show is configured to be skipped"),
				)
				return nil
			}

			if lo.Contains(skipSeasons, file.Season) {
				logger.Debug("Skipped",
					zap.String("show", file.Show),
					zap.String("file", file.Filename),
					zap.String("reason", "Season is configured to be skipped"),
				)
				return nil
			}

			tvShowFiles = append(tvShowFiles, file)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tvShowFiles, nil
}

func processTvShowFile(mediafile *mediafile.TVShowFile, user *trakt.User) error {
	logger.Debug("Processing tv show file",
		zap.String("dir", mediafile.Dir),
		zap.String("file", mediafile.Filename),
	)

	var watchedShow *trakt.WatchedShow
	if mediafile.Mappings.IMDBID != "" {
		watchedShow = user.FindWatchedShowByIMDBID(mediafile.Mappings.IMDBID)
	} else if mediafile.Mappings.TVDBID != 0 {
		watchedShow = user.FindWatchedShowByTVDBID(mediafile.Mappings.TVDBID)
	} else if mediafile.Mappings.TraktName != "" {
		watchedShow = user.FindWatchedShowByName(mediafile.Mappings.TraktName)
	} else {
		watchedShow = user.FindWatchedShowByName(mediafile.Show)
	}

	if watchedShow == nil {
		logger.Debug("Skipped",
			zap.String("show", mediafile.Show),
			zap.String("file", mediafile.Filename),
			zap.String("reason", "Show is unwatched or could not be found"),
		)
		return nil
	}

	season := watchedShow.FindSeason(mediafile.Season)
	if season == nil {
		logger.Debug("Skipped",
			zap.String("show", watchedShow.Show.Title),
			zap.String("file", mediafile.Filename),
			zap.String("reason", "Season is unwatched"),
		)
		return nil
	}

	episode := season.FindEpisode(mediafile.Episode)
	if episode == nil {
		logger.Debug("Skipped",
			zap.String("show", watchedShow.Show.Title),
			zap.String("file", mediafile.Filename),
			zap.String("reason", "Episode is unwatched"),
		)
		return nil
	}

	watchedBeforeTime := time.Now().Add(-time.Duration(int64(config.Config.DeleteAfterHours) * int64(time.Hour)))
	if episode.LastWatchedBefore(watchedBeforeTime) {
		if config.Config.DryRun {
			logger.Info("TV show file would have been removed",
				zap.String("dir", mediafile.Dir),
				zap.String("file", mediafile.Filename),
			)
		} else {
			logger.Info("Removing tv show file",
				zap.String("dir", mediafile.Dir),
				zap.String("file", mediafile.Filename),
			)
			err := mediafile.DeleteWithSubtitleFiles()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
