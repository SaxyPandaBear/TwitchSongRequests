/**
 * Lambda handler that accepts events from SQS and attempts to queue songs into
 * a connected Spotify player.
 */
const AWS = require('aws-sdk');
AWS.config.update({ region: 'us-east-1' });
const fetch = require('node-fetch');

const TABLE_NAME = 'connections';

// Localstack dynamo points to localhost:4566, but to connect to it from within the 
// lambda container, we need the LOCALSTACK_HOSTNAME env variable
let config = { apiVersion: '2012-08-10', region: 'us-east-1' }
if ('LOCALSTACK_HOSTNAME' in process.env) {
    console.log('======= RUNNING LAMBDA IN LOCALSTACK ======');
    let ep = new AWS.Endpoint(`http://${process.env['LOCALSTACK_HOSTNAME']}:4566`);
    config.endpoint = ep;
}
console.log(`AWS Configs: ${JSON.stringify(config)}`);
var dynamo = new AWS.DynamoDB(config);

/**
 * Find the first active computer device. If there is no active computer device to connect to,
 * return null
 * @param {array} devices
 */
function findFirstComputer(devices) {
    console.log("findFirstComputer");
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
    console.log("getDevices");
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
    console.log("queueSong");
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

function fetchConnectionDetails(channelId) {
    const params = {
        TableName: TABLE_NAME,
        Key: {
            channelId: {
                "S": `${channelId}`
            }
        },
        ConsistentRead: true,
        ProjectionExpression: "sess, connectionStatus"
    };
    console.log("Getting stuff from dynamo");
    return dynamo.getItem(params, function (err, data) {
        if (err) {
            console.log('oopsie whoopsie ddb fetch failed', err);
            console.log(err);
        } else {
            console.log('found a record', data.Item);

            return data.Item;
        }
    }).promise();
}

async function refreshSpotifyToken(clientId, clientSecret, refreshToken) {
    let request = {
        'grant_type': 'refresh_token',
        'refresh_token': refreshToken,
        'client_id': clientId,
        'client_secret': clientSecret
    }

    let data = Object.
        entries(request).
        map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`).
        join('&');

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
exports.handler = async function (event, context, callback) {
    const clientId = process.env['SpotifyClientId'];
    const clientSecret = process.env['SpotifyClientSecret'];

    console.log(`Client ID: ${clientId}`);
    console.log(`Client secret: ${clientSecret}`);

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
        console.log("Received a new record to process")
        let channelId = record.messageAttributes['channelId']["stringValue"];
        let spotifyUri = record.body;

        console.log(`Channel ID is: ${channelId}`);
        console.log(`Spotify URI is: ${spotifyUri}`);

        fetchConnectionDetails(channelId).then((data) => {
            console.log("Got our stuff from dynamo");
            console.log(data);
            // if the connection statis is not active, then we shouldn't try to queue
            // a song.
            if (data.connectionStatus !== 'Active') {
                console.log('User disconnected, dropping record');
            } else {
                getDevices(data.access_token).then(foundDevices => {
                    console.log(`GET devices responded with ${JSON.stringify(foundDevices)}`);
                    let devices = foundDevices.devices;
                    let activeDevice = findFirstComputer(devices);
                    if (activeDevice === null) {
                        console.log('No active device found. Write error to dynamo');
                    } else {
                        queueSong(data.access_token, activeDevice, spotifyUri).then(data => {
                            console.log(`Successfully queued song. Spotify responded with ${JSON.stringify(data)}`);
                        }).catch(err => {
                            // need to check if the error is because of an invalid token. if it is, we need
                            // to refresh it and write the new token back to dynamo.
                            // TODO: how do we handle actual errors here? 
                            console.log(err);
                            console.error('oopsie');
                        })
                    }
                }).catch(err => {
                    console.log(err);
                    // need to check if the error is because of an invalid token. if it is, we
                    // need to refresh it and write the new token back to dynamo. 
                    // TODO: how do we handle actual errors? do we throw the message back in the queue? 
                    //       or do we just drop the message, writing the error to dynamo for triaging?
                    console.error('oopsie');
                });
            }
        }).catch(err => {
            console.log("oopsie");
            console.log(err);
            // I think if we fail to retrieve a record from dynamo, presumably because it 
            // doesn't exist, there's not much to do. We can just swallow the exception here.
            // Does it make sense to write this error to the events table if we couldn't look
            // up the channel ID in the connections table?
        });
        console.log("processed record");
    });
};
