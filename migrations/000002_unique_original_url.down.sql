ALTER TABLE short_urls DROP CONSTRAINT IF EXISTS short_urls_original_url_key;
DROP INDEX IF EXISTS short_urls_original_url_key;
