package tvshow

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/bjw-s/series-cleanup/internal/config"
	"github.com/oriser/regroup"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

var mediaFileExtensions = []string{".avi", ".mkv", ".mp4"}
var subtitleFileExtensions = []string{".srt"}

// File represents a tv show file on disk
type File struct {
	path string

	Dir           string
	Filename      string
	Extension     string
	subtitleFiles []string

	Season  int
	Episode int

	Rules config.FolderRules

	Identifiers Identifiers
}

func (tvShowFile *File) determineShow(regex string) error {
	show := filepath.Base(filepath.Dir(tvShowFile.Dir))
	if len(regex) > 0 {
		folderRegex := regroup.MustCompile(regex)
		matches, _ := folderRegex.Groups(show)

		if matchShow, ok := matches["Show"]; ok {
			show = matchShow
		}

		if matchIMDBID, ok := matches["IMDBID"]; ok {
			tvShowFile.Identifiers.IMDB = matchIMDBID
		}

		if matchTVBID, ok := matches["TVDBID"]; ok {
			tvdbID, err := strconv.Atoi(matchTVBID)
			if err != nil {
				return err
			}
			tvShowFile.Identifiers.TVDB = tvdbID
		}
	}
	tvShowFile.Identifiers.Name = show
	return nil
}

func (tvShowFile *File) determineSeason() error {
	r, err := regexp.Compile(`.*[sS](\d+)[eE]\d+.*`)
	if err != nil {
		return err
	}

	result := r.FindAllStringSubmatch(tvShowFile.Filename, -1)
	season, err := strconv.Atoi(result[0][1])
	if err != nil {
		return err
	}
	tvShowFile.Season = season
	return nil
}

func (tvShowFile *File) determineEpisode() error {
	r, err := regexp.Compile(`.*[sS]\d+[eE](\d+).*`)
	if err != nil {
		return err
	}

	result := r.FindAllStringSubmatch(tvShowFile.Filename, -1)
	episode, err := strconv.Atoi(result[0][1])
	if err != nil {
		return err
	}
	tvShowFile.Episode = episode
	return nil
}

func (tvShowFile *File) processFolderSettings() error {
	settings, found := lo.Find(config.AppConfig.FolderSettings, func(item config.FolderSettings) bool {
		parentFolderName := filepath.Base(filepath.Dir(tvShowFile.Dir))
		return strings.EqualFold(parentFolderName, item.Folder)
	})

	if found {
		if settings.Identifiers.IMDB != "" {
			tvShowFile.Identifiers.IMDB = settings.Identifiers.IMDB
		}
		if settings.Identifiers.Slug != "" {
			tvShowFile.Identifiers.Slug = settings.Identifiers.Slug
		}
		if settings.Identifiers.TMDB != 0 {
			tvShowFile.Identifiers.TMDB = settings.Identifiers.TMDB
		}
		if settings.Identifiers.Trakt != 0 {
			tvShowFile.Identifiers.Trakt = settings.Identifiers.Trakt
		}
		if settings.Identifiers.TVDB != 0 {
			tvShowFile.Identifiers.TVDB = settings.Identifiers.TVDB
		}

		tvShowFile.Rules = settings.Rules
	}
	return nil
}

func (tvShowFile *File) getBasicFileData() error {
	tvShowFile.Dir = filepath.Dir(tvShowFile.path)
	tvShowFile.Filename = filepath.Base(tvShowFile.path)
	tvShowFile.Extension = filepath.Ext(tvShowFile.path)
	return nil
}

func (tvShowFile *File) getSubtitleFiles() error {
	fileBaseName := strings.TrimSuffix(tvShowFile.Filename, tvShowFile.Extension)
	filesInSameFolder, err := os.ReadDir(tvShowFile.Dir)
	if err != nil {
		return err
	}

	lop.ForEach(filesInSameFolder, func(file fs.DirEntry, _ int) {
		if file.IsDir() {
			return
		}

		fileExtension := strings.ToLower(filepath.Ext(file.Name()))
		if !lo.Contains(subtitleFileExtensions, fileExtension) {
			return
		}

		if !strings.HasPrefix(file.Name(), fileBaseName) {
			return
		}

		tvShowFile.subtitleFiles = append(tvShowFile.subtitleFiles, file.Name())
	})

	return nil
}

// Delete will remove the TV show file from disk
func (tvShowFile *File) Delete() error {
	err := os.Remove(tvShowFile.path)
	if err != nil {
		return err
	}

	return nil
}

// DeleteWithSubtitleFiles will remove the tv show file from disk along with all subtitle files
func (tvShowFile *File) DeleteWithSubtitleFiles() error {
	for _, subtitleFile := range tvShowFile.subtitleFiles {
		err := os.Remove(fmt.Sprintf("%v/%v", tvShowFile.Dir, subtitleFile))
		if err != nil {
			return err
		}
	}

	return tvShowFile.Delete()
}

// IsMediaFile indicates if a file has a valid media file extension
func IsMediaFile(path string) bool {
	return lo.Contains(mediaFileExtensions, strings.ToLower(filepath.Ext(path)))
}

// NewTVShowFile creates a new MediaFile instance
func NewTVShowFile(path string, folderRegex string) (*File, error) {
	tvshowfile := new(File)
	tvshowfile.path = path
	err := tvshowfile.getBasicFileData()
	if err != nil {
		return nil, err
	}
	err = tvshowfile.getSubtitleFiles()
	if err != nil {
		return nil, err
	}
	err = tvshowfile.determineShow(folderRegex)
	if err != nil {
		return nil, err
	}
	err = tvshowfile.determineSeason()
	if err != nil {
		return nil, err
	}
	err = tvshowfile.determineEpisode()
	if err != nil {
		return nil, err
	}
	err = tvshowfile.processFolderSettings()
	if err != nil {
		return nil, err
	}
	return tvshowfile, nil
}
