// Package handler is the Vercel serverless entrypoint. Every request under
// /api/* is rewritten here (see vercel.json) and routed by path.
package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

type Movie struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Year        *int     `json:"year"`
	Genre       *string  `json:"genre"`
	Director    *string  `json:"director"`
	Rating      *float64 `json:"rating"`
	PosterURL   *string  `json:"poster_url"`
	Description *string  `json:"description"`
	CreatedAt   string   `json:"created_at"`
}

var (
	dbOnce sync.Once
	db     *sql.DB
	dbErr  error
)

// getDB opens the pool once per process; serverless instances reuse it across invocations.
func getDB() (*sql.DB, error) {
	dbOnce.Do(func() {
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dbErr = errors.New("DATABASE_URL is not set")
			return
		}
		cfg, err := pgx.ParseConfig(dsn)
		if err != nil {
			dbErr = err
			return
		}
		// Supabase's transaction pooler (port 6543) does not support prepared statements.
		cfg.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
		db = stdlib.OpenDB(*cfg)
		db.SetMaxOpenConns(4)
		db.SetConnMaxIdleTime(5 * time.Minute)
	})
	return db, dbErr
}

const selectCols = `id, title, year, genre, director, rating, poster_url, description, created_at`

func scanMovie(s interface{ Scan(...any) error }) (Movie, error) {
	var m Movie
	var created time.Time
	err := s.Scan(&m.ID, &m.Title, &m.Year, &m.Genre, &m.Director, &m.Rating, &m.PosterURL, &m.Description, &created)
	m.CreatedAt = created.UTC().Format(time.RFC3339)
	return m, err
}

// Handler is the function Vercel invokes.
func Handler(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	// Vercel's rewrite replaces the path with /api/index; the original is passed in __path.
	if p := r.URL.Query().Get("__path"); p != "" {
		path = strings.TrimSuffix("/api/"+p, "/")
	}
	switch {
	case path == "/api/health":
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	case path == "/api/movies":
		listMovies(w, r)
	case strings.HasPrefix(path, "/api/movies/"):
		getMovie(w, r, strings.TrimPrefix(path, "/api/movies/"))
	default:
		writeErr(w, http.StatusNotFound, "not found")
	}
}

// GET /api/movies?q=&genre=&limit=&offset=
func listMovies(w http.ResponseWriter, r *http.Request) {
	conn, err := getDB()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	q := r.URL.Query()
	limit := clampInt(q.Get("limit"), 20, 1, 100)
	offset := clampInt(q.Get("offset"), 0, 0, 1_000_000)

	where := []string{"true"}
	args := []any{}
	if s := strings.TrimSpace(q.Get("q")); s != "" {
		args = append(args, "%"+s+"%")
		where = append(where, "title ilike $"+strconv.Itoa(len(args)))
	}
	if g := strings.TrimSpace(q.Get("genre")); g != "" {
		args = append(args, g)
		where = append(where, "genre ilike $"+strconv.Itoa(len(args)))
	}
	args = append(args, limit, offset)
	sqlStr := "select " + selectCols + " from movies where " + strings.Join(where, " and ") +
		" order by id limit $" + strconv.Itoa(len(args)-1) + " offset $" + strconv.Itoa(len(args))

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	rows, err := conn.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		log.Println("query:", err)
		writeErr(w, http.StatusInternalServerError, dbErrMsg(err))
		return
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		m, err := scanMovie(rows)
		if err != nil {
			log.Println("scan:", err)
			writeErr(w, http.StatusInternalServerError, "database error")
			return
		}
		movies = append(movies, m)
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": movies, "limit": limit, "offset": offset})
}

// GET /api/movies/{id}
func getMovie(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	conn, err := getDB()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	m, err := scanMovie(conn.QueryRowContext(ctx, "select "+selectCols+" from movies where id = $1", id))
	if errors.Is(err, sql.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "movie not found")
		return
	}
	if err != nil {
		log.Println("query:", err)
		writeErr(w, http.StatusInternalServerError, dbErrMsg(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": m})
}

func setCORS(w http.ResponseWriter) {
	origins := os.Getenv("CORS_ORIGINS")
	if origins == "" {
		origins = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origins)
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// dbErrMsg hides driver errors unless DEBUG=1 is set in the environment.
func dbErrMsg(err error) string {
	if os.Getenv("DEBUG") == "1" {
		return "database error: " + err.Error()
	}
	return "database error"
}

func clampInt(s string, def, min, max int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}
