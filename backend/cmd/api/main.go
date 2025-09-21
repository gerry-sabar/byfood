package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	// Import docs NON-blank so we can set SwaggerInfo fields.
	"github.com/gerry-sabar/byfood/docs"

	httpadapter "github.com/gerry-sabar/byfood/internal/adapters/http"
	mysqladapter "github.com/gerry-sabar/byfood/internal/adapters/mysql"
	app "github.com/gerry-sabar/byfood/internal/app"
	"github.com/gerry-sabar/byfood/internal/logger"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           ByFood Books API
// @version         1.0
// @description     Simple Books API with URL cleanup helper.
// @BasePath        /
// @schemes         http
func main() {
	// --- Config & DB ---
	cfg := loadConfig()

	// Configure (optional) Swagger host/schemes at runtime
	// e.g. set APP_HOST=localhost:8080 and APP_SCHEMES=http (or https)
	if host := os.Getenv("APP_HOST"); host != "" {
		docs.SwaggerInfo.Host = host
	}
	if s := os.Getenv("APP_SCHEMES"); s != "" {
		// comma-separated, e.g. "http,https"
		docs.SwaggerInfo.Schemes = nil
		for _, part := range splitAndTrim(s, ",") {
			docs.SwaggerInfo.Schemes = append(docs.SwaggerInfo.Schemes, part)
		}
	}
	docs.SwaggerInfo.BasePath = "/"

	db, err := sqlx.Open("mysql", cfg.DSN())
	if err != nil {
		logger.Log.Error("open db", "error", err)
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(10 * time.Minute)

	if err := ping(db); err != nil {
		logger.Log.Error("db ping", "error", err)
	}

	// --- Services & HTTP handler ---
	repo := mysqladapter.NewBookRepository(db)
	svc := app.NewBookService(repo)
	h := httpadapter.NewHandler(svc)

	// Root router: mount your app and add Swagger UI
	root := chi.NewRouter()
	root.Mount("/", h.Router())

	// Swagger UI at /swagger/index.html
	// Optionally guard with an ENV check if you want it only in non-prod.
	root.Get("/swagger/*", httpSwagger.WrapHandler)

	addr := ":" + cfg.Port
	logger.Log.Info("Application started",
		slog.String("env", os.Getenv("APP_ENV")),
		slog.String("addr", addr),
	)
	if err := http.ListenAndServe(addr, root); err != nil {
		logger.Log.Error("http server exited", "error", err)
	}
}

type config struct {
	User   string
	Pass   string
	Host   string
	PortDB string
	DBName string
	Params string
	Port   string
}

func loadConfig() config {
	return config{
		User:   os.Getenv("MYSQL_USER"),
		Pass:   os.Getenv("MYSQL_PASSWORD"),
		Host:   getEnv("MYSQL_HOST", "db"),
		PortDB: getEnv("MYSQL_PORT", "3306"),
		DBName: getEnv("MYSQL_DATABASE", "booksdb"),
		Params: getEnv("MYSQL_PARAMS", "parseTime=true&charset=utf8mb4&loc=UTC"),
		Port:   getEnv("PORT", "8080"),
	}
}

func (c config) DSN() string {
	// user:pass@tcp(host:port)/dbname?params
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", c.User, c.Pass, c.Host, c.PortDB, c.DBName, c.Params)
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func ping(db *sqlx.DB) error {
	for i := 0; i < 20; i++ {
		if err := db.Ping(); err == nil {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("unable to connect to DB after retries")
}

func splitAndTrim(s, sep string) []string {
	var out []string
	for _, p := range strings.Split(s, sep) {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
