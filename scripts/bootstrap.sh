#!/bin/bash

# Script that reads the credentials.json file stored in the same directory as
# this script, and exports them to the environment.
# Note: I know that this is cat abuse.
echo "Reading client credentials.."

spotify_client_id=`cat credentials.json | jq '.spotifyClientId'`
export SPOTIFY_CLIENT_ID="${spotify_client_id//\"}"
echo "Persisted Spotify client ID"

spotify_client_secret=`cat credentials.json | jq '.spotifyClientSecret'`
export SPOTIFY_CLIENT_SECRET="${spotify_client_secret//\"}"
echo "Persisted Spotify client secret"

twitch_client_id=`cat credentials.json | jq '.twitchClientId'`
export TWITCH_CLIENT_ID="${twitch_client_id//\"}"
echo "Persisted Twitch client ID"

twitch_client_secret=`cat credentials.json | jq '.twitchClientSecret'`
export TWITCH_CLIENT_SECRET="${twitch_client_secret//\"}"
echo "Persisted Twitch client secret"
