package tvshow

import (
	"strings"
	"time"

	"github.com/bjw-s/series-cleanup/internal/trakt"
	"github.com/samber/lo"
)

// EpisodeHistoryContainer is a container to keep multiple EpisodeHistory entities
type EpisodeHistoryContainer []*EpisodeHistory

// EpisodeHistory keeps track of episode history for a specific show
type EpisodeHistory struct {
	show    Identifiers
	history []*episodeHistoryEntry
}

type episodeHistoryEntry struct {
	season  int
	episode int
	// TODO: Add after Plex integration
	// dateAdded   time.Time
	lastWatched time.Time
	source      string
}

func (episodeHistoryContainer *EpisodeHistoryContainer) addEpisodeHistory(showName string) *EpisodeHistory {
	var currentWatchState *EpisodeHistory

	currentWatchState, found := (*episodeHistoryContainer).FindShowByName(showName)
	if !found {
		*episodeHistoryContainer = append(
			*episodeHistoryContainer,
			&EpisodeHistory{
				show: Identifiers{
					Name: showName,
				},
			},
		)
		currentWatchState = (*episodeHistoryContainer)[len(*episodeHistoryContainer)-1]
	}
	return currentWatchState
}

// UpdateFromTrakt will add Trakt watch history to a WatchedStateCache
func (episodeHistoryContainer *EpisodeHistoryContainer) UpdateFromTrakt(username string, api *trakt.API) error {
	result, err := api.GetWatchedShows(username)
	if err != nil {
		return err
	}

	lo.ForEach(result, func(watchedShow *trakt.WatchedShow, _ int) {
		show := episodeHistoryContainer.addEpisodeHistory(watchedShow.Show.Title)
		show.show.Slug = watchedShow.Show.IDS.Slug
		show.show.Trakt = watchedShow.Show.IDS.Trakt
		show.show.IMDB = watchedShow.Show.IDS.IMDB
		show.show.TVDB = watchedShow.Show.IDS.TVDB
		show.show.TMDB = watchedShow.Show.IDS.TMDB

		lo.ForEach(watchedShow.Seasons, func(traktWatchedShowSeason trakt.Season, _ int) {
			lo.ForEach(traktWatchedShowSeason.Episodes, func(traktWatchedShowEpisode trakt.Episode, _ int) {
				(*show).history = append(
					(*show).history,
					&episodeHistoryEntry{
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
func (episodeHistoryContainer *EpisodeHistoryContainer) FindShowByName(name string) (*EpisodeHistory, bool) {
	return lo.Find(*episodeHistoryContainer, func(item *EpisodeHistory) bool {
		return strings.EqualFold(item.show.Name, name)
	})
}

// FindShowBySlug returns a watched show by Slug
func (episodeHistoryContainer *EpisodeHistoryContainer) FindShowBySlug(slug string) (*EpisodeHistory, bool) {
	return lo.Find(*episodeHistoryContainer, func(item *EpisodeHistory) bool {
		return item.show.Slug == slug
	})
}

// FindShowByTraktID returns a watched show by Trakt ID
func (episodeHistoryContainer *EpisodeHistoryContainer) FindShowByTraktID(traktid int) (*EpisodeHistory, bool) {
	return lo.Find(*episodeHistoryContainer, func(item *EpisodeHistory) bool {
		return item.show.Trakt == traktid
	})
}

// FindShowByTVDBID returns a watched show by TVDB id
func (episodeHistoryContainer *EpisodeHistoryContainer) FindShowByTVDBID(tvdbid int) (*EpisodeHistory, bool) {
	return lo.Find(*episodeHistoryContainer, func(item *EpisodeHistory) bool {
		return item.show.TVDB == tvdbid
	})
}

// FindShowByIMDBID returns a watched show by IMDb id
func (episodeHistoryContainer *EpisodeHistoryContainer) FindShowByIMDBID(imdbid string) (*EpisodeHistory, bool) {
	return lo.Find(*episodeHistoryContainer, func(item *EpisodeHistory) bool {
		return strings.EqualFold(item.show.IMDB, imdbid)
	})
}

// getEpisodeWatchHistory returns the watch history for a given season and episode number
func (watchedState *EpisodeHistory) getEpisodeWatchHistory(season int, episode int) (*episodeHistoryEntry, bool) {
	result, found := lo.Find((*watchedState).history, func(item *episodeHistoryEntry) bool {
		return (item.season == season && item.episode == episode)
	})
	return result, found
}
