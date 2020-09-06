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
