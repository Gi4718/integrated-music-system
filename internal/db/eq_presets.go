package db

import (
	"database/sql"
	"time"
)

type EQPreset struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Gains     string    `json:"gains"` // JSON array of float64
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetEQPresets(userID int) ([]EQPreset, error) {
	rows, err := dbConn.Query("SELECT id, user_id, name, gains, created_at, updated_at FROM eq_presets WHERE user_id = ? ORDER BY name", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var presets []EQPreset
	for rows.Next() {
		var p EQPreset
		var createdAt, updatedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Gains, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			p.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			p.UpdatedAt = updatedAt.Time
		}
		presets = append(presets, p)
	}
	return presets, nil
}

func SaveEQPreset(userID int, name, gains string) (int, error) {
	now := time.Now()
	res, err := dbConn.Exec(
		`INSERT INTO eq_presets (user_id, name, gains, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(user_id, name) DO UPDATE SET gains = excluded.gains, updated_at = excluded.updated_at`,
		userID, name, gains, now, now,
	)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func DeleteEQPreset(userID, presetID int) error {
	_, err := dbConn.Exec("DELETE FROM eq_presets WHERE id = ? AND user_id = ?", presetID, userID)
	return err
}
