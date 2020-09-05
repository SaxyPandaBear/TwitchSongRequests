/**
 * Handles Twitch and Spotify Oauth token fetching
 */

const fetch = require("node-fetch");

module.exports = function () {
    // https://dev.twitch.tv/docs/authentication
    this.fetchTwitchToken = async function (clientId, clientSecret, authorizationCode) {
        let response = await fetch(`https://id.twitch.tv/oauth2/token?client_id=${clientId}&client_secret=${clientSecret}&grant_type=authorization_code&code=${authorizationCode}&redirect_uri=https://github.com/SaxyPandaBear/TwitchSongRequests`, {
            method: "POST",
            mode: "cors"
        });
        return response.json();
    };

    this.refreshTwitchToken = async function (clientId, clientSecret, refreshToken) {
        let request = {
            "grant_type": "refresh_token",
            "refresh_token": refreshToken,
            "client_id": clientId,
            "client_secret": clientSecret
        }
        
        let data = Object.
            entries(request).
            map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`).
            join("&");
        
        let response = await fetch("https://id.twitch.tv/oauth2/token", {
            method: "POST",
            mode: "cors",
            body: data,
            headers: {
                "Accept": "application/json",
                "Content-Type": "application/x-www-form-urlencoded"
            }
        });
        return response.json();
    }

    // // https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow
    this.fetchSpotifyToken = async function (clientId, clientSecret, authorizationCode) {
        let request = {
            "grant_type": "authorization_code",
            "code": authorizationCode,
            "redirect_uri": "https://github.com/SaxyPandaBear/TwitchSongRequests",
            "client_id": clientId,
            "client_secret": clientSecret
        }

        let data = Object.
            entries(request).
            map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`).
            join("&");

        let response = await fetch("https://accounts.spotify.com/api/token", {
            method: "POST",
            mode: "cors",
            body: data,
            headers: {
                "Accept": "application/json",
                "Content-Type": "application/x-www-form-urlencoded"
            }
        });
        return response.json();
    };

    this.refreshSpotifyToken = function (clientId, clientSecret, refreshToken) {
        let request = {
            "grant_type": "refresh_token",
            "refresh_token": refreshToken,
            "client_id": clientId,
            "client_secret": clientSecret
        }
        
        let data = Object.
            entries(request).
            map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`).
            join("&");
        
        let response = await fetch("https://accounts.spotify.com/api/token", {
            method: "POST",
            mode: "cors",
            body: data,
            headers: {
                "Accept": "application/json",
                "Content-Type": "application/x-www-form-urlencoded"
            }
        });
        return response.json();
    }
}
