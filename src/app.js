/**
 * Main app code
 * 
 * TODO: clean up all of this code
 */

const fetch = require("node-fetch");
const WebSocket = require("ws");

const scopes = "channel_read+channel:read:redemptions";
const channelId = "106060203"
var oauth;

// connects to Twitch's WSS event stream for channel point redemptions
function connect() {

    function heartbeat() {
        message = {
            type: "PING",
        };
        console.log("SENT: " + JSON.stringify(message));
        ws.send(JSON.stringify(message));
    }
    var heartbeatInterval = 1000 * 60; //ms between PING's
    var reconnectInterval = 1000 * 3; //ms to wait before reconnect
    var heartbeatHandle;

    ws = new WebSocket("wss://pubsub-edge.twitch.tv");

    ws.onopen = function (event) {
        console.log("INFO: Socket Opened");
        heartbeat();
        heartbeatHandle = setInterval(heartbeat, heartbeatInterval);

        var listenEvent = {
            type: "LISTEN",
            nonce: "abc123",
            data: {
                topics: [`channel-points-channel-v1.${channelId}`], // TODO: generify the channel id
                auth_token: oauth.access_token
            },
        };
        ws.send(JSON.stringify(listenEvent));
    };

    ws.onerror = function (error) {
        console.log("ERR: " + JSON.stringify(error));
    };

    ws.onmessage = function (event) {
        // TODO: add Spotify integration here
        message = JSON.parse(event.data);
        console.log({ message });
        if (message.type == "RECONNECT") {
            console.log("INFO: Reconnecting...");
            setTimeout(connect, reconnectInterval);
        }
    };

    ws.onclose = function () {
        console.log("INFO: Socket Closed");
        clearInterval(heartbeatHandle);
        console.log("INFO: Reconnecting...");
        setTimeout(connect, reconnectInterval);
    };
}

// TODO: This is not correct. need to clean this up.
// We need a user access token, which is a different authorization flow than the
// client credentials. The user access token requires manual intervention from
// the user, in a browser, and cannot be done completely server-side. 
// get auth token from Twitch with the correct scope
async function retrieveToken(clientId, clientSecret) {
    const response = await fetch(`https://id.twitch.tv/oauth2/token?client_id=${clientId}&client_secret=${clientSecret}&grant_type=client_credentials&redirect_uri=https://github.com/SaxyPandaBear/TwitchSongRequests&scopes=${scopes}`, {
        method: "POST",
        mode: "cors"
    });
    return response.json();
}

// this fetches the OAuth token, then connects to the WSS.
retrieveToken()
    .then((data) => {
        console.log({ data });
        oauth = data;
        connect();
    });
