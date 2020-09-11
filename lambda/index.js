/**
 * Lambda handler that accepts events from SQS and attempts to queue songs into
 * a connected Spotify player.
 */
const AWSXRay = require('aws-xray-sdk-core');
const AWS = AWSXRay.captureAWS(require('aws-sdk'));
const fetch = require('node-fetch');

// Create client outside of handler to reuse
const lambda = new AWS.Lambda();

let config = { apiVersion: '2012-08-10' };
// if we are running locally, add the required property for communicating with
// a local instance of DynamoDB
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.UsageNotes.html
if (process.env['env'] === 'local') {
    let ep = new AWS.Endpoint('http://localhost:8000');
    config.endpoint = ep;
}
var dynamo = new AWS.DynamoDB.DocumentClient(config);

/**
 * Find the first active computer device. If there is no active computer device to connect to,
 * return null
 * @param {array} devices
 */
function findFirstComputer(devices) {
    let computers = devices.filter(
        (device) => device.type === 'Computer' && device.is_active
    );
    if (computers.length < 1) {
        return null;
    } else {
        return computers[0];
    }
}

// Get all devices for a user
async function getDevices(oauth) {
    let response = await fetch('https://api.spotify.com/v1/me/player/devices', {
        method: 'GET',
        mode: 'cors',
        headers: {
            Authorization: `Bearer ${oauth}`,
            Accept: 'application/json',
        },
    });
    return response.json();
}

// queue a song URI for the given active player
async function queueSong(oauth, device, uri) {
    let response = await fetch(
        `https://api.spotify.com/v1/me/player/queue?uri=${uri}&device_id=${device.id}`,
        {
            method: 'POST',
            mode: 'cors',
            headers: {
                Authorization: `Bearer ${oauth}`,
                Accept: 'application/json',
            },
        }
    );
    return response.json();
}

async function placeholderGetRecord(channelId) {
    let response = {
        client_id: process.env['client_id'],
        client_secret: process.env['client_secret'],
        access_token: 'foobar',
        refresh_token: 'abc123',
        connection_status: 'Active',
    };
    return new Promise((() => response, () => 'Failed'));
}

async function refreshSpotifyToken(clientId, clientSecret, refreshToken) {
    let request = {
        grant_type: 'refresh_token',
        refresh_token: refreshToken,
        client_id: clientId,
        client_secret: clientSecret,
    };

    let data = Object.entries(request)
        .map(
            ([key, value]) =>
                `${encodeURIComponent(key)}=${encodeURIComponent(value)}`
        )
        .join('&');

    let response = await fetch('https://accounts.spotify.com/api/token', {
        method: 'POST',
        mode: 'cors',
        body: data,
        headers: {
            Accept: 'application/json',
            'Content-Type': 'application/x-www-form-urlencoded',
        },
    });
    return response.json();
}

// Handler
exports.handler = async function (event, context) {
    /**
     * for each message received from SQS, parse the event body,
     * to get the message, which should just be the Spotify entity URI.
     * We associate this request with a Twitch channel by the message
     * attribute appended to the message, which tells us which channel ID
     * this queue event belongs to.
     */
    event.Records.forEach((record) => {
        // get the message attributes to figure out which channel this request
        // is for
        let channelId = record.messageAttributes['channel_id'];
        let spotifyUri = record.body;
        // dynamo.getRecord(channelId) <- TODO: finish me
        placeholderGetRecord(channelId)
            .then((data) => {
                // if the connection statis is not active, then we shouldn't try to queue
                // a song.
                if (data.connection_status !== 'Active') {
                    console.log('User disconnected, dropping record');
                } else {
                    getDevices(data.access_token)
                        .then((foundDevices) => {
                            console.log(
                                `GET devices responded with ${JSON.stringify(
                                    foundDevices
                                )}`
                            );
                            let devices = foundDevices.devices;
                            let activeDevice = findFirstComputer(devices);
                            if (activeDevice === null) {
                                console.log(
                                    'No active device found. Write error to dynamo'
                                );
                            } else {
                                queueSong(
                                    data.access_token,
                                    activeDevice,
                                    spotifyUri
                                )
                                    .then((data) => {
                                        console.log(
                                            `Successfully queued song. Spotify responded with ${JSON.stringify(
                                                data
                                            )}`
                                        );
                                    })
                                    .catch((err) => {
                                        // need to check if the error is because of an invalid token. if it is, we need
                                        // to refresh it and write the new token back to dynamo.
                                        // TODO: how do we handle actual errors here?
                                        console.log(err);
                                        console.error('oopsie');
                                    });
                            }
                        })
                        .catch((err) => {
                            console.log(err);
                            // need to check if the error is because of an invalid token. if it is, we
                            // need to refresh it and write the new token back to dynamo.
                            // TODO: how do we handle actual errors? do we throw the message back in the queue?
                            //       or do we just drop the message, writing the error to dynamo for triaging?
                            console.error('oopsie');
                        });
                }
            })
            .catch((err) => {
                console.log(err);
                // I think if we fail to retrieve a record from dynamo, presumably because it
                // doesn't exist, then we SHOULD be able to assume that the user explicitly
                // chose to have us delete their data from the system (GDPR and CCPA?)
            });
    });
};
