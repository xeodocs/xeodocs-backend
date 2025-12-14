package configurations

import (
	"database/sql"
)

type Configuration struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]Configuration, error) {
	rows, err := r.db.Query("SELECT key, value FROM configurations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []Configuration
	for rows.Next() {
		var c Configuration
		if err := rows.Scan(&c.Key, &c.Value); err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, nil
}

func (r *Repository) GetByKey(key string) (*Configuration, error) {
	var c Configuration
	err := r.db.QueryRow("SELECT key, value FROM configurations WHERE key = $1", key).Scan(&c.Key, &c.Value)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) Upsert(c *Configuration) error {
	_, err := r.db.Exec(`
		INSERT INTO configurations (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2
	`, c.Key, c.Value)
	return err
}

func (r *Repository) Delete(key string) error {
	_, err := r.db.Exec("DELETE FROM configurations WHERE key = $1", key)
	return err
}
