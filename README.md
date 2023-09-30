# TwitchSongRequests

Integrate Twitch channel point rewards directly with a user's Spotify player.
Configure a channel point reward to accept a Spotify song URL, and enqueue
the song in the user's current playing session.

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/SaxyPandaBear/TwitchSongRequests/ci.yml?style=for-the-badge) ![Onboarded](https://img.shields.io/endpoint?url=https%3A%2F%2Ftwitchsongrequests-production.up.railway.app%2Fstats%2Fonboarded) ![Total number of songs queued](https://img.shields.io/endpoint?url=https%3A%2F%2Ftwitchsongrequests-production.up.railway.app%2Fstats%2Ftotal) ![Queued last 30 days](https://img.shields.io/endpoint?url=https%3A%2F%2Ftwitchsongrequests-production.up.railway.app%2Fstats%2Frunning%3Fdays%3D30)

---

# Inspiration
This project was inspired by [Kaije](https://www.twitch.tv/kaije), because their 
Twitch channel has a channel point reward for requesting a song on Spotify. On that 
fateful day, I requested multiple songs in quick succession, and they told me to hold 
off on requesting songs because they have to stop working on art in order to manually 
copy the Spotify URIs I was sending, in order to queue the songs in their Spotify 
player. Between jobs, I coded up an initial proof of concept and the rest is history.

# For Users

## Update 2023-05-19
Signing up now automatically creates a custom reward with a default cost and title.
This can be updated manually to fit your own needs. No more need to use a keyword in
the reward title! Thanks @rosethornttv for the feedback!

## How do I sign up?
1. Navigate to the site: https://twitchsongrequests-production.up.railway.app
1. Authorize the service with Twitch
1. Authorize the service with Spotify
1. If it worked, you should see a `Subscribe` and `Revoke Access` button
1. Click `Subscribe`. If successful, you should see the UI update accordingly
1. Submit an [onboarding request](https://github.com/SaxyPandaBear/TwitchSongRequests/issues/new?assignees=&labels=onboarding&template=Onboarding-Form.yml&title=Onboarding+Request) on this GitHub project, and I will manually allowlist your Spotify account to let you access the service (this is a limitation on Spotify, not on me)
1. Wait for me to manually onboard you and then you're done! Now start using it!

### From Spotify
> We appreciate your interest and efforts in using Spotify's open platform to innovate and build interesting integrations. However, after reviewing we found that your app does not comply with our terms and conditions for the following reasons:
> > The product or service is integrated with streams or content from another service. The application is integrated with other services (e.g. Twitch) in a way that is prohibited according to section III in our Developer Policy .

As much as I would love to make this more broadly accessible, looks like it goes against their ToS. This means that there will be limited access to the service, since I have to manually allowlist users. If you really want to use it, please sign up via the above steps. 

### What happens after all 25 slots are taken?
I am working on figuring out how to audit usage of my service, so that I can 
give priority to streamers that would actually use it. Even if the service is
at capacity, please still go through all of the onboarding steps. I will 
evaluate onboarding newer users first-in-first-out, so getting in line early
is to your own benefit!

## How do I use it?
1. Open Spotify on your computer
1. Navigate to your Twitch channel
1. Start playing music on Spotify (the service needs an "active" player to work)
1. Find your favorite song on Spotify, copy the URI
1. Navigate to the Twitch channel, and find the song request reward (the default title is `Spotify Song Request`)
1. Submit the Spotify URI as part of the reward redemption, e.g.: https://open.spotify.com/track/6mfiGqZw4AqXA1nqo3EzIF
1. If your Spotify player has queued the song, you're good!
1. If your player did not queue the song, make sure you copied the URL correctly, and use the right Channel Point reward
1. If you are pretty sure you didn't mess anything up, make an issue [here](https://github.com/SaxyPandaBear/TwitchSongRequests/issues/new?assignees=&labels=bug&template=User-Bug-Report.yml&title=%5BBug%5D%3A+%5BDescribe+the+issue%5D)
1. If you want to allow your viewers to submit song requests for [explicit songs](https://support.spotify.com/us/article/explicit-content/), you have to opt in to this by updating your [preferences](https://twitchsongrequests-production.up.railway.app/preferences)
1. If you want to limit the length of the songs chatters can submit, specify the max song length in seconds in your [preferences](https://twitchsongrequests-production.up.railway.app/preferences). Any value less than or equal to zero means any song length is allowed
1. If you want to display your song queue (current playing, and next two songs)
on your stream, use the link shown on the page for the OBS browser source. Feel
free to override the CSS with your own custom CSS to make it your own. The page
auto refreshes every 10 seconds. 

## Demo
[![Demo](https://img.youtube.com/vi/Oz5Zs8mVDRY/hqdefault.jpg)](https://youtu.be/Oz5Zs8mVDRY)

## Who can use it?
In order to use custom channel point rewards, you must be an affiliate or partner
streamer.

The Spotify account used must have Spotify Premium. If it turns out you do not
have Spotify Premium and you are onboarded but fail to queue songs because of 
that, I will take the liberty of removing you from the allowed users in order to
give a chance to other users who want to use the service. 

## How do I stop using it?
1. Navigate to the site: https://twitchsongrequests-production.up.railway.app
1. If you are fully authenticated, you should see a `Revoke Access` button
1. Click it. This will revoke access to Twitch, which means the application won't receive new channel point redemptions
1. Navigate to your [Spotify account](https://www.spotify.com/us/account/apps/) to revoke access for the service. This can't be done by the service, because they don't expose an API for it. See [here](https://github.com/spotify/web-api/issues/600) for why.
1. You're done!

## This was working but isn't anymore. What happened?
In order to better accommodate people who want to use this project, I am going to start actively revoking access
to streamers that have not used it in 30 days. Personally, I think it sucks, but it's the only way to fairly keep
up with demand considering I am only allowed to serve 25 users. I will keep track of Twitch users that I revoke access to 
via the `CHANGELOG` file.

### Query ran on the database to get Twitch IDs
```sql
select distinct broadcaster_id from messages 
  where broadcaster_id != '' 
  and success = 1 
  and age(messages.created_at) > 30 * INTERVAL '1 day' 
except select distinct broadcaster_id from messages 
  where broadcaster_id != '' 
  and age(messages.created_at) <= 30 * INTERVAL '1 day';
```

So the criteria are:
1. Had at least 1 successful song request redeemed more than 30 days ago
1. Has not had ANY redeems in the past 30 days

This allows for errors such as issues with the API, credentials, etc in the past 30 days, because
at least it was attempted.

### API call to get usernames
```bash
# for each user ID
twitch token
twitch api get users -q id=$ID
```

If you believe that I revoked your access in error, please feel free to open a GitHub issue to appeal it, otherwise
you'll want to submit a new onboarding request.

If you want to have better control over this and are willing to host the project yourself, I will be writing up
a guide on how to self-host. TBD

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
| Name                  | Purpose                                                          |
| --------------------- | ---------------------------------------------------------------- |
| PORT                  | Override default port for the HTTP server                        |
| DATABASE_URL          | PostgresDB URL to connect to                                     |
| SITE_REDIRECT_URL     | URL for the base path for the main site                          |
| TWITCH_SECRET         | Passphrase to verify subscription requests for Twitch EventSub   |
| TWITCH_CLIENT_ID      | Twitch app OAuth client ID                                       |
| TWITCH_CLIENT_SECRET  | Twitch app OAuth client secret                                   |
| TWITCH_STATE          | Twitch app OAuth state key                                       |
| TWITCH_REDIRECT_URL   | Twitch OAuth redirect URL                                        |
| MOCK_SERVER_URL       | Arbitrary mock URL for local testing with Twitch CLI mock server |
| SPOTIFY_CLIENT_ID     | Spotify app OAuth client ID                                      |
| SPOTIFY_CLIENT_SECRET | Spotify app OAuth client secret                                  |
| SPOTIFY_REDIRECT_URL  | Spotify OAuth redirect URL                                       |
| SPOTIFY_STATE         | Spotify app OAuth state key                                      |
| ONBOARDED_USERS       | Number of onboarded users to display stats for                   |
| ALLOWED_USERS         | Number of users that are allowed to be onboarded                 |

## Testing

### Unit testing
```bash
go test ./... -cover -v -timeout 10s -short
```

### Integration testing
Integration testing for Postgres queries is done via GitHub actions. 
Take a look at the `.github/workflows/ci.yml` file for how they are run.

Technically they can be run locally by standing up a Postgres db and injecting 
the same test data into a local instance. 

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
twitch api get /eventsub/subscriptions

# jot down the relevant EventSub subscription ID

curl -X DELETE https://api.twitch.tv/helix/eventsub/subscriptions?id=$SUB_ID -H 'Authorization: Bearer $TOKEN' -H 'Client-Id: $CLIENT_ID'
```

## Deploying
This service is deployed directly to Railway, via the supplied Dockerfile at the root of the repo.
