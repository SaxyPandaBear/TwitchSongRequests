Local Scripts
=============

This directory holds the required local scripts to run the project on your
machine.

## Required setup
* Make sure you have [jq](https://stedolan.github.io/jq/) installed

## Bootstrapping credentials
With a JSON file at `scripts/credentials.json` which looks like:
```json
{
  "spotifyClientId": "foo",
  "spotifyClientSecret": "bar",
  "twitchClientId": "baz",
  "twitchClientSecret": "abc"
}
```
Source the `bootstrap.sh` script in order to persist the credentials that
are stored in the JSON file to the local environment variables.
> `source scripts/bootstrap.sh`
