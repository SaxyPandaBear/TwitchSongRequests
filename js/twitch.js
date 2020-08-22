/**
 * 1. Retrieving/refreshing OAuth token
 * 2. Websocket connection to read from Twitch event topic
 * 3. Websocket recurring ping
 * 4. Websocket reconnect
 * 5. GET channel ID
 * 6. Parse WSS event from Twitch for channel point redemption
 */
class Twitch {
    
    /**
     * @param {*} clientId     App client ID
     * @param {*} clientSecret App client secret
     * @param {*} authUrl      OAuth URL
     * @param {*} wsUrl        Websocket URL
     * @param {*} rewardTitle  Title of the Channel Point reward, to match with this.
     * @param {*} spotify      Spotify client
     */
    constructor(clientId, clientSecret, authUrl, wsUrl, rewardTitle, spotify) {
        this.clientId = clientId;
        this.clientSecret = clientSecret;
        this.authUrl = authUrl;
        this.wsUrl = wsUrl;
        this.rewardTitle = rewardTitle;
        this.spotify = spotify;
        this.expirationTime = Date.now();
        this.activeToken = null;
    }
    
    /**
     * Twitch currently supports two different APIs. Subscribing to the channel point topic
     * requires the new API, but also requires fetching a channel ID from the old API. 
     */
    retrieveHelixToken() {
        let currentTime = Date.now();
        if (this.activeToken == null || this.expirationTime <= currentTime) {
            let request = new XMLHttpRequest();
            let url = `${this.authUrl}?client_id=${this.clientId}&client_secret=${this.clientSecret}&grant_type=client_credentials&scope=channel:read:redemptions` 
            request.open("POST", url, false);
            // TODO: figure out what to do for error handling here.
            request.onerror = function() {
                console.log("oopsie whoopsie");
            }
            request.send();
            let response = JSON.parse(request.responseText);
            this.activeToken = response.access_token
            this.expirationTime = currentTime + (response.expires_in * 1000);
        } else {
            return this.activeToken;
        }
    }

    /**
     * Process an incoming WSS event from Twitch.
     * If the event is redeeming channel points, then we act on it by taking the user input from the 
     * request, and attempting to add the assumed Spotify URI to the Spotify player's queue.
     * @param {*} event Twitch channel event
     * 
     * @returns true if successfully processed event, false if the event was not the right type, or was not successfully processed
     */
    processChannelPointEvent(event) {
        // if the type of the event is NOT "reward-redeemed", then we drop it
        // TODO: this might not be necessary because the topic that we are subscribing to should only contain
        //       channel point redemptions.
        if (event.type !== "reward-redeemed") {
            return;
        }

        let message = JSON.parse(event.data)
        // 1. check if the title of the redemption matches our reward title that we expect is used for Spotify song requests
        // 2. If it is, then caputre the user input (expected to be a Spotify URI), and send it off to the Spotify client
        if (message.redemption.reward.title === this.rewardTitle) {
            let spotifyUri = message.user_input
            console.log(`Sending request to Spotify to queue ${spotifyUri}`)
            // TODO: add spotify client integration here.
            this.spotify.queue(spotifyUri)
        }
    }
}