#!/bin/bash

# The purpose of this script is to take the zipped lambda function code and
# push it to S3. This will push to S3 in localstack. 
aws s3 cp build/distributions/queue-song-lambda.zip s3://twitch-song-requests --endpoint-url http://localhost:4566
