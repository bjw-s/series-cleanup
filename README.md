# series-cleanup
![GitHub last commit](https://img.shields.io/github/last-commit/bjws/series-cleanup)    ![GitHub Workflow Status](https://img.shields.io/github/workflow/status/bjws/series-cleanup/docker)    [![Docker](https://img.shields.io/docker/pulls/bjws/series-cleanup)](https://hub.docker.com/r/bjws/series-cleanup)

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

Create a copy of [settings.sample.json](settings.sample.json) and modify the settings to your preferences. Make sure to specify the Trakt Client ID and Client Secret according to your [Trakt API app](https://trakt.tv/oauth/applications) values.

## Docker

A Docker image can be found at [Docker Hub](https://hub.docker.com/r/bjws/series-cleanup). This image expects the configuration file to be available at `/data/settings.json`.