package main

import (
	"database/sql"
)

type promotionModel struct {
	ID             string  `json:"id"`
	Price          float64 `json:"price"`
	ExpirationDate string  `json:"expiration_date"`
}

func (p *promotionModel) getPromotion(db *sql.DB) error {
	return db.QueryRow("SELECT price, expiration_date FROM promotions WHERE id=$1",
		p.ID).Scan(&p.Price, &p.ExpirationDate)
}

func (p *promotionModel) updatePromotion(db *sql.DB) error {
	_, err :=
		db.Exec("UPDATE promotions SET price=$2, expiration_date=$3 WHERE id=$1",
			p.ExpirationDate, p.Price, p.ID)

	return err
}

func (p *promotionModel) deletePromotion(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM promotions WHERE id=$1", p.ID)

	return err
}

func (p *promotionModel) createPromotion(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO promotions(id, price, expiration_date) VALUES($1, $2, $3) RETURNING id",
		p.Price, p.ExpirationDate).Scan(&p.ID)

	if err != nil {
		return err
	}

	return nil
}

func getPromotions(db *sql.DB) ([]promotionModel, error) {
	rows, err := db.Query(
		"SELECT id, price, expiration_date FROM promotions LIMIT 50")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var promotions []promotionModel

	for rows.Next() {
		var p promotionModel
		if err := rows.Scan(&p.ID, &p.Price, &p.ExpirationDate); err != nil {
			return nil, err
		}
		promotions = append(promotions, p)
	}

	return promotions, nil
}
