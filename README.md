# TwitchSongRequests

Integrate Twitch channel point rewards directly with a user's Spotify player.
Configure a channel point reward to accept a Spotify song URL, and enqueue
the song in the user's current playing session.

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/SaxyPandaBear/TwitchSongRequests/ci.yml?style=for-the-badge) ![GitHub commit activity](https://img.shields.io/github/commit-activity/m/SaxyPandaBear/TwitchSongRequests?style=for-the-badge)

---

# Inspiration
This project was inspired by [Kaije](https://www.twitch.tv/kaije), because their 
Twitch channel has a channel point reward for requesting a song on Spotify. On that 
fateful day, I requested multiple songs in quick succession, and they told me to hold 
off on requesting songs because they have to stop working on art in order to manually 
copy the Spotify URIs I was sending, in order to queue the songs in their Spotify 
player. Between jobs, I coded up an initial proof of concept and the rest is history.

# For Users
## How do I sign up?
1. Navigate to the site: https://twitchsongrequests-production.up.railway.app
1. Authorize the service with Twitch
1. Authorize the service with Spotify
1. If it worked, you should see a `Subscribe` and `Revoke Access` button
1. Click `Subscribe`. If successful, you should see the UI update accordingly
1. You're done! Now start using it!

## How do I use it?
1. Open Spotify on your computer
1. Navigate to your Twitch channel
1. Create a channel point reward with `TwitchSongRequests` somewhere in the title, that accepts text input
1. Start playing music on Spotify (the service needs an "active" player to work)
1. Find your favorite song on Spotify, copy the URI and use it as input when redeeming
1. If your Spotify player has queued the song, you're good!
1. If your player did not queue the song, make sure you copied the URL correctly, and use the right Channel Point reward
1. If you are pretty sure you didn't mess anything up, make an issue [here](https://github.com/SaxyPandaBear/TwitchSongRequests/issues/new?assignees=&labels=bug&template=User-Bug-Report.yml&title=%5BBug%5D%3A+%5BDescribe+the+issue%5D)

## Demo
![Demo](https://youtu.be/Oz5Zs8mVDRY)

## Who can use it?
In order to use custom channel point rewards, you must be an affiliate or partner
streamer.

Last I checked, in order to queue songs into a player, the user must have Spotify
Premium.

## How do I stop using it?
1. Navigate to the site: https://twitchsongrequests-production.up.railway.app
1. If you are fully authenticated, you should see a `Revoke Access` button
1. Click it. This will revoke access to Twitch, which means the application won't receive new channel point redemptions
1. Navigate to your [Spotify account](https://www.spotify.com/us/account/apps/) to revoke access for the service. This can't be done by the service, because they don't expose an API for it. See [here](https://github.com/spotify/web-api/issues/600) for why.
1. You're done!

## Troubleshooting
TBD

## How does it work?
The TwitchSongRequests service will authorize to your Twitch account so that the 
service can listen for channel point redemption events from your channel. Once the 
service receives an event, it will process that event in order to queue the given 
Spotify song into your connected Spotify player. 

The service requires access to read Twitch channel point redemptions, and also needs 
access to modify a user's Spotify playback state. The Twitch access is needed in order 
to allow the service to subscribe to the events, and the Spotify access is required to 
queue the song in a user's active player.

The site uses cookies to track whether a user is authenticated or not. 

## Why not YouTube?
Well, the YouTube API doesn't let you queue videos outside of an iframe player,
which means I would need to embed the player inside my website. Which, I'm not
necessariliy opposed to doing, but that wasn't the original point of this. 

If you want this feature, react and comment on [this issue](https://github.com/SaxyPandaBear/TwitchSongRequests/issues/131)

---

# Support the Project
I'm paying to host this service completely out of pocket. If you would like to help
pitch in for the cost of hosting (it's not that much right now), please let me know
and I'll set up a way to contribute in that way. 

---
# For Developers
## Running the project
Run the project locally:
```bash
go run main.go
# or
go build .
./twitchsongrequests
```

### Required variables
TODO - project env vars

## Testing

### Unit testing
```bash
go test ./... -cover -v -timeout 10s
```

### Set up Twitch CLI
Use the [Twitch CLI](https://dev.twitch.tv/docs/cli) for local development. 

### Set up an EventSub subscription for testing end-to-end

> Note: the Twitch CLI mock server does not support EventSub, so it can not be
> used for local testing of creating subscriptions. We must use the real API

Authenticate with the Twitch CLI:
```bash
twitch token
```

Then, use the app access token from the above command to authentication.

Get your own user payload, and note the user ID from the response.
```bash
twitch api get /users -q login=YourTwitchUsername
```

The response will look something like this:
```json
{
  "data": [
    {
      "broadcaster_type": "affiliate",
      "created_at": "2015-11-03T23:03:04Z",
      "description": "A description about the user",
      "display_name": "SaxyPandaBear",
      "id": "1234567890",
      "login": "saxypandabear",
      "offline_image_url": "https://some-image1.png",
      "profile_image_url": "https://some-image2.png",
      "type": "",
      "view_count": 1337
    }
  ]
}
```

Jot down the `id` in the JSON response body

Subscribe to the channel point reward redemption event, replacing the `USER_ID`,
`TOKEN`, `CLIENT_ID`, `CALLBACK_URL`, and `SUB_SECRET` with your own values.

```bash
curl -X POST https://api.twitch.tv/helix/eventsub/subscriptions \
-H 'Authorization: Bearer $TOKEN' \
-H 'Client-Id: $CLIENT_ID' \
-H 'Content-Type: application/json' \
-d '{"type": "channel.channel_points_custom_reward_redemption.add", "version": "1", "condition": {"broadcaster_user_id": "$USER_ID"}, "transport": {"method": "webhook", "callback":"$CALLBACK_URL", "secret": "$SUB_SECRET"}}'
```

Make a note of the `id` value from the response.

Verify that the subscription was made successfully:
```bash
twitch api get /eventsub/subscriptions -q user_id=$USER_ID
```

The response will look like:
```json
{
  "data": [
    {
      "condition": {
        "broadcaster_user_id": "1234567890",
        "reward_id": ""
      },
      "cost": 0,
      "created_at": "2023-01-29T04:23:21.008351071Z",
      "id": "abc-123-def-456",
      "status": "webhook_callback_verification_pending",
      "transport": {
        "callback": "https://this-service/endpoint",
        "method": "webhook"
      },
      "type": "channel.channel_points_custom_reward_redemption.add",
      "version": "1"
    }
  ],
  "pagination": {
    "cursor": ""
  },
  "total": 1
}
```

### Run local API server to handle callback
Start the server locally
```bash
export TWITCH_CLIENT_ID=$CLIENT_ID
go run main.go
```

Make sure that the server is up
```bash
curl localhost:8080/
```

### Create and send a test webhook payload to the local server
```bash
twitch event trigger add-redemption -s $SUB_SECRET -F http://localhost:8080/callback
```

### Check the server logs to confirm that the event was received and processed.
```
2022/09/05 10:49:46 verified signature for subscription
2022/09/05 10:49:46 52f644c8-33da-4a30-bc81-beccb4cb678a Test Reward from CLI
```

### Cleanup
After testing, clean up the EventSub subscription:
```bash
twitch api get /eventsub/subscriptions -q user_id=$USER_ID

# jot down the relevant EventSub subscription ID

curl -X DELETE https://api.twitch.tv/helix/eventsub/subscriptions?id=$SUB_ID -H 'Authorization: Bearer $TOKEN' -H 'Client-Id: $CLIENT_ID'
```

## Deploying
This service is deployed directly to Railway, via the supplied Dockerfile at the root of the repo.
