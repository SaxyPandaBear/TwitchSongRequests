# TwitchSongRequestsUi

TwitchSongRequestsUI

# Providing Client ID for Spotify and Twitch

This application relies upon OAuth to grant access to user resources. To run this application locally, you must supply the following properties: `twitchClientId` and `spotifyClientId`.

As per usual Angular patterns, these two properties are supplied in the `environment.ts` file located at `ui/src/environment.ts`

This file simply exports this object to be consumed by the dependent components: a sample object of the environmnet will posess the following strcuture:

```js
export const environment = {
    production: false,
    twitchClientId: 'mySuperPublicTwitchClientId',
    spotifyClientId: 'mySuperPublicSpotifyClientId',
};
```

To run this application, you must provide your own pair of client ids. These clients ids must match the client ids found in the server-side configuration.
