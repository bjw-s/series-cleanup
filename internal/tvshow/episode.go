package tvshow

import (
	"fmt"
	"time"

	"github.com/bjw-s/series-cleanup/internal/config"
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

func (ep *episode) delete(dryrun bool) {
	logger := zap.S()
	lop.ForEach((*ep).files, func(item *File, _ int) {
		if dryrun {
			logger.Infow("Skipped",
				zap.String("show", ep.parentSeason.parentShow.ids.Name),
				zap.Int("season", ep.parentSeason.number),
				zap.Int("episode", ep.number),
				zap.String("reason", "Dry-run mode enabled"),
			)
		} else {
			// TODO: Re-enable after testing
			item.DeleteWithSubtitleFiles()
			logger.Infow("Removed",
				zap.String("show", ep.parentSeason.parentShow.ids.Name),
				zap.Int("season", ep.parentSeason.number),
				zap.Int("episode", ep.number),
			)
		}
	})
}

func (ep *episode) deleteIfWatchedMoreThanHoursAgo(hours int, conf *config.Config) {
	logger := zap.S()
	if !ep.isWatched() || !ep.watchedBefore(time.Now().UTC().Add(-time.Duration(int64(hours)*int64(time.Hour)))) {
		logger.Infow("Skipped",
			zap.String("show", ep.parentSeason.parentShow.ids.Name),
			zap.Int("season", ep.parentSeason.number),
			zap.Int("episode", ep.number),
			zap.String("reason", fmt.Sprintf("Episode was marked as watched < %v hours ago", hours)),
		)
		return
	}

	ep.delete(conf.DryRun)
}

func (ep *episode) process(conf *config.Config) {
	logger := zap.S()
	if !ep.isWatched() {
		logger.Infow("Skipped",
			zap.String("show", ep.parentSeason.parentShow.ids.Name),
			zap.Int("season", ep.parentSeason.number),
			zap.Int("episode", ep.number),
			zap.String("reason", "Episode is not marked as watched"),
		)
		return
	}

	deleteAfterHours := conf.Rules.DeleteWatchedAfterHours
	if ep.parentSeason.parentShow.rules.DeleteWatchedAfterHours != 0 {
		deleteAfterHours = ep.parentSeason.parentShow.rules.DeleteWatchedAfterHours
	}
	if deleteAfterHours != 0 {
		ep.deleteIfWatchedMoreThanHoursAgo(deleteAfterHours, conf)
	}
}
