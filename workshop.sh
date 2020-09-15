# A placeholder shell script that contains commands to run in order to test the 
# functionality locally

# just so no one tries to run this as an actual script. this is just a place
# for me to hold all of the scripts in one place
exit 1 

# write an item to dynamo 
aws dynamodb put-item \
    --endpoint-url http://localhost:4566 \
    --table-name connections \
    --item file://sandbox/demo-ddb-record.json

# list queues
aws sqs list-queues --endpoint-url http://localhost:4566

# send message to SQS to queue a JJBA song
aws sqs send-message \
    --endpoint-url http://localhost:4566 \
    --message-body spotify:track:2TsyTag2aNa4wmUNmExzcI \
    --message-attributes '{"channelId": {"DataType": "String", "StringValue": "1599883101"}}' \
    --queue-url http://localhost:4566/000000000000/song-requests

# list log groups first
aws logs --endpoint-url http://localhost:4566 describe-log-groups

# list log streams for a log group 
aws logs describe-log-streams \
    --endpoint-url http://localhost:4566 \
    --log-group-name /aws/lambda/song-requests-lambda-c8d557f6

# fetch the logs for a log stream
aws logs get-log-events \
    --endpoint-url http://localhost:4566 \
    --log-group-name /aws/lambda/song-requests-lambda-c8d557f6 \
    --log-stream-name 2020/09/15/[LATEST]44a14a01 >> sandbox/log.json
