CREATE DATABASE testdb;
USE testdb;

CREATE TABLE users(
    id TEXT PRIMARY KEY, 
    twitch_access TEXT, 
    twitch_refresh TEXT, 
    spotify_access TEXT, 
    spotify_refresh TEXT, 
    last_updated DATE, 
    spotify_expiry DATE, 
    subscribed BOOLEAN, 
    subscription_id TEXT, 
    email TEXT
);

INSERT INTO users(id, twitch_access, twitch_refresh, spotify_access, spotify_refresh, subscribed, subscription_id, email) 
VALUES ('12345', 'a', 'b', 'c', 'd', true, 'abc-123', 'foo@bar');
