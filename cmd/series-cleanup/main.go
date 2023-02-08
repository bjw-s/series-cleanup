// Package main implements all application logic for series-cleanup
package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bjw-s/series-cleanup/internal/config"
	"github.com/bjw-s/series-cleanup/internal/helpers"
	"github.com/bjw-s/series-cleanup/internal/logger"
	"github.com/bjw-s/series-cleanup/internal/trakt"
	"github.com/bjw-s/series-cleanup/internal/tvshow"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

func main() {
	logger.SetLevel(config.AppConfig.LogLevel)

	logger.Debug("Loaded configuration",
		zap.Any("configuration", config.AppConfig),
	)

	var episodeHistory tvshow.EpisodeHistoryContainer

	// Fetch watched states from Trakt if enabled
	if config.AppConfig.Sources.Trakt.Enabled {
		if err := trakt.TraktAPI.Authenticate(
			config.AppConfig.Sources.Trakt.ClientID,
			string(config.AppConfig.Sources.Trakt.ClientSecret),
			config.AppConfig.Sources.Trakt.CacheFolder,
		); err != nil {
			logger.Fatal("Could not authenticate with Trakt",
				zap.Error(err),
			)
		}

		logger.Info("Successfully authenticated with Trakt")

		if err := episodeHistory.UpdateFromTrakt(
			config.AppConfig.Sources.Trakt.User,
			&trakt.TraktAPI,
		); err != nil {
			logger.Fatal("Could not fetch watched states from Trakt",
				zap.Error(err),
			)
		}
	}

	// Fetch watched states from Plex if enabled
	if config.AppConfig.Sources.Plex.Enabled {
	}

	// Collect all media files across the configured scanFolders
	var tvShowCollection tvshow.Collection
	lo.ForEach(config.AppConfig.ScanFolders, func(scanFolder string, _ int) {
		logger.Info("Processing...",
			zap.String("folder", scanFolder),
		)

		if !helpers.FolderExists(scanFolder) {
			logger.Fatal("Folder does not exist",
				zap.String("folder", scanFolder),
			)
		}

		err := collectTvShowFiles(&tvShowCollection, scanFolder, episodeHistory)
		if err != nil {
			logger.Fatal("Could not collect TV show files",
				zap.Error(err),
			)
		}
	})

	tvShowCollection.Process()

	logger.Info("Finished...")
}

func collectTvShowFiles(collection *tvshow.Collection, scanFolder string, watchedStateCache tvshow.EpisodeHistoryContainer) error {
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

		file, err := tvshow.NewTVShowFile(path, config.AppConfig.FolderRegex)
		if err != nil {
			return err
		}
		if file != nil {
			item, err := collection.AddMediaFile(file, config.AppConfig, watchedStateCache)
			if err != nil {
				return err
			}
			logger.Debug("Found media file",
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
