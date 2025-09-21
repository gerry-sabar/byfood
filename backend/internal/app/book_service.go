package app

import (
	"context"
	"errors"
	"time"

	"github.com/gerry-sabar/byfood/internal/domain"
	"github.com/gerry-sabar/byfood/internal/ports"
)

type bookService struct {
	repo ports.BookRepository
}

func NewBookService(repo ports.BookRepository) ports.BookService {
	return &bookService{repo: repo}
}

func (s *bookService) ListBooks(ctx context.Context) ([]domain.Book, error) {
	return s.repo.List(ctx)
}

func (s *bookService) GetBook(ctx context.Context, id int64) (*domain.Book, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *bookService) CreateBook(ctx context.Context, in ports.CreateBookInput) (*domain.Book, error) {
	inNorm, err := validateAndNormalizeCreate(in)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	book := &domain.Book{
		Title:           inNorm.Title,
		Author:          inNorm.Author,
		ISBN:            inNorm.ISBN, // normalized
		PublicationYear: inNorm.PublicationYear,
		Price:           inNorm.Price,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	id, err := s.repo.Create(ctx, book)
	if err != nil {
		return nil, err
	}
	book.ID = id
	return book, nil
}

func (s *bookService) UpdateBook(ctx context.Context, id int64, in ports.UpdateBookInput) (*domain.Book, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("book not found")
	}

	inNorm, err := validateAndNormalizeUpdate(in)
	if err != nil {
		return nil, err
	}

	if inNorm.Title != nil {
		existing.Title = *inNorm.Title
	}
	if inNorm.Author != nil {
		existing.Author = *inNorm.Author
	}
	if inNorm.ISBN != nil {
		existing.ISBN = *inNorm.ISBN // normalized
	}
	if inNorm.PublicationYear != nil {
		existing.PublicationYear = *inNorm.PublicationYear
	}
	if inNorm.Price != nil {
		existing.Price = *inNorm.Price
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *bookService) DeleteBook(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
