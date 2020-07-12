
ALTER TABLE photo RENAME TO media;
ALTER TABLE photo_url RENAME TO media_url;
ALTER TABLE photo_exif RENAME TO media_exif;

ALTER TABLE media CHANGE COLUMN photo_id media_id int NOT NULL AUTO_INCREMENT;
ALTER TABLE media_url
  CHANGE COLUMN photo_id media_id int NOT NULL,
  CHANGE COLUMN photo_name media_name varchar(512) NOT NULL;
ALTER TABLE share_token CHANGE COLUMN photo_id media_id int;

CREATE TABLE video_metadata (
  metadata_id int NOT NULL AUTO_INCREMENT,

  width int(6) NOT NULL,
  height int(6) NOT NULL,
  duration double NOT NULL,
  codec varchar(128),
  framerate double,
  bitrate int(24),
  color_profile varchar(128),
  audio varchar(128),

  PRIMARY KEY (metadata_id)
);

ALTER TABLE media
  ADD COLUMN media_type varchar(64) NOT NULL DEFAULT "photo",
  ADD COLUMN video_metadata_id int,
  ADD FOREIGN KEY (video_metadata_id) REFERENCES video_metadata(metadata_id);
