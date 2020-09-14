const AWS = require('aws-sdk');
var dynamodb = new AWS.DynamoDB({
    region: 'us-east-1',
    endpoint: 'http://localhost:8000',
});

AWS.config.update({
    region: 'us-east-1',
    endpoint: 'http://localhost:8000',
});
var docClient = new AWS.DynamoDB.DocumentClient();
function queryDynamoByChannel(channelId) {
    const params = {
        TableName: 'twitch-sessions',
        Key: {
            channelId: { S: `${channelId}` },
        },
        ProjectionExpression: 'connectionStatus',
    };

    return new Promise((res, rej) => {
        dynamodb.getItem(params, function (err, data) {
            if (err) {
                rej(err);
            } else {
                res(data.Item);
            }
        });
    });
}

function updateConnectionStatusByChannelId(channelId, connectionStatus) {
    var params = {
        TableName: 'twitch-sessions',
        Key: {
            channelId: `${channelId}`,
        },
        UpdateExpression: 'set connectionStatus = :c',
        ExpressionAttributeValues: {
            ':c': connectionStatus,
        },
        ReturnValues: 'ALL_NEW',
    };
    return new Promise((res, rej) => {
        docClient.update(params, function (err, data) {
            if (err) {
                rej(err);
            } else {
                res(data);
            }
        });
    });
}
module.exports = { queryDynamoByChannel, updateConnectionStatusByChannelId };
