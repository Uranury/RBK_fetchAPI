package repositories

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type SteamRepository interface {
	SaveRequestHistory(endpoint string, params map[string]interface{}, success bool, errorMessage string, duration time.Duration) error
}

type steamRepository struct {
	db *sqlx.DB
}

func NewSteamRepository(db *sqlx.DB) SteamRepository {
	return &steamRepository{db: db}
}

func (r *steamRepository) SaveRequestHistory(endpoint string, params map[string]interface{}, success bool, errorMessage string, duration time.Duration) error {
	jsonParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal params: %w", err)
	}
	_, err = r.db.Exec(
		`INSERT INTO request_history (endpoint, params, success, error_message, response_time_ms) VALUES ($1, $2, $3, $4, $5)`,
		endpoint, jsonParams, success, errorMessage, duration,
	)
	if err != nil {
		return err
	}
	return nil
}
