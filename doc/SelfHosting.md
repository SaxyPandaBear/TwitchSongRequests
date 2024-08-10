Hosting the project yourself
============================
In order to run this project yourself, you have two main options - self host on [Railway](https://railway.app), 
or run it from your computer with [Docker](https://www.docker.com).

The intent of this doc is to help make the process as painless as possible, but it's still cumbersome. 

## Twitch API access
1. Go to https://dev.twitch.tv/console and click `Register Your Application`
1. Fill out the info for the application
    1. For now, set the Redirect URL to `http://localhost:8000` as a placeholder. 
1. Jot down the client ID and client secret. These are necessary for authenticating your service against Twitch.

## Spotify API access
1. Go to https://developer.spotify.com/dashboard and click `Create app`
1. Fill out the info for the application
    1. For now, set the Redirect URI to `http://localhost:8000` as a placeholder.
1. This will give you a client ID and client secret. These are necessary for authenticating your service against Spotify.

## Railway
### Launching with a service template

[![Deploy on Railway](https://railway.app/button.svg)](https://railway.app/template/DoTt23?referralCode=a3qIN3)

This template will require you to input values for the application. The values are described below:

| Name                  | Purpose                                                          |
| --------------------- | ---------------------------------------------------------------- |
| PORT                  | Override default port for the HTTP server                        |
| TWITCH_SECRET         | Passphrase to verify subscription requests for Twitch EventSub   |
| TWITCH_CLIENT_ID      | Twitch app OAuth client ID                                       |
| TWITCH_CLIENT_SECRET  | Twitch app OAuth client secret                                   |
| TWITCH_STATE          | Twitch app OAuth state key                                       |
| SPOTIFY_CLIENT_ID     | Spotify app OAuth client ID                                      |
| SPOTIFY_CLIENT_SECRET | Spotify app OAuth client secret                                  |
| SPOTIFY_STATE         | Spotify app OAuth state key                                      |

**IMPORTANT NOTE**: the `PORT` value MUST be 443 to work in Railway.

Note that there is a difference between the `TWITCH_SECRET` and the `TWITCH_CLIENT_SECRET`. 
The `TWITCH_SECRET`, `TWITCH_STATE`, and `SPOTIFY_STATE` can be arbitrary passphrases. They are used as an added
layer of security for accessing their APIs. 

### Configuring OAuth redirects
The redirect URL in the project will be correct because it derives the domain from `RAILWAY_PUBLIC_DOMAIN`, 
which is an environment variable injected by Railway into the application. The problem is that this will not
match what you have for the redirect URL from when you registered for API access in Twitch and Spotify.

Copy the value for `RAILWAY_PUBLIC_DOMAIN`, as we will need it in two spots.

Go back to the Twitch [dev console](https://dev.twitch.tv/console) and manage your application. Update the
redirect URL to be `https://{RAILWAY_PUBLIC_DOMAIN}/oauth/twitch`. Note that if you don't copy domain exactly,
it will not match what the application sends, which will result in an API rejection from Twitch.

Similarly, go back to the Spotify [dev dashboard](https://developer.spotify.com/dashboard), go into the app and click
`Settings` on the top right. Scroll to the bottom to hit `Edit`, and update the Redirect URI for the application to be
`https://{RAILWAY_PUBLIC_DOMAIN}/oauth/spotify`.  

### Bootstrapping the database

**IMPORTANT NOTE**:
> This step is optional. The database is required if you want to interact with the website, which is not required functionality. 
If you do not plan on hosting for other people to use, then skip this step. Otherwise, continue reading. 

You can go into the Railway UI and configure the required tables yourself, or you can
log into the Postgres database and run the SQL script in `/testdata/bootstrap/db.sql` for
the table creation.

Once the tables are created, go into your Railway application and remove the `SKIP_POSTGRES` variable, and 
redeploy.

## Locally from your computer

TODO: need to iron this out because I'm not sure if this works with Twitch EventSub. 
