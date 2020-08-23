# TwitchSongRequests
This is yet another project that I will probably never finish. A Twitch 
integration to enqueue song requests in Spotify

## Table of Contents
1. [What does this project do?](#what-does-this-project-do)
    1. [How it works](#how-it-works)
    1. [Required Twitch accesses](#required-twitch-accesses)
    1. [Required Spotify accesses](#required-spotify-accesses)
1. [Running it locally](#running-it-locally)
1. [Testing](#testing)
1. [Secret management](#secret-management)
1. [Deployments](#deployments)

## What does this project do?
The purpose of this project is to allow Twitch streamers to integrate their 
channel point redemptions with Spotify song requests.

#### How it works

TBD for the Spotify portion

#### Required Twitch accesses

#### Required Spotify accesses

TBD

## Running it locally

TBD

## Testing

TBD

## Secret management
For now, secret management is done simply in the `src/credentials.json` file. 
The template file helps to show what to expect in the credentials file:

```javascript
{
    "twitch_client_id": "TWITCH_ID",
    "twitch_client_secret": "TWITCH_SECRET",
    "spotify_client_id": "SPOTIFY_ID",
    "spotify_client_secret": "SPOTIFY_SECRET"
}
```

Replacing the values to the existing keys should suffice. 

## Deployments

TBD
