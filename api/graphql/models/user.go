package models

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	Model
	Username string  `gorm:"unique,size:128"`
	Password *string `gorm:"size:256`
	RootPath string  `gorm:"size:512`
	Admin    bool    `gorm:"default:false"`
}

// func (u *User) ID() int {
// 	return u.UserID
// }

type AccessToken struct {
	Model
	UserID int
	User   User   `gorm:"constraint:OnDelete:CASCADE;"`
	Value  string `gorm:"size:24`
	Expire time.Time
}

var ErrorInvalidUserCredentials = errors.New("invalid credentials")

// func NewUserFromRow(row *sql.Row) (*User, error) {
// 	user := User{}

// 	if err := row.Scan(&user.UserID, &user.Username, &user.Password, &user.RootPath, &user.Admin); err != nil {
// 		return nil, errors.Wrap(err, "failed to scan user from database")
// 	}

// 	return &user, nil
// }

// func NewUsersFromRows(rows *sql.Rows) ([]*User, error) {
// 	users := make([]*User, 0)

// 	for rows.Next() {
// 		var user User
// 		if err := rows.Scan(&user.UserID, &user.Username, &user.Password, &user.RootPath, &user.Admin); err != nil {
// 			return nil, errors.Wrap(err, "failed to scan users from database")
// 		}
// 		users = append(users, &user)
// 	}

// 	rows.Close()

// 	return users, nil
// }

func AuthorizeUser(db *gorm.DB, username string, password string) (*User, error) {
	// row := database.QueryRow("SELECT * FROM user WHERE username = ?", username)

	var user User

	result := db.Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrorInvalidUserCredentials
		}
		return nil, errors.Wrap(result.Error, "failed to get user by username when authorizing")
	}

	if user.Password == nil {
		return nil, errors.New("user does not have a password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, ErrorInvalidUserCredentials
		} else {
			return nil, errors.Wrap(err, "compare user password hash")
		}
	}

	return &user, nil
}

var ErrorInvalidRootPath = errors.New("invalid root path")

func ValidRootPath(rootPath string) bool {
	_, err := os.Stat(rootPath)
	if err != nil {
		log.Printf("Warn: invalid root path: '%s'\n%s\n", rootPath, err)
		return false
	}

	return true
}

func RegisterUser(db *gorm.DB, username string, password *string, rootPath string, admin bool) (*User, error) {
	if !ValidRootPath(rootPath) {
		return nil, ErrorInvalidRootPath
	}

	user := User{
		Username: username,
		RootPath: rootPath,
		Admin:    admin,
	}

	if password != nil {
		hashedPassBytes, err := bcrypt.GenerateFromPassword([]byte(*password), 12)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash password")
		}
		hashedPass := string(hashedPassBytes)

		user.Password = &hashedPass
	}

	result := db.Create(&user)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "insert new user with password into database")
	}

	return &user, nil
}

func (user *User) GenerateAccessToken(db *gorm.DB) (*AccessToken, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return nil, errors.New(fmt.Sprintf("Could not generate token: %s\n", err.Error()))
	}
	const CHARACTERS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for i, b := range bytes {
		bytes[i] = CHARACTERS[b%byte(len(CHARACTERS))]
	}

	token_value := string(bytes)
	expire := time.Now().Add(14 * 24 * time.Hour)
	// expireString := expire.UTC().Format("2006-01-02 15:04:05")

	// if _, err := database.Exec("INSERT INTO access_token (value, expire, user_id) VALUES (?, ?, ?)", token_value, expireString, user.UserID); err != nil {
	// 	return nil, err
	// }

	token := AccessToken{
		UserID: user.ID,
		Value:  token_value,
		Expire: expire,
	}

	result := db.Create(&token)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "saving access token to database")
	}

	return &token, nil
}

func VerifyTokenAndGetUser(db *gorm.DB, token string) (*User, error) {

	now := time.Now().UTC().Format("2006-01-02 15:04:05")

	// row := database.QueryRow("SELECT (user_id) FROM access_token WHERE expire > ? AND value = ?", now, token)

	var accessToken AccessToken
	result := db.Where("expire > ? AND value = ?", now, token).First(&accessToken)
	if result.Error != nil {
		return nil, result.Error
	}

	// var userId string

	// if err := row.Scan(&userId); err != nil {
	// 	log.Println(err.Error())
	// 	return nil, err
	// }

	// row = db.QueryRow("SELECT * FROM user WHERE user_id = ?", userId)
	// user, err := NewUserFromRow(row)
	// if err != nil {
	// 	return nil, err
	// }

	var user User
	result = db.First(&user, accessToken.ID)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}
