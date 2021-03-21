# A placeholder shell script that contains commands to run in order to test the 
# functionality locally

# just so no one tries to run this as an actual script. this is just a place
# for me to hold all of the scripts in one place
exit 1 

# write an item to dynamo 
aws dynamodb put-item \
    --endpoint-url http://localhost:4566 \
    --table-name connections \
    --item file:///Users/andrew_huynh/Desktop/ddb.json

# scan table
aws dynamodb scan \
    --endpoint-url http://localhost:4566 \
    --table-name connections

# list queues
aws sqs list-queues --endpoint-url http://localhost:4566

# send message to SQS to queue a JJBA song
aws sqs send-message \
    --endpoint-url http://localhost:4566 \
    --message-body spotify:track:5Cjkfft7iRWJp4elZXgjkc \
    --message-attributes '{"channelId": {"DataType": "String", "StringValue": "106060203"}}' \
    --queue-url http://localhost:4566/000000000000/song-requests

# list log groups first
aws logs --endpoint-url http://localhost:4566 describe-log-groups

# list log streams for a log group 
aws logs describe-log-streams \
    --endpoint-url http://localhost:4566 \
    --log-group-name /aws/lambda/song-requests-lambda-38762d93

# fetch the logs for a log stream
aws logs get-log-events \
    --endpoint-url http://localhost:4566 \
    --log-group-name /aws/lambda/song-requests-lambda-38762d93 \
    --log-stream-name 2020/09/18/[LATEST]70ef1b81 >> log.json
