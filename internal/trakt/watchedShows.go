package trakt

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// Episode represents an episode that was reported as watched on Trakt
type Episode struct {
	Number      int       `json:"number"`
	LastWatched time.Time `json:"last_watched_at"`
}

// WatchedShow represents a show that reported watched on Trakt
type WatchedShow struct {
	Show    show     `json:"show"`
	Seasons []Season `json:"seasons"`
}

// GetWatchedShows returns the watched shows
func (api *API) GetWatchedShows(username string) ([]*WatchedShow, error) {
	var watchedShows []*WatchedShow
	err := api.validate()
	if err != nil {
		return nil, err
	}

	result, err := api.sendRequest(http.MethodGet, fmt.Sprintf("/users/%v/watched/shows", username), nil)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(result.Body, &watchedShows)
	if err != nil {
		return nil, err
	}

	return watchedShows, nil
}
