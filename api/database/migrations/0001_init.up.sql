CREATE TABLE IF NOT EXISTS users (
  user_id int NOT NULL AUTO_INCREMENT,
  username varchar(255) NOT NULL UNIQUE,
  password varchar(255) NOT NULL,
  root_path varchar(512),
  admin boolean NOT NULL DEFAULT 0,

  PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS access_tokens (
	token_id int NOT NULL AUTO_INCREMENT,
  user_id int NOT NULL,
	value char(24) NOT NULL UNIQUE,
	expire timestamp NOT NULL,

	PRIMARY KEY (token_id),
  FOREIGN KEY (user_id) REFERENCES users(user_id)
);
