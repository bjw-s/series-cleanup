package trakt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type show struct {
	Title string `json:"title"`
	IDS   struct {
		Trakt int
		Slug  string
		TVDB  int
		IMDB  string
		TMDB  int
	} `json:"ids"`
}

// Season represents a season that was reported as (partially) watched on Trakt
type Season struct {
	Number   int       `json:"number"`
	Episodes []Episode `json:"episodes"`
}

// FindEpisode tries to find if a (partially) watched season on Trakt contains an episode with a specified number
func (s *Season) FindEpisode(episodeNumber int) *Episode {
	if len(s.Episodes) == 0 {
		return nil
	}

	for _, episode := range s.Episodes {
		if episode.Number == episodeNumber {
			return &episode
		}
	}

	return nil
}

// Episode represents an episode that was reported as watched on Trakt
type Episode struct {
	Number      int       `json:"number"`
	LastWatched time.Time `json:"last_watched_at"`
}

// LastWatchedBefore returns when an episode was reported as last watched on Trakt
func (e *Episode) LastWatchedBefore(t time.Time) bool {
	return e.LastWatched.Sub(t) < 0
}

// WatchedShow represents a show that reported watched on Trakt
type WatchedShow struct {
	Show    show     `json:"show"`
	Seasons []Season `json:"seasons"`
}

// FindSeason tries to find if a watched show on Trakt contains a season with a specified number
func (w *WatchedShow) FindSeason(seasonNumber int) *Season {
	if len(w.Seasons) == 0 {
		return nil
	}

	for _, season := range w.Seasons {
		if season.Number == seasonNumber {
			return &season
		}
	}

	return nil
}

// User represents the Trakt User
type User struct {
	Name         string
	WatchedShows []WatchedShow
}

// GetWatchedShows returns the watched shows for this user
func (user *User) GetWatchedShows(api API) error {
	err := api.validate()
	if err != nil {
		return err
	}

	result, err := api.sendRequest(http.MethodGet, fmt.Sprintf("/users/%v/watched/shows", user.Name), nil)
	if err != nil {
		return err
	}

	err = json.Unmarshal(result.Body, &user.WatchedShows)
	if err != nil {
		return err
	}

	return nil
}

// FindWatchedShowByName returns a watched show for this user by name
func (user *User) FindWatchedShowByName(name string) *WatchedShow {
	if user.WatchedShows == nil {
		return nil
	}

	for _, show := range user.WatchedShows {
		if strings.ToLower(show.Show.Title) == strings.ToLower(name) {
			return &show
		}
	}

	return nil
}

// FindWatchedShowByTVDBID returns a watched show for this user by TVDB id
func (user *User) FindWatchedShowByTVDBID(tvdbid int) *WatchedShow {
	if user.WatchedShows == nil {
		return nil
	}

	for _, show := range user.WatchedShows {
		if show.Show.IDS.TVDB == tvdbid {
			return &show
		}
	}

	return nil
}
