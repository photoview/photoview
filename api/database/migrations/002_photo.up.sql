CREATE TABLE IF NOT EXISTS photo_exif (
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

  PRIMARY KEY (exif_id)
);

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
);

CREATE TABLE IF NOT EXISTS photo (
  photo_id int NOT NULL AUTO_INCREMENT,
  title varchar(256) NOT NULL,
  path varchar(1024) NOT NULL,
  path_hash varchar(32) NOT NULL UNIQUE,
  album_id int NOT NULL,
  exif_id int,

  PRIMARY KEY (photo_id),
  FOREIGN KEY (album_id) REFERENCES album(album_id) ON DELETE CASCADE,
  FOREIGN KEY (exif_id) REFERENCES photo_exif(exif_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS photo_url (
  url_id int NOT NULL AUTO_INCREMENT,
  photo_id int NOT NULL,
  photo_name varchar(512) NOT NULL,
  width int NOT NULL,
  height int NOT NULL,
  purpose varchar(64) NOT NULL,
  content_type varchar(64) NOT NULL,

  PRIMARY KEY (url_id),
  FOREIGN KEY (photo_id) REFERENCES photo(photo_id) ON DELETE CASCADE
);