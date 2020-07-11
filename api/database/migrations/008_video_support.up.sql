
ALTER TABLE photo RENAME TO media;
ALTER TABLE photo_url RENAME TO media_url;
ALTER TABLE photo_exif RENAME TO media_exif;

ALTER TABLE media CHANGE COLUMN photo_id media_id int NOT NULL AUTO_INCREMENT;
ALTER TABLE media_url CHANGE COLUMN photo_id media_id int NOT NULL;
ALTER TABLE media_url CHANGE COLUMN photo_name media_name varchar(512) NOT NULL;
ALTER TABLE share_token CHANGE COLUMN photo_id media_id int;

ALTER TABLE media ADD COLUMN media_type varchar(64) NOT NULL DEFAULT "photo";
