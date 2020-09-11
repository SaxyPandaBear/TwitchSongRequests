#!/bin/bash

# This assumes that we are already in the pipenv shell!

# First step: make sure that we have all of the required credentials stored in our environment.
# They are needed to run all of the services (the lambda, the main server code, etc).
# We expect the following: 
#   TWITCH_CLIENT_ID, TWITCH_CLIENT_SECRET, SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET
if [[ -z "$TWITCH_CLIENT_ID" ]] ; then
    echo "Missing TWITCH_CLIENT_ID env var"
    exit 1
fi
if [[ -z "$TWITCH_CLIENT_SECRET" ]] ; then
    echo "Missing TWITCH_CLIENT_SECRET env var"
    exit 1
fi
if [[ -z "$SPOTIFY_CLIENT_ID" ]] ; then
    echo "Missing SPOTIFY_CLIENT_ID env var"
    exit 1
fi
if [[ -z "$SPOTIFY_CLIENT_SECRET" ]] ; then
    echo "Missing SPOTIFY_CLIENT_SECRET env var"
    exit 1
fi

echo "Credentials are set. Starting up localstack.."

localstack start &

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
echo "Localstack is running!"

# after confirming the health of localstack, we can create our services
# TODO: write cloudformation template to create SQS queue, lambda, dynamo
# Before we can create the lambda, we need to put our function code that we zipped up ourselves into S3,
# so that the template can read that zip from S3 to create the lambda.

# First, we need to create S3 bucket to put lambda code in
echo "Making bucket at s3://twitch-song-requests"
aws s3 mb s3://twitch-song-requests --endpoint-url http://localhost:4566

# Then, we can package and deploy the lambda function to S3
cd ./lambda
./build.sh && ./deploy.sh

cd ..
# Now we can use the cloudformation template to create all of our services.
# We can read client credentials from the environment because they are necessary before we can do any of this work.
echo "Creating CloudFormation stack..."
aws cloudformation create-stack \
    --endpoint-url http://localhost:4566 \
    --stack-name song-requests \
    --template-body file://services.json \
    --parameters ParameterKey
