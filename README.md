# series-cleanup
![GitHub last commit](https://img.shields.io/github/last-commit/bjw-s/series-cleanup?style=flat-square)    ![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/bjw-s/series-cleanup/release.yaml?style=flat-square)    [![Go Report Card](https://goreportcard.com/badge/github.com/bjw-s/series-cleanup?style=flat-square)](https://goreportcard.com/report/github.com/bjw-s/series-cleanup)    [![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/bjw-s/series-cleanup?sort=semver&style=flat-square)](https://github.com/bjw-s/series-cleanup/releases/latest)

This app searches the specified folders for TV shows, and removes them if they are marked as watched on Trakt.tv.

## Usage

This app requires a [Trakt API app](https://trakt.tv/oauth/applications) to be created.

## Configuration

Create a copy of [examples/settings.json](examples/settings.json) and modify the settings to your preferences. Make sure to specify the Trakt Client ID and Client Secret according to your [Trakt API app](https://trakt.tv/oauth/applications) values.

## Docker

A Docker image can be found here: [GitHub Container Registry](https://ghcr.io/bjw-s/series-cleanup). This image expects the configuration file to be available at `/config/settings.json`.
