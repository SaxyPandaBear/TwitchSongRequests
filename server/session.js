const AWS = require('aws-sdk');
const session = require('express-session');
const DynamoDBStore = require('connect-dynamodb')(session);
const cookieIsPresent = require('./lib/getCookies');
AWS.config.update({
    region: 'us-east-1',
    endpoint: process.env.DYNAMO_ENDPOINT || 'http://localhost:8000',
});

const options = {
    // Optional DynamoDB table name, defaults to 'sessions'
    table: 'twitch-sessions',

    // Optional path to AWS credentials and configuration file
    // AWSConfigPath: './path/to/credentials.json',

    // Optional JSON object of AWS credentials and configuration

    // Optional client for alternate endpoint, such as DynamoDB Local
    client: new AWS.DynamoDB({
        endpoint: new AWS.Endpoint('http://localhost:8000'),
    }),
    hashKey: 'channelId',
    prefix: 'channel:',

    // Optional ProvisionedThroughput params, defaults to 5
    // readCapacityUnits: 25,
    // writeCapacityUnits: 25,
};

function checkForExistingSessionAndAssignAccessKeys(req, res, next) {
    if (!req.session.accessKeys) {
        req.session.accessKeys = {};
    }
    next();
}

function intializeSesionStore() {
    AWS.config.update({
        region: 'us-east-1',
        endpoint: 'http://localhost:8000',
    });

    return session({
        store: new DynamoDBStore(options),
        //TODO: use a more robust secret
        secret: 'keyboard cat',
        resave: false,
        saveUninitialized: true,
        genid: function (req) {
            return req.channelId;
        },
    });
}
function assignTwitchTokenToSession(req, res, next) {
    const { twitchTokenConfiguration } = req;
    console.log({ twitchTokenConfiguration });
    req.session.accessKeys.twitchToken = twitchTokenConfiguration;
    console.log({ session: req.session });

    next();
}

function assignSpotifyTokenToSession(req, res, next) {
    const { spotifyTokenConfiguration } = req;
    req.session.accessKeys.spotifyToken = spotifyTokenConfiguration;
    next();
}
function intializeSesionStoreIfCookieIsPresentInRequest(req, res, next) {
    if (cookieIsPresent(req)) {
        return intializeSesionStore()(req, res, next);
    } else {
        return next();
    }
}
module.exports = {
    intializeSesionStore,
    checkForExistingSessionAndAssignAccessKeys,
    assignTwitchTokenToSession,
    assignSpotifyTokenToSession,
    intializeSesionStoreIfCookieIsPresentInRequest,
};
