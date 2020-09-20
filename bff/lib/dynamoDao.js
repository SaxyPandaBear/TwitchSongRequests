const AWS = require('aws-sdk');
const { awsConfig } = require('../session');

var dynamodb = new AWS.DynamoDB(awsConfig);
var docClient = new AWS.DynamoDB.DocumentClient(awsConfig);
function queryDynamoByChannel(channelId) {
    const params = {
        TableName: 'connections',
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
        TableName: 'connections',
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
