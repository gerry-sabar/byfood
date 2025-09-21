package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gerry-sabar/byfood/internal/domain"
	"github.com/gerry-sabar/byfood/internal/ports"
)

// ---- Minimal mock for ports.BookRepository ----

type mockRepo struct {
	ListFn    func(ctx context.Context) ([]domain.Book, error)
	GetByIDFn func(ctx context.Context, id int64) (*domain.Book, error)
	CreateFn  func(ctx context.Context, b *domain.Book) (int64, error)
	UpdateFn  func(ctx context.Context, b *domain.Book) error
	DeleteFn  func(ctx context.Context, id int64) error
}

func (m *mockRepo) List(ctx context.Context) ([]domain.Book, error) { return m.ListFn(ctx) }
func (m *mockRepo) GetByID(ctx context.Context, id int64) (*domain.Book, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *mockRepo) Create(ctx context.Context, b *domain.Book) (int64, error) {
	return m.CreateFn(ctx, b)
}
func (m *mockRepo) Update(ctx context.Context, b *domain.Book) error { return m.UpdateFn(ctx, b) }
func (m *mockRepo) Delete(ctx context.Context, id int64) error       { return m.DeleteFn(ctx, id) }

// ---- Small helpers ----

func f64ptr(v float64) *float64 { return &v }
func strptr(s string) *string   { return &s }
func iptr(i int) *int           { return &i }

// ---- Tests ----

func TestListBooks_OK(t *testing.T) {
	m := &mockRepo{
		ListFn: func(ctx context.Context) ([]domain.Book, error) {
			return []domain.Book{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}}, nil
		},
	}
	svc := NewBookService(m)

	got, err := svc.ListBooks(context.Background())
	if err != nil {
		t.Fatalf("ListBooks err: %v", err)
	}
	if len(got) != 2 || got[0].Title != "A" || got[1].Title != "B" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestGetBook_PassThrough(t *testing.T) {
	m := &mockRepo{
		GetByIDFn: func(ctx context.Context, id int64) (*domain.Book, error) {
			if id != 10 {
				t.Fatalf("expected id 10; got %d", id)
			}
			return &domain.Book{ID: 10, Title: "X"}, nil
		},
	}
	svc := NewBookService(m)

	got, err := svc.GetBook(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetBook err: %v", err)
	}
	if got == nil || got.ID != 10 || got.Title != "X" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestCreateBook_OK(t *testing.T) {
	var captured *domain.Book
	m := &mockRepo{
		CreateFn: func(ctx context.Context, b *domain.Book) (int64, error) {
			captured = b
			// pretend DB assigned id
			return 42, nil
		},
	}
	svc := NewBookService(m)

	// Use already-normalized inputs so the test doesn't depend on normalization internals.
	in := ports.CreateBookInput{
		Title:           "Clean Code",
		Author:          "Robert C. Martin",
		ISBN:            "9780132350884",
		PublicationYear: 2008,
		Price:           33.50,
	}
	start := time.Now().UTC()
	got, err := svc.CreateBook(context.Background(), in)
	if err != nil {
		t.Fatalf("CreateBook err: %v", err)
	}

	// Repo was called with a book:
	if captured == nil {
		t.Fatalf("repo.Create wasn't called")
	}
	// Field mapping
	if captured.Title != in.Title ||
		captured.Author != in.Author ||
		captured.ISBN != in.ISBN ||
		captured.PublicationYear != in.PublicationYear ||
		captured.Price != in.Price {
		t.Fatalf("captured mismatch: %+v", captured)
	}
	// Timestamps are set and sane
	if captured.CreatedAt.IsZero() || captured.UpdatedAt.IsZero() {
		t.Fatalf("timestamps not set: %+v", captured)
	}
	if captured.UpdatedAt.Sub(captured.CreatedAt) < 0 {
		t.Fatalf("updatedAt before createdAt")
	}
	// Returned book contains assigned ID and same fields
	if got.ID != 42 || got.Title != in.Title || got.PublicationYear != in.PublicationYear {
		t.Fatalf("got mismatch: %+v", got)
	}
	// Time should be >= start (coarse sanity)
	if got.CreatedAt.Before(start) {
		t.Fatalf("createdAt too early: %v < %v", got.CreatedAt, start)
	}
}

func TestCreateBook_RepoError(t *testing.T) {
	called := false
	m := &mockRepo{
		CreateFn: func(ctx context.Context, b *domain.Book) (int64, error) {
			called = true
			return 0, errors.New("insert failed")
		},
	}
	svc := NewBookService(m)

	// Use a payload that passes validation so we reach repo.Create
	in := ports.CreateBookInput{
		Title:           "Domain-Driven Design",
		Author:          "Eric Evans",
		ISBN:            "9780321125217", // valid 13-digit ISBN
		PublicationYear: 2003,
		Price:           49.99, // positive
	}

	_, err := svc.CreateBook(context.Background(), in)
	if !called {
		t.Fatalf("repo.Create was not called; validation likely failed before repo")
	}
	if err == nil || err.Error() != "insert failed" {
		t.Fatalf("want insert failed; got %v", err)
	}
}

func TestUpdateBook_NotFound(t *testing.T) {
	m := &mockRepo{
		GetByIDFn: func(ctx context.Context, id int64) (*domain.Book, error) { return nil, nil },
	}
	svc := NewBookService(m)

	_, err := svc.UpdateBook(context.Background(), 9, ports.UpdateBookInput{
		Title: strptr("New"),
	})
	if err == nil || err.Error() != "book not found" {
		t.Fatalf("want 'book not found'; got %v", err)
	}
}

func TestUpdateBook_OK_PartialFields(t *testing.T) {
	orig := &domain.Book{
		ID:              7,
		Title:           "Old",
		Author:          "Someone",
		ISBN:            "111",
		PublicationYear: 1999,
		Price:           10.0,
	}
	var updatedToRepo *domain.Book

	m := &mockRepo{
		GetByIDFn: func(ctx context.Context, id int64) (*domain.Book, error) {
			if id != 7 {
				t.Fatalf("expected id 7; got %d", id)
			}
			// return a copy (simulate DB fetch)
			cp := *orig
			return &cp, nil
		},
		UpdateFn: func(ctx context.Context, b *domain.Book) error {
			updatedToRepo = b
			return nil
		},
	}

	svc := NewBookService(m)

	in := ports.UpdateBookInput{
		Title: strptr("NewTitle"),
		Price: f64ptr(12.34),
		// Author, ISBN, PublicationYear remain nil → unchanged
	}

	before := time.Now().UTC()
	got, err := svc.UpdateBook(context.Background(), 7, in)
	if err != nil {
		t.Fatalf("UpdateBook err: %v", err)
	}

	// Ensure repo.Update received merged entity
	if updatedToRepo == nil {
		t.Fatalf("repo.Update not called")
	}
	if updatedToRepo.Title != "NewTitle" || updatedToRepo.Price != 12.34 {
		t.Fatalf("updated fields not applied: %+v", updatedToRepo)
	}
	if updatedToRepo.Author != "Someone" || updatedToRepo.ISBN != "111" || updatedToRepo.PublicationYear != 1999 {
		t.Fatalf("unchanged fields modified: %+v", updatedToRepo)
	}
	if !updatedToRepo.UpdatedAt.After(before) {
		t.Fatalf("UpdatedAt not bumped: %v <= %v", updatedToRepo.UpdatedAt, before)
	}

	// Returned value mirrors what was persisted
	if got.Title != "NewTitle" || got.Price != 12.34 || got.PublicationYear != 1999 {
		t.Fatalf("returned mismatch: %+v", got)
	}
}

func TestUpdateBook_UpdateOnlyPublicationYear(t *testing.T) {
	orig := &domain.Book{
		ID:              8,
		Title:           "Same Title",
		Author:          "Same Author",
		ISBN:            "222",
		PublicationYear: 2010,
		Price:           20.0,
	}
	var updatedToRepo *domain.Book

	m := &mockRepo{
		GetByIDFn: func(ctx context.Context, id int64) (*domain.Book, error) {
			if id != 8 {
				t.Fatalf("expected id 8; got %d", id)
			}
			cp := *orig
			return &cp, nil
		},
		UpdateFn: func(ctx context.Context, b *domain.Book) error {
			updatedToRepo = b
			return nil
		},
	}
	svc := NewBookService(m)

	in := ports.UpdateBookInput{
		PublicationYear: iptr(2020),
		// all other fields nil → unchanged
	}

	got, err := svc.UpdateBook(context.Background(), 8, in)
	if err != nil {
		t.Fatalf("UpdateBook err: %v", err)
	}

	if updatedToRepo == nil {
		t.Fatalf("repo.Update not called")
	}
	if updatedToRepo.PublicationYear != 2020 {
		t.Fatalf("PublicationYear not updated: %+v", updatedToRepo)
	}
	if updatedToRepo.Title != orig.Title || updatedToRepo.Author != orig.Author || updatedToRepo.ISBN != orig.ISBN || updatedToRepo.Price != orig.Price {
		t.Fatalf("other fields should remain unchanged: %+v", updatedToRepo)
	}

	if got.PublicationYear != 2020 {
		t.Fatalf("returned PublicationYear mismatch: %+v", got)
	}
}

func TestUpdateBook_RepoUpdateError(t *testing.T) {
	m := &mockRepo{
		GetByIDFn: func(ctx context.Context, id int64) (*domain.Book, error) {
			return &domain.Book{ID: id, Title: "Old"}, nil
		},
		UpdateFn: func(ctx context.Context, b *domain.Book) error {
			return errors.New("update failed")
		},
	}
	svc := NewBookService(m)

	_, err := svc.UpdateBook(context.Background(), 1, ports.UpdateBookInput{
		Title: strptr("X"),
	})
	if err == nil || err.Error() != "update failed" {
		t.Fatalf("want update failed; got %v", err)
	}
}

func TestDeleteBook_PassThrough(t *testing.T) {
	called := false
	m := &mockRepo{
		DeleteFn: func(ctx context.Context, id int64) error {
			called = true
			if id != 3 {
				t.Fatalf("id mismatch: %d", id)
			}
			return nil
		},
	}
	svc := NewBookService(m)

	if err := svc.DeleteBook(context.Background(), 3); err != nil {
		t.Fatalf("DeleteBook err: %v", err)
	}
	if !called {
		t.Fatalf("repo.Delete not called")
	}
}

func TestDeleteBook_Error(t *testing.T) {
	m := &mockRepo{
		DeleteFn: func(ctx context.Context, id int64) error { return errors.New("boom") },
	}
	svc := NewBookService(m)

	if err := svc.DeleteBook(context.Background(), 9); err == nil || err.Error() != "boom" {
		t.Fatalf("want boom; got %v", err)
	}
}
