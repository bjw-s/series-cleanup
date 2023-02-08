// Package tvshow contains everything required to describe / structure tv shows
package tvshow

import (
	"github.com/bjw-s/series-cleanup/internal/config"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

// Collection is a container to keep multiple tvShows
type Collection []*Show

func (collection *Collection) getShow(showName string) (*Show, bool) {
	return lo.Find(*collection, func(item *Show) bool {
		return item.ids.Name == showName
	})
}

func (collection *Collection) addShow(showName string) *Show {
	var currentShow *Show

	currentShow, found := (*collection).getShow(showName)
	if !found {
		*collection = append(
			*collection,
			&Show{
				ids: Identifiers{
					Name: showName,
				},
			},
		)
		currentShow = (*collection)[len(*collection)-1]
	}
	return currentShow
}

// AddMediaFile adds a given MediaFile to a TVShowCollection
func (collection *Collection) AddMediaFile(file *File, config config.Config, watchedStateCache EpisodeHistoryContainer) (*File, error) {
	show := collection.addShow(file.Identifiers.Name)
	show.ids = file.Identifiers
	show.rules = file.Rules

	var showHistory *EpisodeHistory
	showHistoryFound := false
	if show.ids.Trakt != 0 {
		showHistory, showHistoryFound = watchedStateCache.FindShowByTraktID(show.ids.Trakt)
	} else if show.ids.IMDB != "" {
		showHistory, showHistoryFound = watchedStateCache.FindShowByIMDBID(show.ids.IMDB)
	} else if show.ids.TVDB != 0 {
		showHistory, showHistoryFound = watchedStateCache.FindShowByTVDBID(show.ids.TVDB)
	} else if show.ids.Slug != "" {
		showHistory, showHistoryFound = watchedStateCache.FindShowBySlug(show.ids.Slug)
	} else {
		showHistory, showHistoryFound = watchedStateCache.FindShowByName(show.ids.Name)
	}
	season := show.addSeason(file.Season)

	episode := season.addEpisode(file.Episode)
	episode.addFile(file)
	if showHistoryFound {
		episode.episodeHistory, _ = showHistory.getEpisodeWatchHistory(season.number, episode.number)
	}

	return file, nil
}

// Process will run any required processing for each collected TV show
func (collection *Collection) Process() {
	lop.ForEach((*collection), func(item *Show, _ int) {
		item.process()
	})
}
