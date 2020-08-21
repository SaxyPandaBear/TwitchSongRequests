class Auth {

    /**
     * @param {*} url whole URL for fetching the auth token
     */
    constructor(url) {
        this.url = url;
        this.expirationTime = Date.now();
        this.activeToken = null;
    }

    get token(authPayload) {
        return this.retrieveToken(authPayload);
    }

    /**
     * Fetch an OAuth token from the given URL, storing the token
     * This also takes the time of expiration, transforms it to milliseconds
     * 
     * @param {*} payload The payload to be sent in order to get an OAuth token.
     *                    This is different for Spotify and Twitch authentication, 
     *                    so rather than duplicating code
     */
    retrieveToken(payload) {
        let currentTime = Date.now();
        if (this.activeToken == null || this.expirationTime <= currentTime) {
            let request = new XMLHttpRequest();
            request.open("POST", this.url, false);
            // TODO: figure out what to do for error handling here.
            request.onerror = function() {
                console.log("oopsie whoopsie");
            }
            request.send(payload);
            let response = JSON.parse(request.responseText);
            this.activeToken = response.access_token
            this.expirationTime = currentTime + (response.expires_in * 1000);
        } else {
            return this.activeToken;
        }
    }
}