#!/bin/bash

localstack & echo "Waiting for localstack to finish setting up"

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
