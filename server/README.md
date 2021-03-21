Twitch Event Interceptor
========================

> [wiki](https://github.com/SaxyPandaBear/TwitchSongRequests/wiki/Architecture-Deep-Dive#twitch-event-interceptor)

This module is what drives a good majority of the flow for the overall service.
It accepts SQS messages that either initiate a connection to the Twitch PubSub
API, or disconnect from a specific channel topic. When an event occurs, the 
client code for the interceptor parses the event JSON, and uses the user input
in the event to submit a message to an SQS topic that queues the song into the
user's Spotify player.

The idea is a **song** request, not an _album_ request. The URI follows a specific
[pattern](https://developer.spotify.com/documentation/web-api/#spotify-uris-and-ids).
We expect something that looks like `spotify:track:6rqhFgbbKwnb9MLmUQDhG6`, so we
are able to perform very simple pattern matching to check for `"spotify:track:"` as
the head of the string, and that should give us a go-ahead to queue the input. If
the actual URI does not exist, that's not a big problem.

If there is a need to allow for requesting entities that aren't just songs, we can
look into that in the future.

### Build
`gradle assemble`

### Test
`gradle test`

### Run locally
`gradle :server:run --args path/to/local/config/file`
