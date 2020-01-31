CREATE TABLE IF NOT EXISTS users (
  user_id int NOT NULL AUTO_INCREMENT,
  username varchar(255) NOT NULL UNIQUE,
  password varchar(255) NOT NULL,
  root_path varchar(512),
  admin boolean NOT NULL DEFAULT 0,

  PRIMARY KEY (user_id)
);
