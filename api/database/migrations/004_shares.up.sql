CREATE TABLE IF NOT EXISTS share_token (
  token_id int(11) AUTO_INCREMENT,
  value char(24) NOT NULL UNIQUE,
  owner_id int(11) NOT NULL,
  expire timestamp NULL DEFAULT NULL,
  password varchar(256),
  album_id int(11),
  photo_id int(11),

  PRIMARY KEY (token_id)
  -- CHECK (album_id IS NOT NULL OR photo_id IS NOT NULL)
);
