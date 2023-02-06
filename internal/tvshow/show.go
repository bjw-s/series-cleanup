package tvshow

import (
	"github.com/bjw-s/series-cleanup/internal/logger"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"go.uber.org/zap"
)

// Identifiers structures the ways a show can be identified
type Identifiers struct {
	Name  string
	Trakt int
	Slug  string
	IMDB  string
	TVDB  int
	TMDB  int
}

type showSettings struct {
	keep        bool
	keepSeasons []int
}

// Show represents a TV show and it's underlying data
type Show struct {
	ids      Identifiers
	seasons  []*season
	settings showSettings
}

// getSeason returns the season with the specified number and a boolean indicating
// if the season was found
func (shw *Show) getSeason(number int) (*season, bool) {
	return lo.Find((*shw).seasons, func(seas *season) bool {
		return seas.number == number
	})
}

// getLatestSeason returns the number of the latest season within the list of
// seasons
func (shw *Show) getLatestSeason() int {
	var seasons []int
	lop.ForEach((*shw).seasons, func(seas *season, _ int) {
		seasons = append(seasons, seas.number)
	})
	return lo.Max(seasons)
}

// addSeason adds a new season to the list of seasons
func (shw *Show) addSeason(number int) *season {
	var currentSeason *season
	currentSeason, found := shw.getSeason(number)
	if !found {
		shw.seasons = append(shw.seasons, &season{number: number})
		currentSeason = (shw.seasons[len(shw.seasons)-1])
	}
	return currentSeason
}

// Process will run any required processing for the TV show
func (shw *Show) process() {
	if shw.settings.keep {
		logger.Info("Skipped",
			zap.String("show", shw.ids.Name),
			zap.String("reason", "Show is configured to be skipped"),
		)
		return
	}

	lop.ForEach((*shw).seasons, func(seas *season, _ int) {
		seas.parentShow = *shw
		seas.process()
	})
}
