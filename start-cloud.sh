#!/bin/bash

# Usage:
#   ./start-cloud.sh twitch_client_id twitch_client_secret spotify_client_id spotify_client_secret

# This assumes that we are already in the pipenv shell!

# First step: Persist credential details as environment variables so that they can be referenced later
# TODO: do this

# Now, we can start the whole thing
localstack start & echo "Waiting for localstack to finish setting up"

# pings localhost, waiting for when:
# 1. it responds properly (something is running on the port)
# 2. all of the services are "running" (aka none of the services are "starting")
function healthcheck {
    while :; do
        status_code=`curl -s -o /dev/null -w "%{http_code}" http://localhost:4566/health?reload`
        if [[ $status_code = 200 ]] ; then
            echo "Service has started, checking health"
            curl localhost:4566/health?reload | python3 healthcheck.py
            if [[ $? = 0 ]] ; then
                echo "Localstack is up and running!"
                break
            fi
        fi
        sleep 1
    done
}

healthcheck

# after confirming the health of localstack, we can create our services
# TODO: write cloudformation template to create SQS queue, lambda, dynamo
# Before we can create the lambda, we need to put our function code that we zipped up ourselves into S3,
# so that the template can read that zip from S3 to create the lambda.

# First, we need to create S3 bucket to put lambda code in
aws s3 mb s3://twitch-song-requests --endpoint-url http://localhost:4566

# Then, we can package and deploy the lambda function to S3
cd ./lambda
./package.sh && ./deploy.sh

cd ..
# Now we can use the cloudformation template to create all of our services
aws cloudformation create-stack --endpoint-url http://localhost:4566 --stack-name song-requests --template-body file://services.json --parameters ParameterKey
