const AWS = require('aws-sdk');

let config = { apiVersion: '2012-08-10', region: 'us-east-1' };
// see start-cloud.sh in the root of the project for context.
// use the existence of the LOCALSTACK key to know to use a different endpoint
if ('LOCALSTACK' in process.env) {
    console.info('======= RUNNING LAMBDA IN LOCALSTACK ======');
    AWS.config.update({ region: 'us-east-1' });
    let ep = new AWS.Endpoint(`http://localhost:4566`);
    config.endpoint = ep;
}
AWS.config.update(config);

var dynamodb = new AWS.DynamoDB(config);
var docClient = new AWS.DynamoDB.DocumentClient(config);
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
