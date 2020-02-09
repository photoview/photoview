CREATE TABLE IF NOT EXISTS share_token (
  token_id int AUTO_INCREMENT,
  value char(24) NOT NULL UNIQUE,
  owner_id int NOT NULL,
  expire timestamp,
  password varchar(256) NOT NULL,
  album_id int,
  photo_id int,

  PRIMARY KEY (token_id),
  CHECK (album_id IS NOT NULL OR photo_id IS NOT NULL)
);
