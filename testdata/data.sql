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

CREATE TABLE preferences(
    id TEXT PRIMARY KEY,
    explicit BOOLEAN,
    reward_id TEXT,
    last_updated DATE, 
    max_song_length INT
);

INSERT INTO preferences(id, explicit, reward_id, max_song_length)
VALUES ('12345', false, "abc-123", 0), ('23456', true, 'bcd-234', 50000);

CREATE TABLE messages(
    id SERIAL PRIMARY KEY,
    created_at DATE,
    success INT,
    broadcaster_id TEXT, 
    spotify_track TEXT
);

INSERT INTO messages(success, broadcaster_id, spotify_track)
VALUES (1, '12345', 'abc'), (0, '23456', ''), (1, '12345', 'bcd');
