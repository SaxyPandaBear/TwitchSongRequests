Main server code
================

This module contains the main server code, which is handling requests from the UI,
consuming events from Twitch, and sending messages to SQS to queue songs.

The server should expose endpoints to accept authorization codes for both Twitch 
and Spotify. It should write oauth details from both of them to DynamoDB. The server
should also trigger a connection to Twitch via websocket in order to consume events 
from Twitch. On event, it should filter for a `song request` type of channel point
redemption, and that should be enough to trigger sending the data to SQS, so the 
lambda can pick it up to queue it. The handler that accepts the events from Twitch
should do validation on the user input, so that we don't have it blow up downstream
in the lambda if we can catch it early. 

The idea is a **song** request, not an _album_ request. The URI follows a specific 
pattern - see [here](https://developer.spotify.com/documentation/web-api/#spotify-uris-and-ids).
We expect something that looks like `spotify:track:6rqhFgbbKwnb9MLmUQDhG6`, so we
are able to perform very simple pattern matching to check for `"spotify:track:"` as
the head of the string, and that should give us a go-ahead to queue the input. If
the actual URI does not exist, that's not a big problem. 

If there is a need to allow for requesting entities that aren't just songs, we can
look into that in the future.
