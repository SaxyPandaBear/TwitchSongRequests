/**
 * Lambda handler that accepts events from SQS and attempts to queue songs into
 * a connected Spotify player.
 */
const AWS = require('aws-sdk');
const fetch = require('node-fetch');

const TABLE_NAME = 'connections';

// Localstack dynamo points to localhost:4566, but to connect to it from within the 
// lambda container, we need the LOCALSTACK_HOSTNAME env variable
let config = { apiVersion: '2012-08-10', region: 'us-east-1' }
if ('LOCALSTACK_HOSTNAME' in process.env) {
    console.info('======= RUNNING IN LOCALSTACK ======');
    AWS.config.update({ region: 'us-east-1' });
    let ep = new AWS.Endpoint(`http://${process.env['LOCALSTACK_HOSTNAME']}:4566`);
    config.endpoint = ep;
}
var dynamo = new AWS.DynamoDB(config);

// return true if there is no error in the object, false otherwise
function isSuccessfulResponseObj(obj) {
    if (Object.keys(obj).find(key => key === 'error')) {
        return false;
    } else {
        return true;
    }
}

// Get all devices for a user
async function getDevices(accessToken) {
    let response = await fetch('https://api.spotify.com/v1/me/player/devices', {
        method: 'GET',
        mode: 'cors',
        headers: {
            Authorization: `Bearer ${accessToken}`,
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
    if (response.ok) {
        return {};
    } else {
        return { 'error': response.text() };
    }
}

function fetchConnectionDetails(channelId) {
    const params = {
        TableName: TABLE_NAME,
        Key: {
            channelId: {
                'S': `${channelId}`
            }
        },
        ConsistentRead: true,
        ProjectionExpression: 'sess, connectionStatus'
    };
    return dynamo.getItem(params).promise();
}

function updateAccessToken(channelId, session) {
    const params = {
        TableName: TABLE_NAME,
        Key: {
            channelId: {
                'S': `${channelId}`
            }
        },
        UpdateExpression: 'set sess=:session',
        ExpressionAttributeValues: {
            ':session': {'S': JSON.stringify(session)}
        },
        ReturnValues: 'UPDATED_NEW'
    }
    return dynamo.updateItem(params).promise();
}

/**
 * When we get an error from Spotify about expired auth, it looks like this:
 * { error: { status: 401, message: 'The access token expired' } }
 * In this case, we want to catch it, refresh our access token, and write it back to
 * the database. 
 * @param {string} clientId 
 * @param {string} clientSecret 
 * @param {string} refreshToken 
 */
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

    /**
     * for each message received from SQS, parse the event body,
     * to get the message, which should just be the Spotify entity URI.
     * We associate this request with a Twitch channel by the message
     * attribute appended to the message, which tells us which channel ID
     * this queue event belongs to.
     * 
     * https://stackoverflow.com/questions/37576685/using-async-await-with-a-foreach-loop
     * Because the handler is async, we have to await for a response on the stuff we have
     * in here, or else it will fire and forget, and we will not see any result in the 
     * lambda logs. On top of that, a foreach is async. 
     */
    for (const record of event.Records) {
        // get the message attributes to figure out which channel this request
        // is for
        let channelId = record.messageAttributes['channelId']['stringValue'];
        let spotifyUri = record.body;

        let connectionDetails;
        try {
            connectionDetails = await fetchConnectionDetails(channelId);
        } catch (error) {
            console.error(error);
            throw error;
        }
        // if the connection statis is not active, then we shouldn't try to queue
        // a song.
        if (connectionDetails.Item.connectionStatus.S !== 'active') {
            console.info('User is not connected to the main service. Dropping record');
        } else {
            // need to parse the session object to get the spotify credentials
            let sessionObj = JSON.parse(connectionDetails.Item.sess.S);
            let accessToken = sessionObj.accessKeys.spotifyToken.access_token;
            let refreshToken = sessionObj.accessKeys.spotifyToken.refresh_token; // TODO: use refresh token

            let foundDevices;
            try {
                foundDevices = await getDevices(accessToken);
            } catch (err) {
                console.error(err);
                throw err;
            }

            // check for an unauthorized response
            if (!isSuccessfulResponseObj(foundDevices)) {
                // we only want to do it on the case where the status is 401.
                if (foundDevices.error.status === 401) {
                    let refreshTokenResponse;
                    try {
                        refreshTokenResponse = await refreshSpotifyToken(clientId, clientSecret, refreshToken);
                    } catch (err) {
                        console.error(err);
                    }
                    if (!isSuccessfulResponseObj(refreshTokenResponse)) {
                        // now we're just unlucky.
                        throw refreshTokenResponse.error;
                    } else {
                        // write the response back to the database
                        sessionObj.accessKeys.spotifyToken.access_token = refreshTokenResponse.access_token;
                        let updateSessionResponse;
                        try {
                            updateSessionResponse = await updateAccessToken(channelId, sessionObj);
                        } catch (err) {
                            console.error(err);
                            throw err;
                        }
                    }                
                } else {
                    console.error(foundDevices);
                }
                // we throw an error so that the Lambda does not fully consume the SQS message,
                // and retries it.
                throw JSON.stringify(foundDevices.error);
            }
            
            let devices = foundDevices.devices;
            let activeDevice = devices.find(device => device.type === 'Computer' && device.is_active);
            if (activeDevice === undefined) {
                // TODO: after events table is implemented, write an error about this to the table
                console.info('There was no active player that Spotify could connect to');
            } else {
                let queueResponse;
                try {
                    queueResponse = await queueSong(accessToken, activeDevice, spotifyUri);
                } catch (err) {
                    console.error(err);
                    throw err;
                }
                if (!isSuccessfulResponseObj(queueResponse)) {
                    // TODO: after events table is implemented, write an error about this to the table
                    console.error(`Spotify responded with an error: ${queueResponse.error}`);
                } else {
                    console.info(`Successfully queued song. Spotify responded with ${JSON.stringify(queueResponse)}`);
                }
            }
        }
    }
    return 'Successfully processed all records';
};
