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
	number         int
	files          []*File
	episodeHistory *episodeHistoryEntry
	parentSeason   season
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
	return ep.episodeHistory != nil
}

func (ep *episode) watchedBefore(t time.Time) bool {
	if ep.episodeHistory != nil {
		episodeLastWatched := ep.episodeHistory.lastWatched
		return t.After(episodeLastWatched)
	}
	return false
}

func (ep *episode) delete() {
	lop.ForEach((*ep).files, func(item *File, _ int) {
		if config.AppConfig.DryRun {
			logger.Info("Skipped",
				zap.String("show", ep.parentSeason.parentShow.ids.Name),
				zap.Int("season", ep.parentSeason.number),
				zap.Int("episode", ep.number),
				zap.String("reason", "Dry-run mode enabled"),
			)
		} else {
			// item.DeleteWithSubtitleFiles()
			logger.Info("Removed",
				zap.String("show", ep.parentSeason.parentShow.ids.Name),
				zap.Int("season", ep.parentSeason.number),
				zap.Int("episode", ep.number),
			)
		}
	})
}

func (ep *episode) deleteIfWatchedMoreThanHoursAgo(hours int) {
	if !ep.isWatched() || !ep.watchedBefore(time.Now().UTC().Add(-time.Duration(int64(hours)*int64(time.Hour)))) {
		logger.Info("Skipped",
			zap.String("show", ep.parentSeason.parentShow.ids.Name),
			zap.Int("season", ep.parentSeason.number),
			zap.Int("episode", ep.number),
			zap.String("reason", fmt.Sprintf("Episode was marked as watched < %v hours ago", hours)),
		)
		return
	}

	ep.delete()
}

func (ep *episode) process() {
	if !ep.isWatched() {
		logger.Info("Skipped",
			zap.String("show", ep.parentSeason.parentShow.ids.Name),
			zap.Int("season", ep.parentSeason.number),
			zap.Int("episode", ep.number),
			zap.String("reason", "Episode is not marked as watched"),
		)
		return
	}

	deleteAfterHours := config.AppConfig.Rules.DeleteWatchedAfterHours
	if ep.parentSeason.parentShow.rules.DeleteWatchedAfterHours != nil {
		deleteAfterHours = ep.parentSeason.parentShow.rules.DeleteWatchedAfterHours
	}
	if deleteAfterHours != nil {
		ep.deleteIfWatchedMoreThanHoursAgo(*deleteAfterHours)
	}
}
