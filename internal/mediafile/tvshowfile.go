package mediafile

import (
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/oriser/regroup"
)

// TVShowFile represents a tv show file on disk
type TVShowFile struct {
	MediaFile

	Show     string
	Season   int
	Episode  int
	Mappings ShowMapping
}

// ShowMapping contain any mappings that need to be done for a show
type ShowMapping struct {
	TraktName string
	TVDBID    int
	IMDBID    string
}

func (tvShowFile *TVShowFile) determineShow(regex string) error {
	show := filepath.Base(filepath.Dir(tvShowFile.Dir))
	if len(regex) > 0 {
		folderRegex := regroup.MustCompile(regex)
		matches, err := folderRegex.Groups(show)
		if err != nil {
			return err
		}

		if matchShow, ok := matches["Show"]; ok {
			show = matchShow
		}

		if matchIMDBID, ok := matches["IMDBID"]; ok {
			tvShowFile.Mappings.IMDBID = matchIMDBID
		}

		if matchTVBID, ok := matches["TVDBID"]; ok {
			tvdbID, err := strconv.Atoi(matchTVBID)
			if err != nil {
				return err
			}
			tvShowFile.Mappings.TVDBID = tvdbID
		}
	}
	tvShowFile.Show = show
	return nil
}

func (tvShowFile *TVShowFile) determineSeason() error {
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

func (tvShowFile *TVShowFile) determineEpisode() error {
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

// NewTVShowFile creates a new MediaFile instance
func NewTVShowFile(path string, folderRegex string) (*TVShowFile, error) {
	tvshowfile := new(TVShowFile)
	tvshowfile.path = path
	tvshowfile.getBasicFileData()
	tvshowfile.getSubtitleFiles()
	tvshowfile.determineShow(folderRegex)
	tvshowfile.determineSeason()
	tvshowfile.determineEpisode()
	return tvshowfile, nil
}
