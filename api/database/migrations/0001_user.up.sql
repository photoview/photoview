CREATE TABLE IF NOT EXISTS user (
  user_id int NOT NULL AUTO_INCREMENT,
  username varchar(256) NOT NULL UNIQUE,
  password varchar(256) NOT NULL,
  root_path varchar(512),
  admin boolean NOT NULL DEFAULT 0,

  PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS access_token (
	token_id int NOT NULL AUTO_INCREMENT,
  user_id int NOT NULL,
	value char(24) NOT NULL UNIQUE,
	expire timestamp NOT NULL,

	PRIMARY KEY (token_id),
  FOREIGN KEY (user_id) REFERENCES user(user_id)
);
