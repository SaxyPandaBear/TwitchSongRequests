-- Create tables in the railway Postgres database
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    twitch_access TEXT NULL,
    twitch_refresh TEXT NULL,
    spotify_access TEXT NULL,
    spotify_refresh TEXT NULL,
    last_updated DATE NULL,
    spotify_expiry DATE NULL,
    subscribed BOOLEAN NULL,
    subscription_id TEXT NULL,
    email TEXT NULL
);

CREATE TABLE IF NOT EXISTS preferences (
    id TEXT PRIMARY KEY,
    `explicit` BOOLEAN NULL,
    reward_id TEXT NULL,
    last_updated DATE NULL,
    max_song_length INT NULL
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    created_at DATE NULL,
    success TINYINT NULL,
    broadcaster_id TEXT NULL,
    spotify_track TEXT NULL
);