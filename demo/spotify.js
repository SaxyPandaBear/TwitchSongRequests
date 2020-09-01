/**
 * POC for connecting to a Spotify player. This is extremely important in order to 
 * figure out what works and what doesn't.
 * 
 * 1. Get past authorization workflow
 * 2. Connect to player (this is actually implicit)
 * 3. Queue a song URI for the player.
 */

 // read in credentials
const fetch = require("node-fetch");
const credentials = require("../src/credentials.json");

// modify playback state in order to hit:
// POST https://api.spotify.com/v1/me/player/queue
// - https://developer.spotify.com/documentation/web-api/reference/player/add-to-queue/
// read playback state in order to hit:
// GET https://api.spotify.com/v1/me/player/devices
// - https://developer.spotify.com/documentation/web-api/reference/player/get-a-users-available-devices/
const scopes = "user-modify-playback-state+user-read-playback-state"
const songUri = "spotify:track:2TsyTag2aNa4wmUNmExzcI"

async function retrieveSpotifyOauthToken(clientId, clientSecret, authorizationCode) {
    let request = {
        "grant_type": "authorization_code",
        "code": authorizationCode,
        "redirect_uri": "https://github.com/SaxyPandaBear/TwitchSongRequests",
        "client_id": clientId,
        "client_secret": clientSecret
    }
    console.log(request);

    let data = Object.
        entries(request).
        map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`).
        join("&");

    // https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow
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

// Get all devices for a user
async function getDevices(oauth) {
    let response = await fetch("https://api.spotify.com/v1/me/player/devices", {
        method: "GET",
        mode: "cors",
        headers: {
            "Authorization": `Bearer ${oauth}`,
            "Accept": "application/json"
        }
    });
    return response.json();
}

// queue a song URI for the given active player
async function queueSong(oauth, device, uri) {
    let response = await fetch(`https://api.spotify.com/v1/me/player/queue?uri=${uri}&device_id=${device.id}`, {
        method: "POST",
        mode: "cors",
        headers: {
            "Authorization": `Bearer ${oauth}`,
            "Accept": "application/json"
        }
    });
    return response.json();
}

// given a list of Spotify device objects, find the first active computer
// device. if there isn't one, return null; 
function findFirstComputer(devices) {
    let computers = devices.filter(device => device.type === "Computer" && device.is_active);
    if (computers.length < 1) {
        return null;
    } else {
        return computers[0];
    }
}

retrieveSpotifyOauthToken(credentials.spotify_client_id, credentials.spotify_client_secret, credentials.spotify_auth_code)
    .then((oauth) => {
        console.log(`OAuth call responded with: ${JSON.stringify(oauth)}`);
        // This is going to be a little funky. 
        // GET the available devices (filtering for computer)
        // if there is no active computer source, we can't do more work
        // if there is, we queue the song up
        getDevices(oauth.access_token)
            .then((found) => {
                console.log(`GET devices responded with ${JSON.stringify(found)}`);
                let devices = found.devices;
                // check the list of devices by filtering for computer type devices, sorted by `is_active`
                // if the first device ID is active, then connect to it and queue the song.
                // else play the song on the first device
                let computer = findFirstComputer(devices);
                if (computer === null) {
                    // no active player to connect to, which means the rest of the player API endpoints
                    // won't work.
                    console.error("Oopsie");
                }
                else {
                    console.log(`Queueing song on active player: ${computer}`);
                    queueSong(oauth.access_token, computer, songUri)
                        .then((res) => {
                            // TODO: not sure how to handle a 204
                            console.log(`Queuing song responded with ${res.text()}`);
                        });
                }
            });
    });
