Spotify Queueing function
=========================

This module holds the code for the lambda function part of the architecture. The 
purpose of this lambda is to ingest messages from an SQS queue 
(this is set at an event source for the lambda), and to parse the messages in order 
to read individual queue requests for a Spotify player.

## How it works
1. Lambda receives event from SQS
2. Read message attribute to get the Twitch Channel ID
3. Lookup record in DynamoDB with the channel ID
4. Queue song in Spotify via the Player API
5. (maybe?) Emit an event saying we successfully queued the song, or failed

The lambda shouldn't be running very long. the Dynamo work will definitely be the 
bottleneck here. 

## Developing
Developing code meant for lambda can be tricky. We are going to leverage 
[localstack](https://github.com/localstack/localstack) to run everything, including
this lambda code, so that will make this a little tricky for us, but manageable. 

This code makes HTTP requests using the 
[node-fetch](https://github.com/node-fetch/node-fetch) library, and since that is
an external dependency, we are required to package the lambda ourselves the hard way.

A `package.json` file is provided in the `lambda` module for this. Locally, you need to
`npm install` the dependencies (they will be gitignored) so that they can be zipped 
with the necessary dependencies. 

Check out the instructions in the root [README](../README.md) to see how to get 
localstack started up.

When localstack is up:
```bash
curl http://localhost:4566/health
```
We can start using the local services to create our lambda.

TODO: put instructions for how to package our lambda
