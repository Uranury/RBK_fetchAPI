package repositories

import (
	"github.com/jmoiron/sqlx"
)

type SteamRepository interface {
	SaveRequestHistory(userID string, endpoint string) error
}

type steamRepository struct {
	db *sqlx.DB
}

func NewSteamRepository(db *sqlx.DB) SteamRepository {
	return &steamRepository{db: db}
}

func (r *steamRepository) SaveRequestHistory(userID string, endpoint string) error {
	// TODO: implement DB insert here
	return nil
}
