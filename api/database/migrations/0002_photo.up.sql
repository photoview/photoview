CREATE TABLE IF NOT EXISTS photo_url (
  url_id int NOT NULL AUTO_INCREMENT,
  token varchar(256) NOT NULL,
  width int NOT NULL,
  height int NOT NULL,

  PRIMARY KEY (url_id)
);

CREATE TABLE IF NOT EXISTS photo_exif (
  exif_id int NOT NULL AUTO_INCREMENT,
  camera varchar(256),
  maker varchar(256),
  lens varchar(256),
  dateShot timestamp,
  file_size_bytes bigint,
  exposure varchar(256),
  aperature float,
  iso int(6),
  focal_length float,
  flash varchar(256),

  PRIMARY KEY (exif_id)
);

CREATE TABLE IF NOT EXISTS album (
  album_id int NOT NULL AUTO_INCREMENT,
  title varchar(256) NOT NULL,
  parent_album int,
  owner_id int NOT NULL,
  path varchar(512) NOT NULL UNIQUE,

  PRIMARY KEY (album_id),
  FOREIGN KEY (parent_album) REFERENCES album(album_id),
  FOREIGN KEY (owner_id) REFERENCES user(user_id)
);

CREATE TABLE IF NOT EXISTS photo (
  photo_id int NOT NULL AUTO_INCREMENT,
  title varchar(256) NOT NULL,
  path varchar(512) NOT NULL UNIQUE,
  original_url int NOT NULL,
  thumbnail_url int NOT NULL,
  album_id int NOT NULL,
  exif_id int,

  PRIMARY KEY (photo_id),
  FOREIGN KEY (album_id) REFERENCES album(album_id),
  FOREIGN KEY (exif_id) REFERENCES photo_exif(exif_id),
  FOREIGN KEY (original_url) REFERENCES photo_url(url_id),
  FOREIGN KEY (thumbnail_url) REFERENCES photo_url(url_id)
);
