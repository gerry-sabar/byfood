package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	appsvc "github.com/gerry-sabar/byfood/internal/app"
	"github.com/gerry-sabar/byfood/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	svc ports.BookService
}

func NewHandler(svc ports.BookService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)

	r.Route("/books", func(r chi.Router) {
		r.Get("/", h.ListBooks)
		r.Post("/", h.CreateBook)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.GetBook)
			r.Put("/", h.UpdateBook)
			r.Delete("/", h.DeleteBook)
		})
	})

	// 👇 NEW endpoint
	r.Post("/url/cleanup", h.CleanupURL)

	return r
}

// GET /books
// --- ListBooks ---
// ListBooks godoc
// @Summary      List books
// @Description  Returns all books
// @Tags         books
// @Produce      json
// @Success      200  {array}   domain.Book
// @Failure      500  {object}  ports.ErrorResponse
// @Router       /books/ [get]
func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.svc.ListBooks(r.Context())
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, books)
}

// POST /books
// --- CreateBook ---
// CreateBook godoc
// @Summary      Create book
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        body  body      ports.CreateBookInput  true  "New book"
// @Success      201   {object}  domain.Book
// @Failure      400   {object}  ports.ErrorResponse
// @Failure      422   {object}  validationPayload
// @Router       /books/ [post]
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var in ports.CreateBookInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	book, err := h.svc.CreateBook(r.Context(), in)
	if err != nil {
		if ve, ok := err.(*appsvc.ValidationError); ok {
			httpValidation(w, ve)
			return
		}
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonCreated(w, book)
}

// GET /books/{id}
// --- GetBook ---
// GetBook godoc
// @Summary      Get a book
// @Tags         books
// @Produce      json
// @Param        id   path      int  true  "Book ID"  minimum(1)
// @Success      200  {object}  domain.Book
// @Failure      400  {object}  ports.ErrorResponse
// @Failure      404  {object}  ports.ErrorResponse
// @Failure      500  {object}  ports.ErrorResponse
// @Router       /books/{id}/ [get]
func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}
	book, err := h.svc.GetBook(r.Context(), id)
	if err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if book == nil {
		httpError(w, http.StatusNotFound, "not found")
		return
	}
	jsonOK(w, book)
}

// PUT /books/{id}
// --- UpdateBook ---
// UpdateBook godoc
// @Summary      Update a book
// @Tags         books
// @Accept       json
// @Produce      json
// @Param        id    path      int              true  "Book ID"  minimum(1)
// @Param        body  body      ports.UpdateBookInput  true  "Partial update"
// @Success      200   {object}  domain.Book
// @Failure      400   {object}  ports.ErrorResponse
// @Failure      404   {object}  ports.ErrorResponse
// @Failure      422   {object}  validationPayload
// @Router       /books/{id}/ [put]
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}

	var in ports.UpdateBookInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	book, err := h.svc.UpdateBook(r.Context(), id, in)
	if err != nil {
		if ve, ok := err.(*appsvc.ValidationError); ok {
			httpValidation(w, ve)
			return
		}
		// distinguish not found
		if err.Error() == "book not found" {
			httpError(w, http.StatusNotFound, "not found")
			return
		}
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, book)
}

// DELETE /books/{id}
// --- DeleteBook ---
// DeleteBook godoc
// @Summary      Delete a book
// @Tags         books
// @Param        id  path  int  true  "Book ID"  minimum(1)
// @Success      204  "No Content"
// @Failure      400  {object}  ports.ErrorResponse
// @Failure      500  {object}  ports.ErrorResponse
// @Router       /books/{id}/ [delete]
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIDParam(w, r)
	if !ok {
		return
	}
	if err := h.svc.DeleteBook(r.Context(), id); err != nil {
		httpError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helpers

func parseIDParam(w http.ResponseWriter, r *http.Request) (int64, bool) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		httpError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonCreated(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ---- URL Cleanup ----

type cleanupRequest struct {
	URL       string `json:"url"`
	Operation string `json:"operation"` // "redirection" | "canonical" | "all"
}

type cleanupResponse struct {
	ProcessedURL string `json:"processed_url"`
}

// cleanupRequest and cleanupResponse are already declared in your file.

// CleanupURL godoc
// @Summary      Normalize/cleanup a URL
// @Description  operation: "redirection" | "canonical" | "all"
// @Tags         tools
// @Accept       json
// @Produce      json
// @Param        body  body      cleanupRequest   true  "Cleanup payload"
// @Success      200   {object}  cleanupResponse
// @Failure      400   {object}  ports.ErrorResponse
// @Router       /url/cleanup [post]
func (h *Handler) CleanupURL(w http.ResponseWriter, r *http.Request) {
	var req cleanupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	op := strings.ToLower(strings.TrimSpace(req.Operation))
	out, err := processURL(op, req.URL)
	if err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	jsonOK(w, cleanupResponse{ProcessedURL: out})
}

func processURL(op, raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid url")
	}

	switch op {
	case "canonical":
		// Keep host/path as-is; drop query & fragment.
		u.RawQuery = ""
		u.Fragment = ""
		u.Path = strings.TrimSuffix(u.Path, "/")
		return u.String(), nil

	case "redirection":
		return applyRedirection(u), nil

	case "all":
		// redirection + canonical
		redir := applyRedirection(cloneURL(u))
		u2, err := url.Parse(redir)
		if err != nil {
			return "", fmt.Errorf("unexpected parse error")
		}
		u2.RawQuery = ""
		u2.Fragment = ""
		return u2.String(), nil

	default:
		return "", fmt.Errorf("invalid operation (use: redirection|canonical|all)")
	}
}

func applyRedirection(u *url.URL) string {
	// 1) lowercase host and add www. for bare domains (example.com -> www.example.com)
	host := strings.ToLower(u.Host)
	if idx := strings.IndexByte(host, ':'); idx != -1 { // strip port for decision
		hostOnly := host[:idx]
		if needsWWW(hostOnly) {
			hostOnly = "www." + hostOnly
		}
		host = hostOnly + host[idx:]
	} else if needsWWW(host) {
		host = "www." + host
	}

	// 2) lowercase path & drop trailing slash
	path := strings.TrimSuffix(strings.ToLower(u.Path), "/")

	// 3) keep query params but trim trailing slashes from values
	q := u.Query()
	for k, vals := range q {
		for i, v := range vals {
			vals[i] = strings.TrimSuffix(v, "/")
		}
		q[k] = vals
	}

	u.Host = host
	u.Path = path
	u.RawQuery = q.Encode()
	u.Fragment = "" // normalize: drop fragment for redirects
	return u.String()
}

func needsWWW(host string) bool {
	// Add www. only for simple root domains (one dot), e.g., example.com
	// Avoid breaking subdomains like api.example.com
	if strings.HasPrefix(host, "www.") {
		return false
	}
	return strings.Count(host, ".") == 1
}

func cloneURL(u *url.URL) *url.URL {
	c := *u
	return &c
}

type validationPayload struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

func httpValidation(w http.ResponseWriter, ve *appsvc.ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity) // 422
	_ = json.NewEncoder(w).Encode(validationPayload{
		Error:  "validation error",
		Fields: ve.Fields,
	})
}
