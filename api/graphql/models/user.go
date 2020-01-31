package models

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID   int
	Username string
	Password string
	RootPath string
	Admin    bool
}

func (u *User) ID() string {
	return strconv.Itoa(u.UserID)
}

type AccessToken struct {
	Value  string
	Expire time.Time
}

var ErrorInvalidUserCredentials = errors.New("invalid credentials")

func NewUserFromRow(row *sql.Row) (*User, error) {
	user := User{}

	if err := row.Scan(&user.UserID, &user.Username, &user.Password, &user.RootPath, &user.Admin); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorInvalidUserCredentials
		} else {
			return nil, err
		}
	}

	return &user, nil
}

func NewUsersFromRows(rows *sql.Rows) ([]*User, error) {
	users := make([]*User, 0)

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserID, &user.Username, &user.Password, &user.RootPath, &user.Admin); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func AuthorizeUser(database *sql.DB, username string, password string) (*User, error) {
	row := database.QueryRow("SELECT * FROM users WHERE username = ?", username)

	user, err := NewUserFromRow(row)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, ErrorInvalidUserCredentials
		} else {
			return nil, err
		}
	}

	return user, nil
}

func RegisterUser(database *sql.DB, username string, password string, rootPath string) (*User, error) {
	hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}
	hashedPass := string(hashedPassBytes)

	if _, err := database.Exec("INSERT INTO users (username, password, root_path) VALUES (?, ?, ?)", username, hashedPass, rootPath); err != nil {
		return nil, err
	}

	row := database.QueryRow("SELECT * FROM users WHERE username = ?", username)
	if row == nil {
		return nil, ErrorInvalidUserCredentials
	}

	user, err := NewUserFromRow(row)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (user *User) GenerateAccessToken(database *sql.DB) (*AccessToken, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not generate token: %s\n", err.Error()))
	}
	const CHARACTERS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	for i, b := range bytes {
		bytes[i] = CHARACTERS[b%byte(len(CHARACTERS))]
	}

	token_value := string(bytes)
	expire := time.Now().Add(14 * 24 * time.Hour)
	expireString := expire.UTC().Format("2006-01-02 15:04:05")

	if _, err := database.Exec("INSERT INTO access_tokens (value, expire, user_id) VALUES (?, ?, ?)", token_value, expireString, user.UserID); err != nil {
		return nil, err
	}

	token := AccessToken{
		Value:  token_value,
		Expire: expire,
	}

	return &token, nil
}

func VerifyTokenAndGetUser(database *sql.DB, token string) (*User, error) {

	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	row := database.QueryRow("SELECT (user_id) FROM access_tokens WHERE expire > ? AND value = ?", now, token)

	var userId string

	if err := row.Scan(&userId); err != nil {
		log.Println(err.Error())
		return nil, err
	}

	fmt.Printf("Userid: %s\n", userId)

	row = database.QueryRow("SELECT * FROM users WHERE user_id = ?", userId)
	user, err := NewUserFromRow(row)
	if err != nil {
		return nil, err
	}

	return user, nil
}
