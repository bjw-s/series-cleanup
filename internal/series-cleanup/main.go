// Package series_cleanup implements all series-cleanup functionality
package series_cleanup

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bjw-s/series-cleanup/internal/config"
	"github.com/bjw-s/series-cleanup/internal/helpers"
	"github.com/bjw-s/series-cleanup/internal/trakt"
	"github.com/bjw-s/series-cleanup/internal/tvshow"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// AppConfig contains the Config instance used by KoboMail
var AppConfig *config.Config

// Run executes the main series-cleanup logic
func Run() {
	logger := zap.S()
	var episodeHistory tvshow.EpisodeHistoryContainer

	// Fetch watched states from Trakt if enabled
	if AppConfig.Trakt.Enabled {
		if err := trakt.TraktAPI.Authenticate(
			AppConfig.Trakt.ClientID,
			string(AppConfig.Trakt.ClientSecret),
			AppConfig.Trakt.CacheDir,
		); err != nil {
			logger.Fatal("Could not authenticate with Trakt",
				zap.Error(err),
			)
		}

		logger.Infow("Successfully authenticated with Trakt")

		if err := episodeHistory.UpdateFromTrakt(
			AppConfig.Trakt.User,
			&trakt.TraktAPI,
		); err != nil {
			logger.Fatal("Could not fetch watched states from Trakt",
				zap.Error(err),
			)
		}
	}

	// TODO: implement this after Plex integration
	// Fetch watched states from Plex if enabled
	// if AppConfig.Plex.Enabled {}

	// Collect all media files across the configured scanFolders
	var tvShowCollection tvshow.Collection
	lo.ForEach(AppConfig.ScanFolders, func(scanFolder string, _ int) {
		logger.Infow("Processing...",
			zap.String("folder", scanFolder),
		)

		if !helpers.FolderExists(scanFolder) {
			logger.Fatalw("Folder does not exist",
				zap.String("folder", scanFolder),
			)
		}

		err := collectTvShowFiles(&tvShowCollection, scanFolder, episodeHistory)
		if err != nil {
			logger.Fatalw("Could not collect TV show files",
				zap.Error(err),
			)
		}
	})

	tvShowCollection.Process(AppConfig)

	logger.Infow("Finished...")
}

func collectTvShowFiles(collection *tvshow.Collection, scanFolder string, watchedStateCache tvshow.EpisodeHistoryContainer) error {
	logger := zap.S()
	err := filepath.WalkDir(scanFolder, func(path string, info os.DirEntry, nestedErr error) error {
		if info.IsDir() {
			return nil
		}

		fileName := filepath.Base(path)
		if strings.HasPrefix(fileName, ".") {
			return nil
		}

		if !tvshow.IsMediaFile(path) {
			return nil
		}

		file, err := tvshow.NewTVShowFile(path, AppConfig)
		if err != nil {
			return err
		}
		if file != nil {
			item, err := collection.AddMediaFile(file, AppConfig, watchedStateCache)
			if err != nil {
				return err
			}
			logger.Debugw("Found media file",
				zap.String("folder", item.Dir),
				zap.String("file", item.Filename),
			)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
