package domain

import "time"

// Book represents the API response shape for a book.
// swagger:model Book
type Book struct {
	ID              int64     `db:"id" json:"id"`
	Title           string    `db:"title" json:"title"`
	Author          string    `db:"author" json:"author"`
	ISBN            string    `db:"isbn" json:"isbn"`
	Price           float64   `db:"price" json:"price"`
	PublicationYear int       `db:"publication_year" json:"publication_year"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}
