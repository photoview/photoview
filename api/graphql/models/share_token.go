package models

import (
	"database/sql"
	"time"
)

type ShareToken struct {
	TokenID  int
	Value    string
	OwnerID  int
	Expire   *time.Time
	Password *string
	AlbumID  *int
	PhotoID  *int
}

func (share *ShareToken) Token() string {
	return share.Value
}

func (share *ShareToken) ID() int {
	return share.TokenID
}

func NewShareTokenFromRow(row *sql.Row) (*ShareToken, error) {
	token := ShareToken{}

	if err := row.Scan(&token.TokenID, &token.Value, &token.OwnerID, &token.Expire, &token.Password, &token.AlbumID, &token.PhotoID); err != nil {
		return nil, err
	}

	return &token, nil
}

func NewShareTokensFromRows(rows *sql.Rows) ([]*ShareToken, error) {
	tokens := make([]*ShareToken, 0)

	for rows.Next() {
		var token ShareToken
		if err := rows.Scan(&token.TokenID, &token.Value, &token.OwnerID, &token.Expire, &token.Password, &token.AlbumID, &token.PhotoID); err != nil {
			return nil, err
		}
		tokens = append(tokens, &token)
	}

	rows.Close()

	return tokens, nil
}
