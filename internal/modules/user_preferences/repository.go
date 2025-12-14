package user_preferences

import (
	"database/sql"
)

type UserPreference struct {
	UserID string `json:"userId"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(userID string) ([]UserPreference, error) {
	rows, err := r.db.Query("SELECT user_id, key, value FROM user_preferences WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefs []UserPreference
	for rows.Next() {
		var p UserPreference
		if err := rows.Scan(&p.UserID, &p.Key, &p.Value); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func (r *Repository) GetByKey(userID, key string) (*UserPreference, error) {
	var p UserPreference
	err := r.db.QueryRow("SELECT user_id, key, value FROM user_preferences WHERE user_id = $1 AND key = $2", userID, key).Scan(&p.UserID, &p.Key, &p.Value)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Upsert(p *UserPreference) error {
	_, err := r.db.Exec(`
		INSERT INTO user_preferences (user_id, key, value) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, key) DO UPDATE SET value = $3
	`, p.UserID, p.Key, p.Value)
	return err
}

func (r *Repository) Delete(userID, key string) error {
	_, err := r.db.Exec("DELETE FROM user_preferences WHERE user_id = $1 AND key = $2", userID, key)
	return err
}
