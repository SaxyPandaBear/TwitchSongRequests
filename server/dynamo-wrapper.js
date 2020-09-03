// TODO: finish this
const AWS = require("aws-sdk");
const properties = require("./properties.json");

AWS.config.update({region: properties.region});
// var docClient = new AWS.DynamoDB.DocumentClient({apiVersion: '2012-08-10'});

// TODO: will need put AWS credentials into this
let config = {apiVersion: '2012-08-10'}
// if we are running locally, add the required property for communicating with 
// a local instance of DynamoDB
// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/DynamoDBLocal.UsageNotes.html
if (properties.env === "local") {
    let ep = new AWS.Endpoint("http://localhost:8000");
    config.endpoint = ep;
}
var docClient = new AWS.DynamoDB.DocumentClient(config);

function upsertAuthentication() {
    
}
