# TwitchSongRequests

Integrate Twitch channel point rewards directly with a user's Spotify player.
Configure a channel point reward to accept a Spotify song URL, and enqueue
the song in the user's current playing session.

## How do I use it?
1. Authorize the service with Twitch: [<kbd> <br>Authorize<br> </kbd>](https://id.twitch.tv/oauth2/authorize?response_type=code&client_id=vahfw8gww3oq9g57efph9rjcmtbxwk&redirect_uri=https://twitchsongrequests-production.up.railway.app/oauth/twitch&scope=channel%3Aread%3Aredemptions)
1. Authorize the service with Spotify: **TODO**
1. Navigate to your Twitch channel
1. Redeem a channel point reward with a Spotify URI as input
1. Check if your Spotify player has queued the song

### Example
TDB

## Troubleshooting
TBD

## How does it work?
The TwitchSongRequests service will authorize to your Twitch account so that the service can listen for
channel point redemption events from your channel. Once the service receives an event, it will process that
event in order to queue the given Spotify song into your connected Spotify player. 

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
