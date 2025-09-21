package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appsvc "github.com/gerry-sabar/byfood/internal/app"
	"github.com/gerry-sabar/byfood/internal/domain"
	"github.com/gerry-sabar/byfood/internal/ports"
)

type cleanupResp struct {
	ProcessedURL string `json:"processed_url"`
}

type mockBookService struct {
	ListBooksFn  func(ctx context.Context) ([]domain.Book, error)
	CreateBookFn func(ctx context.Context, in ports.CreateBookInput) (*domain.Book, error)
	GetBookFn    func(ctx context.Context, id int64) (*domain.Book, error)
	UpdateBookFn func(ctx context.Context, id int64, in ports.UpdateBookInput) (*domain.Book, error)
	DeleteBookFn func(ctx context.Context, id int64) error
}

func decodeCleanup(t *testing.T, res *http.Response) cleanupResp {
	t.Helper()
	defer res.Body.Close()
	var cr cleanupResp
	if err := json.NewDecoder(res.Body).Decode(&cr); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return cr
}

func (m *mockBookService) ListBooks(ctx context.Context) ([]domain.Book, error) {
	return m.ListBooksFn(ctx)
}
func (m *mockBookService) CreateBook(ctx context.Context, in ports.CreateBookInput) (*domain.Book, error) {
	return m.CreateBookFn(ctx, in)
}
func (m *mockBookService) GetBook(ctx context.Context, id int64) (*domain.Book, error) {
	return m.GetBookFn(ctx, id)
}
func (m *mockBookService) UpdateBook(ctx context.Context, id int64, in ports.UpdateBookInput) (*domain.Book, error) {
	return m.UpdateBookFn(ctx, id, in)
}
func (m *mockBookService) DeleteBook(ctx context.Context, id int64) error {
	return m.DeleteBookFn(ctx, id)
}

// --- helpers ---

func newTestServer(t *testing.T, svc ports.BookService) *httptest.Server {
	t.Helper()
	h := NewHandler(svc)
	return httptest.NewServer(h.Router())
}

func do(t *testing.T, ts *httptest.Server, method, path string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, ts.URL+path, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return res
}

func readBody(t *testing.T, r *http.Response) string {
	t.Helper()
	defer r.Body.Close()
	b, _ := io.ReadAll(r.Body)
	return string(b)
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }

// --- ListBooks ---

func TestListBooks_OK(t *testing.T) {
	mock := &mockBookService{
		ListBooksFn: func(ctx context.Context) ([]domain.Book, error) {
			return []domain.Book{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}}, nil
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodGet, "/books/", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	body := readBody(t, res)
	if !contains(body, `"title":"A"`) || !contains(body, `"title":"B"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestListBooks_ServiceError(t *testing.T) {
	mock := &mockBookService{
		ListBooksFn: func(ctx context.Context) ([]domain.Book, error) {
			return nil, io.ErrUnexpectedEOF
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodGet, "/books/", nil)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", res.StatusCode)
	}
	body := readBody(t, res)
	if !contains(body, `"error"`) {
		t.Fatalf("body = %s", body)
	}
}

// --- CreateBook ---

func TestCreateBook_InvalidJSON(t *testing.T) {
	mock := &mockBookService{}
	h := NewHandler(mock)
	ts := httptest.NewServer(h.Router())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/books/", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestCreateBook_ValidationError(t *testing.T) {
	mock := &mockBookService{
		CreateBookFn: func(ctx context.Context, in ports.CreateBookInput) (*domain.Book, error) {
			return nil, &appsvc.ValidationError{Fields: map[string]string{"title": "required"}}
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	payload := map[string]any{"title": ""}
	res := do(t, ts, http.MethodPost, "/books/", payload)
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", res.StatusCode)
	}
	body := readBody(t, res)
	if !contains(body, `"validation error"`) || !contains(body, `"title":"required"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestCreateBook_OK(t *testing.T) {
	mock := &mockBookService{
		CreateBookFn: func(ctx context.Context, in ports.CreateBookInput) (*domain.Book, error) {
			return &domain.Book{ID: 10, Title: in.Title}, nil
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	payload := map[string]any{"title": "New Book", "author": "Me"}
	res := do(t, ts, http.MethodPost, "/books/", payload)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", res.StatusCode)
	}
	body := readBody(t, res)
	if !contains(body, `"id":10`) || !contains(body, `"title":"New Book"`) {
		t.Fatalf("body = %s", body)
	}
}

// --- GetBook ---

func TestGetBook_InvalidID(t *testing.T) {
	mock := &mockBookService{}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodGet, "/books/abc/", nil)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestGetBook_NotFound(t *testing.T) {
	mock := &mockBookService{
		GetBookFn: func(ctx context.Context, id int64) (*domain.Book, error) { return nil, nil },
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodGet, "/books/99/", nil)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

func TestGetBook_Error(t *testing.T) {
	mock := &mockBookService{
		GetBookFn: func(ctx context.Context, id int64) (*domain.Book, error) { return nil, io.ErrUnexpectedEOF },
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodGet, "/books/1/", nil)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", res.StatusCode)
	}
}

func TestGetBook_OK(t *testing.T) {
	mock := &mockBookService{
		GetBookFn: func(ctx context.Context, id int64) (*domain.Book, error) {
			return &domain.Book{ID: id, Title: "X"}, nil
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodGet, "/books/1/", nil)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	body := readBody(t, res)
	if !contains(body, `"title":"X"`) {
		t.Fatalf("body = %s", body)
	}
}

// --- UpdateBook ---

func TestUpdateBook_InvalidID(t *testing.T) {
	ts := newTestServer(t, &mockBookService{})
	defer ts.Close()

	res := do(t, ts, http.MethodPut, "/books/zero/", map[string]any{"title": "A"})
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestUpdateBook_BadJSON(t *testing.T) {
	mock := &mockBookService{}
	h := NewHandler(mock)
	ts := httptest.NewServer(h.Router())
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/books/1/", strings.NewReader("{bad"))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestUpdateBook_ValidationError(t *testing.T) {
	mock := &mockBookService{
		UpdateBookFn: func(ctx context.Context, id int64, in ports.UpdateBookInput) (*domain.Book, error) {
			return nil, &appsvc.ValidationError{Fields: map[string]string{"isbn": "invalid"}}
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodPut, "/books/1/", map[string]any{"isbn": "bad"})
	if res.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", res.StatusCode)
	}
	body := readBody(t, res)
	if !contains(body, `"isbn":"invalid"`) {
		t.Fatalf("body = %s", body)
	}
}

func TestUpdateBook_NotFound(t *testing.T) {
	mock := &mockBookService{
		UpdateBookFn: func(ctx context.Context, id int64, in ports.UpdateBookInput) (*domain.Book, error) {
			return nil, fmtError("book not found")
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodPut, "/books/123/", map[string]any{"title": "Y"})
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

func TestUpdateBook_OK(t *testing.T) {
	mock := &mockBookService{
		UpdateBookFn: func(ctx context.Context, id int64, in ports.UpdateBookInput) (*domain.Book, error) {
			return &domain.Book{ID: id, Title: *in.Title}, nil
		},
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodPut, "/books/3/", map[string]any{"title": "Edited"})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	body := readBody(t, res)
	if !contains(body, `"title":"Edited"`) {
		t.Fatalf("body = %s", body)
	}
}

// --- DeleteBook ---

func TestDeleteBook_InvalidID(t *testing.T) {
	ts := newTestServer(t, &mockBookService{})
	defer ts.Close()

	res := do(t, ts, http.MethodDelete, "/books/-1/", nil)
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", res.StatusCode)
	}
}

func TestDeleteBook_Error(t *testing.T) {
	mock := &mockBookService{
		DeleteBookFn: func(ctx context.Context, id int64) error { return io.ErrUnexpectedEOF },
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodDelete, "/books/10/", nil)
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", res.StatusCode)
	}
}

func TestDeleteBook_NoContent(t *testing.T) {
	mock := &mockBookService{
		DeleteBookFn: func(ctx context.Context, id int64) error { return nil },
	}
	ts := newTestServer(t, mock)
	defer ts.Close()

	res := do(t, ts, http.MethodDelete, "/books/10/", nil)
	if res.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", res.StatusCode)
	}
}

// --- URL Cleanup endpoint ---

func TestCleanupURL_Canonical(t *testing.T) {
	ts := newTestServer(t, &mockBookService{})
	defer ts.Close()

	res := do(t, ts, http.MethodPost, "/url/cleanup", map[string]any{
		"url":       "https://Example.com/Path/To/?a=1#frag",
		"operation": "canonical",
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	cr := decodeCleanup(t, res)
	// Expect query & fragment removed; host/path preserved
	if cr.ProcessedURL != "https://Example.com/Path/To" && cr.ProcessedURL != "https://example.com/Path/To" {
		t.Fatalf("processed_url = %q", cr.ProcessedURL)
	}
}

func TestCleanupURL_Redirection(t *testing.T) {
	ts := newTestServer(t, &mockBookService{})
	defer ts.Close()

	res := do(t, ts, http.MethodPost, "/url/cleanup", map[string]any{
		"url":       "https://example.com/Path/To/?x=1&y=2",
		"operation": "redirection",
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	cr := decodeCleanup(t, res)
	// Expect www + lowercase path, keep query exactly with &
	want := "https://www.example.com/path/to?x=1&y=2"
	if cr.ProcessedURL != want {
		t.Fatalf("processed_url = %q, want %q", cr.ProcessedURL, want)
	}
}

func TestCleanupURL_All(t *testing.T) {
	ts := newTestServer(t, &mockBookService{})
	defer ts.Close()

	res := do(t, ts, http.MethodPost, "/url/cleanup", map[string]any{
		"url":       "https://Sub.Example.com/Path/To/?x=1#frag",
		"operation": "all",
	})
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	cr := decodeCleanup(t, res)
	// redirection + canonical => lowercase path, subdomain preserved, no query/fragment
	if cr.ProcessedURL != "https://sub.example.com/path/to" {
		t.Fatalf("processed_url = %q", cr.ProcessedURL)
	}
}

// --- util for "book not found" error string matching ---

type fmtError string

func (e fmtError) Error() string { return string(e) }
