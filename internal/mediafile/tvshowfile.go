package mediafile

import (
	"path/filepath"
	"regexp"
	"strconv"
)

// TVShowFile represents a tv show file on disk
type TVShowFile struct {
	MediaFile

	ShowDir  string
	Season   int
	Episode  int
	Mappings ShowMapping
}

// ShowMapping contain any mappings that need to be done for a show
type ShowMapping struct {
	TraktName string
	TVDBID    int
}

func (tvShowFile *TVShowFile) determineSeason() error {
	r, err := regexp.Compile(".*[sS](\\d+)[eE]\\d+.*")
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

func (tvShowFile *TVShowFile) determineEpisode() error {
	r, err := regexp.Compile(".*[sS]\\d+[eE](\\d+).*")
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

// NewTVShowFile creates a new MediaFile instance
func NewTVShowFile(path string) (*TVShowFile, error) {
	tvshowfile := new(TVShowFile)
	tvshowfile.path = path
	tvshowfile.getBasicFileData()
	tvshowfile.getSubtitleFiles()
	tvshowfile.determineSeason()
	tvshowfile.determineEpisode()

	tvshowfile.ShowDir = filepath.Base(filepath.Dir(tvshowfile.Dir))
	return tvshowfile, nil
}
