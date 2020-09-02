/**
 * Main server code for handling requests from the UI. This should serve some APIs, 
 * the main one being an API that accepts data in order to initiate the workflow for 
 * listening to a Twitch channel. The event handler for listening to the twitch events
 * will handle processing the data.
 */

// TODO: figure out how to get dynamic properties set up so that it can run locally
//       and deployed
const properties = {
    "env": "LOCAL"
}

const express = require("express");
require("auth.js")();
var app = express();

/**
 * Handles validating the contents of a request that contains authentication details.
 * This is used for both Spotify and Twitch authorization handling.
 * 
 * We expect the body to contain an `authorization_code`
 * 
 * TODO: see if there's anything else needed
 * @param {*} body 
 * @returns true if authorization_code is in body, false otherwise
 */
function validateAuthRequestBody(body) {
    return "authorization_code" in body;
}

/**
 * Upserts the Twitch authentication details into the database,
 * and then starts a WebSocket connection to Twitch. 
 * 
 * This is meant to be asynchronous. The client-side UI should poll
 * the GET endpoint to check the status of the 
 */
app.post("/twitch", function (req, res) {
    if (!validateAuthRequestBody(req.body)) {
        res.status(400).send("Missing 'authorization_code' in request body");
    }
});

app.post("/spotify", function (req, res) {
    if (!validateAuthRequestBody(req.body)) {
        res.status(400).send("Missing 'authorization_code' in request body");
    }
});

app.listen(8080);
