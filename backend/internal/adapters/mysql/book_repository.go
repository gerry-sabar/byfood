package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gerry-sabar/byfood/internal/domain"
	"github.com/gerry-sabar/byfood/internal/logger"
	"github.com/gerry-sabar/byfood/internal/ports"
	"github.com/jmoiron/sqlx"
)

type bookRepository struct {
	db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) ports.BookRepository {
	return &bookRepository{db: db}
}

func (r *bookRepository) List(ctx context.Context) ([]domain.Book, error) {
	var books []domain.Book
	err := r.db.SelectContext(ctx, &books, `
		SELECT id, title, author, isbn, price, publication_year, created_at, updated_at
		FROM books
		ORDER BY id DESC`)

	if err != nil {
		logger.Log.Error("failed to list books", "error", err)
	}
	return books, err
}

func (r *bookRepository) GetByID(ctx context.Context, id int64) (*domain.Book, error) {
	var b domain.Book
	err := r.db.GetContext(ctx, &b, `
		SELECT id, title, author, isbn, price, publication_year, created_at, updated_at
		FROM books WHERE id = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		logger.Log.Error("failed to get book by id", "id", id, "error", err)
	}
	return &b, err
}

func (r *bookRepository) Create(ctx context.Context, b *domain.Book) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO books (title, author, isbn, price, publication_year, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.Title, b.Author, b.ISBN, b.Price, b.PublicationYear, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		logger.Log.Error("failed to create book", "book", b, "error", err)
		return 0, err
	}
	return res.LastInsertId()
}

func (r *bookRepository) Update(ctx context.Context, b *domain.Book) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE books
		SET title = ?, author = ?, isbn = ?, price = ?, publication_year = ?, updated_at = ?
		WHERE id = ?`,
		b.Title, b.Author, b.ISBN, b.Price, b.PublicationYear, b.UpdatedAt, b.ID,
	)
	if err != nil {
		logger.Log.Error("failed to update book", "id", b.ID, "error", err)
	}
	return err
}

func (r *bookRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM books WHERE id = ?`, id)
	if err != nil {
		logger.Log.Error("failed to delete book", "id", id, "error", err)
	}
	return err
}
