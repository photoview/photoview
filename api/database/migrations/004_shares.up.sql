CREATE TABLE IF NOT EXISTS share_token (
  token_id int AUTO_INCREMENT,
  value char(24) NOT NULL UNIQUE,
  owner_id int NOT NULL,
  expire timestamp NULL DEFAULT NULL,
  password varchar(256),
  album_id int,
  photo_id int,

  PRIMARY KEY (token_id)
  -- CHECK (album_id IS NOT NULL OR photo_id IS NOT NULL)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
