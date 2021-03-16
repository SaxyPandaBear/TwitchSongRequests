TwitchSongRequestsUi
====================

> [wiki](https://github.com/SaxyPandaBear/TwitchSongRequests/wiki/Architecture-Deep-Dive#web-application)

### Providing Client ID for Spotify and Twitch

This application relies upon OAuth to grant access to user resources.
To run this application locally, you must supply the following properties:
`twitchClientId` and `spotifyClientId`.

As per usual Angular patterns, these two properties are supplied in the
`environment.ts` file located at `ui/src/environment/*.ts`

This file simply exports this object to be consumed by the dependent components:
a sample object of the environmnet will posess the following strcuture:

```js
export const environment = {
    production: false,
    twitchClientId: 'mySuperPublicTwitchClientId',
    spotifyClientId: 'mySuperPublicSpotifyClientId',
};
```

To run this application, you must provide your own pair of client ids.
These clients ids must match the client ids found in the server-side configuration.

### Developing

For local development purposes, we have the `local` environment we can use, located at:
`./src/environment/environment.local.ts`. We can then use `ng build --local` which will
replace the environment object with the contents we have in our local config file.

### Build
`gradle assemble`

> Can build the UI with the underlying Angular CLI command: `ng build --aot --local`

### Test
`gradle test`

> Can run the UI tests with the underlying Angular CLI command: `ng test`

### Run locally
`gradle run`

> Can run the UI with the underlying Angular CLI command: `ng serve -o --aot --local`
