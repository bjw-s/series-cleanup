// Package helpers implements several useful functions
package helpers

import (
	"os"
)

// SliceContainsString takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func SliceContainsString(slice []string, val string) bool {
	s := make([]interface{}, len(slice))
	for i, v := range slice {
		s[i] = v
	}
	return SliceContains(s, val)
}

// SliceContainsInt takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func SliceContainsInt(slice []int, val int) bool {
	s := make([]interface{}, len(slice))
	for i, v := range slice {
		s[i] = v
	}
	return SliceContains(s, val)
}

// SliceContains takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func SliceContains(slice []interface{}, val interface{}) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// FileExists takes a string returns if it is an existing file
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FolderExists takes a string returns if it is an existing folder
func FolderExists(foldername string) bool {
	info, err := os.Stat(foldername)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
