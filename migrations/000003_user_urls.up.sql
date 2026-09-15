CREATE TABLE user_urls (
    user_id TEXT NOT NULL,
    short_url TEXT NOT NULL REFERENCES short_urls(short_url),
    PRIMARY KEY (user_id, short_url)
);
