package ports

import (
	"context"

	"github.com/gerry-sabar/byfood/internal/domain"
)

type BookRepository interface {
	List(ctx context.Context) ([]domain.Book, error)
	GetByID(ctx context.Context, id int64) (*domain.Book, error)
	Create(ctx context.Context, b *domain.Book) (int64, error)
	Update(ctx context.Context, b *domain.Book) error
	Delete(ctx context.Context, id int64) error
}
