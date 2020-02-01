CREATE TABLE IF NOT EXISTS photo_url (
  url_id int NOT NULL AUTO_INCREMENT,
  path varchar(256) NUT NULL,
  width int NOT NULL,
  height int NOT NULL,

  PRIMARY KEY (url_id)
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
  -- original_url int NOT NULL,
  -- thumbnail_url int NOT NULL,
  album_id int NOT NULL,
  -- exif_id int NOT NULL,

  PRIMARY KEY (photo_id),
  FOREIGN KEY (original_url) REFERENCES photo_url(url_id),
  FOREIGN KEY (thumbnail_url) REFERENCES photo_url(url_id)
);