ALTER TABLE short_urls ADD COLUMN is_deleted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE short_urls ADD COLUMN user_id TEXT NOT NULL DEFAULT '';


UPDATE short_urls s SET user_id = u.user_id
FROM (SELECT short_url, MIN(user_id) AS user_id FROM user_urls
      GROUP BY short_url HAVING COUNT(*) = 1) u
WHERE s.short_url = u.short_url;
