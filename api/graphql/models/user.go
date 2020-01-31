package models

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	User_id   int
	Username  string
	Password  string
	Root_path *string
	Admin     bool
}

var UserInvalidCredentialsError = errors.New("invalid credentials")

func NewUserFromRow(row *sql.Row) (*User, error) {
	user := User{}

	row.Scan(&user.User_id, &user.Username, &user.Password, &user.Root_path, &user.Admin)

	return &user, nil
}

func AuthorizeUser(database *sql.DB, username string, password string) (*User, error) {
	row := database.QueryRow("SELECT * FROM users WHERE username = ?", username)
	if row == nil {
		return nil, UserInvalidCredentialsError
	}

	user, err := NewUserFromRow(row)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, UserInvalidCredentialsError
		} else {
			return nil, err
		}
	}

	return user, nil
}

func RegisterUser(database *sql.DB, username string, password string) (*User, error) {
	hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}
	hashedPass := string(hashedPassBytes)

	if _, err := database.Query("INSERT INTO users (username, password) VALUES (?, ?)", username, hashedPass); err != nil {
		return nil, err
	}

	row := database.QueryRow("SELECT * FROM users WHERE username = ?", username)
	if row == nil {
		return nil, UserInvalidCredentialsError
	}

	user, err := NewUserFromRow(row)
	if err != nil {
		return nil, err
	}

	return user, nil
}
