
ALTER TABLE media RENAME TO photo;
ALTER TABLE media_url RENAME TO photo_url;
ALTER TABLE media_exif RENAME TO photo_exif;

ALTER TABLE photo CHANGE COLUMN media_id photo_id int NOT NULL AUTO_INCREMENT;
ALTER TABLE photo_url CHANGE COLUMN media_id photo_id int NOT NULL;
ALTER TABLE photo_url CHANGE COLUMN media_name photo_name varchar(512) NOT NULL;
ALTER TABLE share_token CHANGE COLUMN media_id photo_id int;

ALTER TABLE photo DROP COLUMN media_type;

ALTER TABLE photo
  DROP FOREIGN KEY photo_ibfk_3,
  DROP COLUMN video_metadata_id;

DROP TABLE video_metadata;