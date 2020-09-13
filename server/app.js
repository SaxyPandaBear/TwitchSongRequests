/**
 * Main server code for handling requests from the UI. This should serve some APIs,
 * the main one being an API that accepts data in order to initiate the workflow for
 * listening to a Twitch channel. The event handler for listening to the twitch events
 * will handle processing the data.
 */

// TODO: figure out how to get dynamic properties set up so that it can run locally
//       and deployed.
//       Use an environment variable to define the properties that we want to use.
const properties = {
    env: 'local',
};
// const properties = require("./properties.json");

const express = require('express');
const bodyParser = require('body-parser');

const AWS = require('aws-sdk');
const { intializeSesionStoreIfCookieIsPresentInRequest } = require('./session');
const cors = require('./utils/cors');
const sessionAuthRoutes = require('./api/session-auth');
const connectionStatusRoutes = require('./api/connection-status');

var app = express();

// parse application/x-www-form-urlencoded
app.use(bodyParser.urlencoded({ extended: false }));

// parse application/json
app.use(bodyParser.json());

app.use(cors);

app.use(intializeSesionStoreIfCookieIsPresentInRequest);

/**
 * Create an SDK client for DynamoDB so we can use it to read/write records from the data store.
 */
let config = { apiVersion: '2012-08-10' };
// if we are running locally, add the required property for communicating with
// a local instance of DynamoDB
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.UsageNotes.html
if (properties.env === 'local') {
    let ep = new AWS.Endpoint('http://localhost:8000');
    config.endpoint = ep;
}

/**
 * Handles validating the contents of a request that contains authentication details.
 * This is used for Twitch authorization handling.
 *
 * We expect the body to contain an `authorization_code` key and the value should
 * be non-empty.
 *
 * TODO: see if there's anything else needed
 * @param {object} body
 * @returns true if authorization_code is in body, false otherwise
 */
function validateTwitchAuthRequestBody(body) {
    if (!('authorization_code' in body)) {
        return false;
    }

    let auth = body['authorization_code'];
    return auth && auth.length > 0;
}

/**
 * Validates whether or not the payload is valid for the Spotify endpoint.
 * This includes an `authorization_code` value, and a `id` value.
 * The first is the authorization code retrieved from the user log in, and the
 * id is the Twitch channel ID.
 * @param {object} body
 */
function validateSpotifyAuthRequestBody(body) {
    if (!('authorization_code' in body)) {
        console.warn(
            `Request ${JSON.stringify(
                body
            )} does not contain an authorization code.`
        );
        return false;
    } else if (!('id' in body)) {
        console.warn(
            `Request ${JSON.stringify(
                body
            )} does not contain a Twitch channel ID.`
        );
        return false;
    }

    let auth = body['authorization_code'];
    let twitchId = body['id'];
    return auth && twitchId && auth.length > 0 && twitchId.length > 0;
}

// basic ping endpoint to get the ball rolling
app.get('/ping', function (req, res) {
    res.status(200).send('pong');
});

/**
 * Upserts the Twitch authentication details into the database,
 * and then starts a WebSocket connection to Twitch.
 *
 * This is meant to be asynchronous. The client-side UI should poll
 * the GET endpoint to check the status of the connection.
 *
 * This endpoint should do a synchronous fetch for the Twitch channel ID
 * associated with this request in order to use it as an ID in our database,
 * and to return to the UI so it can poll afterwards.
 */
app.post('/twitch', function (req, res) {
    if (!validateTwitchAuthRequestBody(req.body)) {
        res.status(400).send("Missing 'authorization_code' in request body");
        return;
    }
    res.status(202).send('Twitch Channel ID');
});

/**
 * Retrieve an OAuth token for Spotify with the given authorization code,
 * and insert it into the record with the given twitch channel ID
 */
app.post('/spotify', function (req, res) {
    if (!validateSpotifyAuthRequestBody(req.body)) {
        res.status(400).send("Missing 'authorization_code' in request body");
    }
    res.status(201).send('Spotify credentials saved');
});

app.use('/api/session', sessionAuthRoutes);
app.use('/api/connection-status', connectionStatusRoutes);

app;
app.listen(process.env.PORT || 8080, () =>
    console.log(`Listening to port ${process.env.PORT || 8080}`)
);
