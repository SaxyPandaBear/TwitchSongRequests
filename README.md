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
By reading the specific channel topic for when channel points are redeemed, this
app can intercept those events, and given a specific type of request, we can 
act on a specific channel point redemption to invoke queueing up a song in Spotify.

TBD for the Spotify portion

### Required Twitch accesses
For the Twitch integration, this project requires privileged access to the
`channel_read` and `channel:read:redemptions` scopes. The first one is so that
the app can fetch the channel ID, which is a required step in identifying which
individual channel topic to listen to. The second one is to grant the app access
to read channel point redemption events from their stream.

### Required Spotify accesses

TBD

## Running it locally

Clone the repo, first.
```bash
git clone https://github.com/SaxyPandaBear/TwitchSongRequests.git
```

For now, everything lives in the `./src/demo.js` file. This requires a `./src/credentials.json` 
file to run. That file is gitignored, so its safe to create it and drop it next 
to the main app file. 

Create a `credentials.json` file and place it in the same directory as the main
`./src/demo.js` file. Use the credentials file template as a guideline.

Place all of the required sensitive data in that file so that the app can read 
from it and reference that data. 

Note: This includes adding a new key/value pair in the JSON file: `"twitch_auth_code": "foobarbaz"` 
For now, this is how it runs locally since it can't source an auth code from an external source.

### Running the app

TBD

### Demo applications to proof-of-concept the tech

```bash
cd demo
# populate the authorization codes in your local credentials.json file
node twitch.js
node spotify.js
```

An example of what log output looks like:
```
INFO: Socket Opened
SENT: {"type":"PING"}
{ message: { type: 'PONG' } }
{ message: { type: 'RESPONSE', error: '', nonce: 'abc123' } }
{ message: 
   { type: 'MESSAGE',
     data:
      { topic: 'channel-points-channel-v1.106060203',
        message: '{"type":"reward-redeemed","data":{"timestamp":"2020-08-23T20:21:56.588735036Z","redemption":{"id":"897dd20c-ec7f-42da-9e0a-610091785a4d","user":{"id":"106060203","login":"saxypandabear","display_name":"SaxyPandaBear"},"channel_id":"106060203","redeemed_at":"2020-08-23T20:21:56.588735036Z","reward":{"id":"ca20aaa2-5fa8-4b29-a9a6-34275ee911f4","channel_id":"106060203","title":"Song Request","prompt":"Only applies for music streams. Request a song you want 
me to attempt to learn by ear.","cost":10000,"is_user_input_required":true,"is_sub_only":false,"image":null,"default_image":{"url_1x":"https://static-cdn.jtvnw.net/custom-reward-images/default-1.png","url_2x":"https://static-cdn.jtvnw.net/custom-reward-images/default-2.png","url_4x":"https://static-cdn.jtvnw.net/custom-reward-images/default-4.png"},"background_color":"#FA2929","is_enabled":true,"is_paused":false,"is_in_stock":true,"max_per_stream":{"is_enabled":false,"max_per_stream":0},"should_redemptions_skip_request_queue":false,"template_id":null,"updated_for_indicator_at":"2020-01-01T15:11:26.647212555Z","max_per_user_per_stream":{"is_enabled":false,"max_per_user_per_stream":0},"global_cooldown":{"is_enabled":false,"global_cooldown_seconds":0},"redemptions_redeemed_current_stream":0,"cooldown_expires_at":null},"user_input":"hello","status":"UNFULFILLED"}}}' } } }
```

Note that subsequent runs of the application require a new authorization code 
for each run. 

### OAuth authentication flow
There is a [v2.1 Postman collection](./TwitchSongRequestsReference.postman_collection.json) exported to the root of the project directory
that documents the required authorization calls.

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

Replacing the values to the existing keys should suffice, since the main app is
reading in the raw JSON file, and referencing keys in the JSON object directly. 

## Deployments

TBD
