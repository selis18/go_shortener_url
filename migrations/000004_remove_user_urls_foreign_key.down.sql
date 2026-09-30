ALTER TABLE user_urls
    ADD CONSTRAINT user_urls_short_url_fkey
    FOREIGN KEY (short_url) REFERENCES short_urls(short_url);
