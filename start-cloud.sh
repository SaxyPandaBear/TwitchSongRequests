#!/bin/bash

# This assumes that we are already in the pipenv shell!

# First step: make sure that we have all of the required credentials stored in our environment.
# They are needed to run all of the services (the lambda, the main server code, etc).
# We expect the following: 
#   TWITCH_CLIENT_ID, TWITCH_CLIENT_SECRET, SPOTIFY_CLIENT_ID, SPOTIFY_CLIENT_SECRET
echo "Checking environment for required credentials"
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

# For some reason, we can't run `localstack start` here
echo "Credentials are set properly. Moving on..."

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

echo "Waiting for localstack to be up and running..."
healthcheck
echo "Localstack is running!"

# For the other services, we are going to inject an environment variable
# to let them know to use localstack rather than live AWS
export LOCALSTACK="localstack" # the value isn't that important. we just want the key to exist

# after confirming the health of localstack, we can create our services
# Before we can create the lambda, we need to put our function code that we zipped up ourselves into S3,
# so that the template can read that zip from S3 to create the lambda.

# First, we need to create S3 bucket to put lambda code in
echo "Making bucket at s3://twitch-song-requests"
aws s3 mb s3://twitch-song-requests --endpoint-url http://localhost:4566 --region us-east-1

# Then, we can package and deploy the lambda function to S3
echo "Packaging lambda code and writing the zip file to the S3 bucket"
cd ./queue-song-lambda
./deploy.sh

cd ..
# Now we can use the cloudformation template to create all of our services.
# We can read client credentials from the environment because they are necessary before we can do any of this work.
echo "Creating CloudFormation stack..."
aws cloudformation create-stack \
    --region us-east-1 \
    --endpoint-url http://localhost:4566 \
    --stack-name song-requests \
    --template-body file://services.json \
    --parameters ParameterKey=TwitchClientId,ParameterValue="$TWITCH_CLIENT_ID" \
                 ParameterKey=TwitchClientSecret,ParameterValue="$TWITCH_CLIENT_SECRET" \
                 ParameterKey=SpotifyClientId,ParameterValue="$SPOTIFY_CLIENT_ID" \
                 ParameterKey=SpotifyClientSecret,ParameterValue="$SPOTIFY_CLIENT_SECRET"

# Listing queues here so that we can inject the queue url from this into the environment
# for the server code
echo "Listing SQS queues in the environment"
aws sqs list-queues --endpoint-url http://localhost:4566

echo "Finished standing up infrastructure! Refer to the workshop.sh to get started on using it."
