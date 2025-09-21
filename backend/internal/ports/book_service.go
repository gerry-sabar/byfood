package ports

import (
	"context"

	"github.com/gerry-sabar/byfood/internal/domain"
)

type BookService interface {
	ListBooks(ctx context.Context) ([]domain.Book, error)
	GetBook(ctx context.Context, id int64) (*domain.Book, error)
	CreateBook(ctx context.Context, in CreateBookInput) (*domain.Book, error)
	UpdateBook(ctx context.Context, id int64, in UpdateBookInput) (*domain.Book, error)
	DeleteBook(ctx context.Context, id int64) error
}

// CreateBookInput for POST /books.
// swagger:model CreateBookInput
type CreateBookInput struct {
	Title           string  `json:"title"`
	Author          string  `json:"author"`
	ISBN            string  `json:"isbn"`
	Price           float64 `json:"price"`
	PublicationYear int     `json:"publication_year"`
}

// UpdateBookInput for PUT /books/{id}.
// swagger:model UpdateBookInput
type UpdateBookInput struct {
	Title           *string  `json:"title"`
	Author          *string  `json:"author"`
	ISBN            *string  `json:"isbn"`
	Price           *float64 `json:"price"`
	PublicationYear *int     `json:"publication_year"`
}

// ErrorResponse matches your httpError shape.
// swagger:model ErrorResponse
type ErrorResponse struct {
	Error string `json:"error" example:"not found"`
}
