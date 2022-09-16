package mediafile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bjw-s/seriescleanup/internal/helpers"
)

var mediaFileExtensions = []string{".avi", ".mkv", ".mp4"}
var subtitleFileExtensions = []string{".srt"}

// MediaFile represents a media file on disk
type MediaFile struct {
	path          string
	Dir           string
	Filename      string
	Extension     string
	subtitleFiles []string
}

// NewMediaFile creates a new MediaFile instance
func NewMediaFile(path string) (*MediaFile, error) {
	mediafile := new(MediaFile)
	mediafile.path = path
	err := mediafile.getBasicFileData()
	if err != nil {
		return nil, err
	}
	err = mediafile.getSubtitleFiles()
	if err != nil {
		return nil, err
	}

	return mediafile, nil
}

func (mediafile *MediaFile) getBasicFileData() error {
	mediafile.Dir = filepath.Dir(mediafile.path)
	mediafile.Filename = filepath.Base(mediafile.path)
	mediafile.Extension = filepath.Ext(mediafile.path)
	return nil
}

func (mediafile *MediaFile) getSubtitleFiles() error {
	mediaFileBaseName := strings.TrimSuffix(mediafile.Filename, mediafile.Extension)
	filesInSameFolder, err := os.ReadDir(mediafile.Dir)
	if err != nil {
		return err
	}

	for _, file := range filesInSameFolder {
		if file.IsDir() {
			continue
		}

		_, isSubtitleFile := helpers.FindInSlice(subtitleFileExtensions, strings.ToLower(filepath.Ext(file.Name())))
		if !isSubtitleFile {
			continue
		}

		if !strings.HasPrefix(file.Name(), mediaFileBaseName) {
			continue
		}

		mediafile.subtitleFiles = append(mediafile.subtitleFiles, file.Name())
	}
	return nil
}

// Delete will remove the media file from disk
func (mediafile *MediaFile) Delete() error {
	err := os.Remove(mediafile.path)
	if err != nil {
		return err
	}

	return nil
}

// DeleteWithSubtitleFiles will remove the media file from disk along with all subtitle files
func (mediafile *MediaFile) DeleteWithSubtitleFiles() error {
	for _, subtitleFile := range mediafile.subtitleFiles {
		err := os.Remove(fmt.Sprintf("%v/%v", mediafile.Dir, subtitleFile))
		if err != nil {
			return err
		}
	}

	return mediafile.Delete()
}

// IsMediaFile indicates if a file has a valid media file extension
func IsMediaFile(path string) bool {
	_, isMediaFile := helpers.FindInSlice(mediaFileExtensions, strings.ToLower(filepath.Ext(path)))
	return isMediaFile
}
