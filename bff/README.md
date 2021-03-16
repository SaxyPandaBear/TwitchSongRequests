Backend for frontend
====================

> [wiki](https://github.com/SaxyPandaBear/TwitchSongRequests/wiki/Architecture-Deep-Dive#web-application)

This module is the backend for the frontend UI, which orchestrates initiation for 
OAuth access tokens from the user, persisting them to Secrets Manager, and submitting
messages to the SQS queue that the Twitch event interceptor listens on to know that a
new user wants to start using the service.

The server should expose endpoints to accept authorization codes for both Twitch 
and Spotify. It should write OAuth details from both of those services to Secrets Manager.
The server should also submit a message to a SQS topic so that the event interceptor
microservice can create a new, persistent WebSocket connection to the user's specific channel topic.

### Build
`gradle assemble`

> Can also run the underlying NPM command, `npm install`

### Test
TBD

### Run locally
`node app.js`
