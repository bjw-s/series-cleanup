package tvshow

import (
	"github.com/bjw-s/series-cleanup/internal/logger"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"go.uber.org/zap"
)

type season struct {
	number     int
	episodes   []*episode
	parentShow Show
}

func (seas *season) getEpisode(seasonEpisode int) (*episode, bool) {
	return lo.Find((*seas).episodes, func(ep *episode) bool {
		return ep.number == seasonEpisode
	})
}

func (seas *season) addEpisode(seasonEpisode int) *episode {
	var currentEpisode *episode
	currentEpisode, found := seas.getEpisode(seasonEpisode)
	if !found {
		seas.episodes = append(seas.episodes, &episode{number: seasonEpisode})
		currentEpisode = (seas.episodes[len(seas.episodes)-1])
	}
	return currentEpisode
}

func (seas *season) keepSeason() bool {
	if lo.Contains(seas.parentShow.settings.keepSeasons, -1) && (seas.parentShow.getLatestSeason() == seas.number) {
		return true
	}

	if lo.Contains(seas.parentShow.settings.keepSeasons, seas.number) {
		return true
	}
	return false
}

func (seas *season) process() {
	if seas.keepSeason() {
		logger.Info("Skipped",
			zap.String("show", seas.parentShow.ids.Name),
			zap.Int("season", seas.number),
			zap.String("reason", "Season is configured to be skipped"),
		)
	}

	lop.ForEach((*seas).episodes, func(ep *episode, _ int) {
		ep.parentSeason = *seas
		ep.process()
	})
}
