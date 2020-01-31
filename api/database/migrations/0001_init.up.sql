CREATE TABLE IF NOT EXISTS users (
  user_id int NOT NULL AUTO_INCREMENT,
  username varchar(255) NOT NULL,
  password varchar(255) NOT NULL,
  root_path varchar(512) NOT NULL,
  admin boolean,

  PRIMARY KEY (user_id)
);
