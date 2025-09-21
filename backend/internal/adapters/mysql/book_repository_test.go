package mysql

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

	"github.com/gerry-sabar/byfood/internal/domain"
)

// helper to create a sqlx DB backed by sqlmock
func newMockSQLX(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	sqlxDB := sqlx.NewDb(db, "mysql")
	cleanup := func() {
		_ = sqlxDB.Close()
	}
	return sqlxDB, mock, cleanup
}

func TestList_Success(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	// Columns returned must match your scan targets in domain.Book
	cols := []string{"id", "title", "author", "isbn", "publication_year", "price", "created_at", "updated_at"}
	now := time.Now()
	rows := sqlmock.NewRows(cols).
		AddRow(int64(2), "B", "AuthB", "ISBNB", 2015, 21.50, now, now).
		AddRow(int64(1), "A", "AuthA", "ISBNA", 1999, 10.25, now, now)

	// Keep the query matcher readable but specific
	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, title, author, isbn, price, publication_year, created_at, updated_at
		FROM books
		ORDER BY id DESC`,
	)).WillReturnRows(rows)

	r := NewBookRepository(db)
	books, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(books) != 2 {
		t.Fatalf("got %d books; want 2", len(books))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestList_Error(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	mock.ExpectQuery("SELECT .* FROM books").
		WillReturnError(assertErr("boom"))

	r := NewBookRepository(db)
	_, err := r.List(context.Background())
	if err == nil {
		t.Fatalf("expected error; got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetByID_Found(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	cols := []string{"id", "title", "author", "isbn", "publication_year", "price", "created_at", "updated_at"}
	now := time.Now()
	rows := sqlmock.NewRows(cols).
		AddRow(int64(1), "A", "AuthA", "ISBNA", 2001, 9.99, now, now)

	mock.ExpectQuery("SELECT .* FROM books WHERE id = \\?").
		WithArgs(int64(1)).
		WillReturnRows(rows)

	r := NewBookRepository(db)
	got, err := r.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("unexpected result: %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	cols := []string{"id", "title", "author", "isbn", "publication_year", "price", "created_at", "updated_at"}
	rows := sqlmock.NewRows(cols) // no rows -> sql.ErrNoRows inside sqlx.Get

	mock.ExpectQuery("SELECT .* FROM books WHERE id = \\?").
		WithArgs(int64(99)).
		WillReturnRows(rows)

	r := NewBookRepository(db)
	got, err := r.GetByID(context.Background(), 99)
	if err != nil {
		t.Fatalf("GetByID err = %v; want nil (not found treated as nil,nil)", err)
	}
	if got != nil {
		t.Fatalf("expected nil for not found; got %#v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetByID_QueryError(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	mock.ExpectQuery("SELECT .* FROM books WHERE id = \\?").
		WithArgs(int64(5)).
		WillReturnError(assertErr("db down"))

	r := NewBookRepository(db)
	got, err := r.GetByID(context.Background(), 5)
	if err == nil {
		t.Fatalf("expected error; got nil (got=%#v)", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreate_Success(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	// Expect INSERT with 7 args: title, author, isbn, publication_year, price, created_at, updated_at
	mock.ExpectExec("INSERT INTO books").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(123, 1))

	r := NewBookRepository(db)
	id, err := r.Create(context.Background(), &domain.Book{
		// fields can be zero; Exec args matched by AnyArg()
	})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if id != 123 {
		t.Fatalf("id = %d; want 123", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreate_Error(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	// 7 args with publication_year included
	mock.ExpectExec("INSERT INTO books").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(assertErr("insert failed"))

	r := NewBookRepository(db)
	_, err := r.Create(context.Background(), &domain.Book{})
	if err == nil {
		t.Fatalf("expected error; got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdate_Success(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	// Expect UPDATE with 7 args: title, author, isbn, publication_year, price, updated_at, id
	mock.ExpectExec("UPDATE books").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := NewBookRepository(db)
	err := r.Update(context.Background(), &domain.Book{ID: 7})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdate_Error(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	// 7 args including publication_year and id
	mock.ExpectExec("UPDATE books").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(assertErr("update failed"))

	r := NewBookRepository(db)
	err := r.Update(context.Background(), &domain.Book{ID: 7})
	if err == nil {
		t.Fatalf("expected error; got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDelete_Success(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM books WHERE id = \\?").
		WithArgs(int64(9)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	r := NewBookRepository(db)
	err := r.Delete(context.Background(), 9)
	if err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDelete_Error(t *testing.T) {
	db, mock, cleanup := newMockSQLX(t)
	defer cleanup()

	mock.ExpectExec("DELETE FROM books WHERE id = \\?").
		WithArgs(int64(9)).
		WillReturnError(assertErr("delete failed"))

	r := NewBookRepository(db)
	err := r.Delete(context.Background(), 9)
	if err == nil {
		t.Fatalf("expected error; got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// --- small helper error type (avoids importing fmt just for errors) ---

type assertErr string

func (e assertErr) Error() string { return string(e) }

// Ensure sqlmock can compare any custom driver values you might pass (optional)
var _ driver.Valuer
