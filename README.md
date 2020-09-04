# series-cleanup
![GitHub last commit](https://img.shields.io/github/last-commit/bjw-s/series-cleanup?style=flat-square)    ![GitHub Workflow Status](https://img.shields.io/github/workflow/status/bjw-s/series-cleanup/Docker%20Image%20CI?style=flat-square)    [![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/bjw-s/series-cleanup?sort=semver&style=flat-square)](https://github.com/bjw-s/series-cleanup/releases/latest)

This app searches the specified folders for TV shows, and removes them if they are marked as watched on Trakt.tv.

## Usage

This app requires a [Trakt API app](https://trakt.tv/oauth/applications) to be created. 

Command line options can be specified when running the app:

```
Usage of ./series-cleanup:
  -c string
    	Specify config file to use (default "settings.json")
  -d string
    	Specify folder to store data (default "/data")
```

## Configuration

Create a copy of [examples/settings.json](examples/settings.json) and modify the settings to your preferences. Make sure to specify the Trakt Client ID and Client Secret according to your [Trakt API app](https://trakt.tv/oauth/applications) values.

The Trakt Client ID and Client Secret can also be provided as environment variables instead of putting them in the configuration file:

* `TRAKT_CLIENT_ID`
* `TRAKT_CLIENT_SECRET`

## Docker

A Docker image can be found here: [GitHub Container Registry](https://github.com/users/bjw-s/packages/container/series-cleanup/). This image expects the configuration file to be available at `/data/settings.json`.