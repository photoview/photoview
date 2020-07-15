ALTER TABLE photo RENAME TO media;
ALTER TABLE photo_url RENAME TO media_url;
ALTER TABLE photo_exif RENAME TO media_exif;

ALTER TABLE media RENAME COLUMN photo_id TO media_id;

ALTER TABLE media_url
  RENAME COLUMN photo_id TO media_id,
  RENAME COLUMN photo_name TO media_name;

ALTER TABLE share_token RENAME COLUMN photo_id TO media_id;

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
