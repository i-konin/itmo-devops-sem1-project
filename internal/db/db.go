package db

import (
	"database/sql"
	"fmt"
	"project_sem/internal/models"

	_ "github.com/lib/pq"
)

type PGConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type PG struct {
	DB *sql.DB
}

func NewPostgres(cfg PGConfig) (*PG, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PG{DB: db}, nil
}

func (p *PG) Close() {
	_ = p.DB.Close()
}

func InsertPrices(pg *PG, prices []models.Price) (int, error) {
	tx, err := pg.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO prices (id, name, category, price, create_date) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, p := range prices {
		_, err := stmt.Exec(p.ID, p.Name, p.Category, p.Price, p.CreatedAt)
		if err != nil {
			return 0, err
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return count, nil
}

func GetAllPrices(pg *PG) ([]models.Price, error) {
	rows, err := pg.DB.Query(`SELECT id, name, category, price, create_date FROM prices`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []models.Price
	for rows.Next() {
		var p models.Price
		if err := rows.Scan(&p.ID, &p.Name, &p.Category, &p.Price, &p.CreatedAt); err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

func SelectTotalPrice(pg *PG) (float64, error) {
	var sum float64
	err := pg.DB.QueryRow(`SELECT COALESCE(SUM(price), 0) FROM prices`).Scan(&sum)
	return sum, err
}

func SelectTotalCategories(pg *PG) (int, error) {
	var count int
	err := pg.DB.QueryRow(`SELECT COUNT(DISTINCT category) FROM prices`).Scan(&count)
	return count, err
}