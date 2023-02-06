package tvshow

import (
	"fmt"
	"time"

	"github.com/bjw-s/series-cleanup/internal/config"
	"github.com/bjw-s/series-cleanup/internal/logger"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
	"go.uber.org/zap"
)

type episode struct {
	number       int
	files        []*File
	watchedState *watchedStateEntry
	parentSeason season
}

func (ep *episode) getFile(episodePath string) (*File, bool) {
	return lo.Find((*ep).files, func(item *File) bool {
		return item.path == episodePath
	})
}

func (ep *episode) addFile(episodeFile *File) {
	_, found := ep.getFile(episodeFile.path)
	if !found {
		ep.files = append(ep.files, episodeFile)
	}
}

func (ep *episode) isWatched() bool {
	return ep.watchedState != nil
}

func (ep *episode) watchedBefore(t time.Time) bool {
	if ep.watchedState != nil {
		episodeLastWatched := ep.watchedState.lastWatched
		return t.After(episodeLastWatched)
	}
	return false
}

func (ep *episode) process() {
	if !ep.isWatched() {
		logger.Info("Skipped",
			zap.String("show", ep.parentSeason.parentShow.ids.Name),
			zap.Int("season", ep.parentSeason.number),
			zap.Int("episode", ep.number),
			zap.String("reason", "Episode is not marked as watched"),
		)
	}

	if !ep.watchedBefore(time.Now().Add(-time.Duration(int64(config.AppConfig.DeleteAfterHours) * int64(time.Hour)))) {
		logger.Info("Skipped",
			zap.String("show", ep.parentSeason.parentShow.ids.Name),
			zap.Int("season", ep.parentSeason.number),
			zap.Int("episode", ep.number),
			zap.String("reason", fmt.Sprintf("Episode was marked as watched < %v hours ago", config.AppConfig.DeleteAfterHours)),
		)
	}

	lop.ForEach((*ep).files, func(item *File, _ int) {
		if config.AppConfig.DryRun {
			logger.Info("Skipped",
				zap.String("show", ep.parentSeason.parentShow.ids.Name),
				zap.Int("season", ep.parentSeason.number),
				zap.Int("episode", ep.number),
				zap.String("reason", "Dry-run mode enabled"),
			)
		} else {
			item.DeleteWithSubtitleFiles()
			logger.Info("Removed",
				zap.String("show", ep.parentSeason.parentShow.ids.Name),
				zap.Int("season", ep.parentSeason.number),
				zap.Int("episode", ep.number),
			)
		}
	})
}
