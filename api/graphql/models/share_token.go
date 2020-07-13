package models

import (
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

type ShareToken struct {
	TokenID  int
	Value    string
	OwnerID  int
	Expire   *time.Time
	Password *string
	AlbumID  *int
	MediaID  *int
}

func (share *ShareToken) Token() string {
	return share.Value
}

func (share *ShareToken) ID() int {
	return share.TokenID
}

func NewShareTokenFromRow(row *sql.Row) (*ShareToken, error) {
	token := ShareToken{}

	if err := row.Scan(&token.TokenID, &token.Value, &token.OwnerID, &token.Expire, &token.Password, &token.AlbumID, &token.MediaID); err != nil {
		return nil, errors.Wrap(err, "failed to scan share token from database")
	}

	return &token, nil
}

func NewShareTokensFromRows(rows *sql.Rows) ([]*ShareToken, error) {
	tokens := make([]*ShareToken, 0)

	for rows.Next() {
		var token ShareToken
		if err := rows.Scan(&token.TokenID, &token.Value, &token.OwnerID, &token.Expire, &token.Password, &token.AlbumID, &token.MediaID); err != nil {
			return nil, errors.Wrap(err, "failed to scan share tokens from database")
		}
		tokens = append(tokens, &token)
	}

	rows.Close()

	return tokens, nil
}
