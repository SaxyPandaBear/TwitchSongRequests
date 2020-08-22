/**
 * This client needs to handle:
 * 1. Authenticating against Spotify's Web Player API
 * 2. Connecting to a Spotify player
 * 3. Queueing song requests
 */
class Spotify {
    constructor(clientId, clientSecret, authUrl) {
        this.clientId = clientId;
        this.clientSecret = clientSecret;
        this.authUrl = authUrl;
    }

    // TODO: figure out how to get client credentials for Spotify
    //       note: adding songs to a current player's queue requires a user login to 
    //       grant access, which means no client_credentials granting.
}