-- Users and authentication
CREATE TABLE IF NOT EXISTS user (
  user_id int NOT NULL AUTO_INCREMENT,
  username varchar(256) NOT NULL UNIQUE,
  password varchar(256),
  root_path varchar(512),
  admin boolean NOT NULL DEFAULT 0,

  PRIMARY KEY (user_id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS access_token (
	token_id int NOT NULL AUTO_INCREMENT,
  user_id int NOT NULL,
	value char(24) NOT NULL UNIQUE,
	expire timestamp NOT NULL,

	PRIMARY KEY (token_id),
  FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS site_info (
  initial_setup boolean NOT NULL DEFAULT TRUE
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Video related
CREATE TABLE IF NOT EXISTS video_metadata (
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
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Media related
CREATE TABLE IF NOT EXISTS album (
  album_id int NOT NULL AUTO_INCREMENT,
  title varchar(256) NOT NULL,
  parent_album int,
  owner_id int NOT NULL,
  path varchar(1024) NOT NULL,
  path_hash varchar(32) NOT NULL UNIQUE,

  PRIMARY KEY (album_id),
  FOREIGN KEY (parent_album) REFERENCES album(album_id) ON DELETE CASCADE,
  FOREIGN KEY (owner_id) REFERENCES user(user_id) ON DELETE CASCADE
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS media_exif (
  exif_id int NOT NULL AUTO_INCREMENT,
  camera varchar(256),
  maker varchar(256),
  lens varchar(256),
  dateShot timestamp NULL,
  exposure varchar(256),
  aperture float,
  iso int(6),
  focal_length float,
  flash varchar(256),
  orientation int(1),
  exposure_program int(1),
  gps_latitude float,
  gps_longitude float,

  PRIMARY KEY (exif_id)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS media (
  media_id int NOT NULL AUTO_INCREMENT,
  title varchar(256) NOT NULL,
  path varchar(1024) NOT NULL,
  path_hash varchar(32) NOT NULL UNIQUE,
  album_id int NOT NULL,
  exif_id int,
  favorite boolean DEFAULT FALSE,
  media_type varchar(64) NOT NULL,
  video_metadata_id int,

  PRIMARY KEY (media_id),
  FOREIGN KEY (album_id) REFERENCES album(album_id) ON DELETE CASCADE,
  FOREIGN KEY (exif_id) REFERENCES media_exif(exif_id) ON DELETE CASCADE,
  FOREIGN KEY (video_metadata_id) REFERENCES video_metadata(metadata_id) ON DELETE CASCADE
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS media_url (
  url_id int NOT NULL AUTO_INCREMENT,
  media_id int NOT NULL,
  media_name varchar(512) NOT NULL,
  width int NOT NULL,
  height int NOT NULL,
  purpose varchar(64) NOT NULL,
  content_type varchar(64) NOT NULL,

  PRIMARY KEY (url_id),
  FOREIGN KEY (media_id) REFERENCES media(media_id) ON DELETE CASCADE
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Public shares
CREATE TABLE IF NOT EXISTS share_token (
  token_id int AUTO_INCREMENT,
  value char(24) NOT NULL UNIQUE,
  owner_id int NOT NULL,
  expire timestamp NULL DEFAULT NULL,
  password varchar(256),
  album_id int,
  media_id int,

  PRIMARY KEY (token_id)
  -- CHECK (album_id IS NOT NULL OR media_id IS NOT NULL)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
