-- Rename team media columns: logo_path/banner_path -> logo_key/banner_key
-- They now store only the media identifier (filename), not a full path/url.
ALTER TABLE al_team RENAME COLUMN logo_path TO logo_key;
ALTER TABLE al_team RENAME COLUMN banner_path TO banner_key;
