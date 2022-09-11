# TwitchSongRequests

Integrate Twitch channel point rewards directly with a user's Spotify player.
Configure a channel point reward to accept a Spotify song URL, and enqueue
the song in the user's current playing session.

## Running the project
TBD

## Testing

### Unit testing
```bash
go test -v ./...
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
twitch api get /users -q loginname=YourTwitchUsername
```

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

## Deploying
TBD
