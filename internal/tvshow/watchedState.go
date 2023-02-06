package tvshow

import (
	"strings"
	"time"

	"github.com/bjw-s/series-cleanup/internal/trakt"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

// WatchedStateContainer is a container to keep multiple watchedStates
type WatchedStateContainer []*WatchedState

// WatchedState keeps track of the watch history for a specific show
type WatchedState struct {
	show    Identifiers
	history []*watchedStateEntry
}

type watchedStateEntry struct {
	season      int
	episode     int
	lastWatched time.Time
	source      string
}

func (watchedStateContainer *WatchedStateContainer) addWatchState(showName string) *WatchedState {
	var currentWatchState *WatchedState

	currentWatchState, found := (*watchedStateContainer).FindShowByName(showName)
	if !found {
		*watchedStateContainer = append(
			*watchedStateContainer,
			&WatchedState{
				show: Identifiers{
					Name: showName,
				},
			},
		)
		currentWatchState = (*watchedStateContainer)[len(*watchedStateContainer)-1]
	}
	return currentWatchState
}

// UpdateFromTrakt will add Trakt watch history to a WatchedStateCache
func (watchedStateContainer *WatchedStateContainer) UpdateFromTrakt(username string, api *trakt.API) error {
	result, err := api.GetWatchedShows(username)
	if err != nil {
		return err
	}

	lop.ForEach(result, func(watchedShow *trakt.WatchedShow, _ int) {
		show := watchedStateContainer.addWatchState(watchedShow.Show.Title)
		show.show.Slug = watchedShow.Show.IDS.Slug
		show.show.Trakt = watchedShow.Show.IDS.Trakt
		show.show.IMDB = watchedShow.Show.IDS.IMDB
		show.show.TVDB = watchedShow.Show.IDS.TVDB
		show.show.TMDB = watchedShow.Show.IDS.TMDB

		lop.ForEach(watchedShow.Seasons, func(traktWatchedShowSeason trakt.Season, _ int) {
			lop.ForEach(traktWatchedShowSeason.Episodes, func(traktWatchedShowEpisode trakt.Episode, _ int) {
				(*show).history = append(
					(*show).history,
					&watchedStateEntry{
						season:      traktWatchedShowSeason.Number,
						episode:     traktWatchedShowEpisode.Number,
						lastWatched: traktWatchedShowEpisode.LastWatched,
						source:      "trakt",
					},
				)
			})
		})
	})
	return nil
}

// FindShowByName returns a watched show by name
func (watchedStateContainer *WatchedStateContainer) FindShowByName(name string) (*WatchedState, bool) {
	return lo.Find(*watchedStateContainer, func(item *WatchedState) bool {
		return strings.EqualFold(item.show.Name, name)
	})
}

// FindShowBySlug returns a watched show by Slug
func (watchedStateContainer *WatchedStateContainer) FindShowBySlug(slug string) (*WatchedState, bool) {
	return lo.Find(*watchedStateContainer, func(item *WatchedState) bool {
		return item.show.Slug == slug
	})
}

// FindShowByTraktID returns a watched show by Trakt ID
func (watchedStateContainer *WatchedStateContainer) FindShowByTraktID(traktid int) (*WatchedState, bool) {
	return lo.Find(*watchedStateContainer, func(item *WatchedState) bool {
		return item.show.Trakt == traktid
	})
}

// FindShowByTVDBID returns a watched show by TVDB id
func (watchedStateContainer *WatchedStateContainer) FindShowByTVDBID(tvdbid int) (*WatchedState, bool) {
	return lo.Find(*watchedStateContainer, func(item *WatchedState) bool {
		return item.show.TVDB == tvdbid
	})
}

// FindShowByIMDBID returns a watched show by IMDb id
func (watchedStateContainer *WatchedStateContainer) FindShowByIMDBID(imdbid string) (*WatchedState, bool) {
	return lo.Find(*watchedStateContainer, func(item *WatchedState) bool {
		return strings.EqualFold(item.show.IMDB, imdbid)
	})
}

// getEpisodeWatchHistory returns the watch history for a given season and episode number
func (watchedState *WatchedState) getEpisodeWatchHistory(season int, episode int) (*watchedStateEntry, bool) {
	result, found := lo.Find((*watchedState).history, func(item *watchedStateEntry) bool {
		return (item.season == season && item.episode == episode)
	})
	return result, found
}
