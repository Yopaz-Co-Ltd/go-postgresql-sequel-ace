package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultPort = "8080"

type server struct {
	mux      *http.ServeMux
	sessions *sessionStore
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*pgxpool.Pool
}

type connectionRequest struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"sslMode"`
}

type queryRequest struct {
	SessionID string        `json:"sessionId"`
	SQL       string        `json:"sql"`
	Params    []interface{} `json:"params"`
	Limit     int           `json:"limit"`
}

type rowsResponse struct {
	Columns []string          `json:"columns"`
	Rows    []json.RawMessage `json:"rows"`
	Count   int               `json:"count"`
}

type tableInfoResponse struct {
	Schema  string       `json:"schema"`
	Name    string       `json:"name"`
	Rows    int64        `json:"rows"`
	Size    string       `json:"size"`
	Columns []columnInfo `json:"columns"`
}

type columnInfo struct {
	Name     string `json:"name"`
	DataType string `json:"dataType"`
	Nullable bool   `json:"nullable"`
}

func main() {
	srv := newServer()
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	addr := ":" + port
	log.Printf("server listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, srv.mux); err != nil {
		log.Fatal(err)
	}
}

func newServer() *server {
	s := &server{
		mux: http.NewServeMux(),
		sessions: &sessionStore{
			sessions: make(map[string]*pgxpool.Pool),
		},
	}

	s.mux.HandleFunc("GET /api/health", s.health)
	s.mux.HandleFunc("POST /api/test-connection", s.testConnection)
	s.mux.HandleFunc("POST /api/connect", s.connect)
	s.mux.HandleFunc("GET /api/databases", s.databases)
	s.mux.HandleFunc("GET /api/schemas", s.schemas)
	s.mux.HandleFunc("GET /api/tables", s.tables)
	s.mux.HandleFunc("GET /api/table-info", s.tableInfo)
	s.mux.HandleFunc("GET /api/table-rows", s.tableRows)
	s.mux.HandleFunc("POST /api/query", s.query)
	s.mux.HandleFunc("DELETE /api/session", s.closeSession)
	s.mux.Handle("/", staticHandler())

	return s
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) testConnection(w http.ResponseWriter, r *http.Request) {
	var req connectionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	pool, err := openPool(ctx, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer pool.Close()

	writeJSON(w, http.StatusOK, map[string]string{"status": "connected"})
}

func (s *server) connect(w http.ResponseWriter, r *http.Request) {
	var req connectionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	pool, err := openPool(ctx, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	id, err := randomID()
	if err != nil {
		pool.Close()
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.sessions.set(id, pool)

	writeJSON(w, http.StatusOK, map[string]string{"sessionId": id})
}

func (s *server) databases(w http.ResponseWriter, r *http.Request) {
	pool, ok := s.poolFromRequest(w, r)
	if !ok {
		return
	}

	rows, err := pool.Query(r.Context(), `
		select datname
		from pg_database
		where datistemplate = false
		order by datname
	`)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		names = append(names, name)
	}
	writeJSON(w, http.StatusOK, names)
}

func (s *server) schemas(w http.ResponseWriter, r *http.Request) {
	pool, ok := s.poolFromRequest(w, r)
	if !ok {
		return
	}

	rows, err := pool.Query(r.Context(), `
		select schema_name
		from information_schema.schemata
		where schema_name not like 'pg_%'
		  and schema_name <> 'information_schema'
		order by schema_name
	`)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		schemas = append(schemas, schema)
	}
	writeJSON(w, http.StatusOK, schemas)
}

func (s *server) tables(w http.ResponseWriter, r *http.Request) {
	pool, ok := s.poolFromRequest(w, r)
	if !ok {
		return
	}

	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = "public"
	}

	rows, err := pool.Query(r.Context(), `
		select table_name, table_type
		from information_schema.tables
		where table_schema = $1
		order by table_name
	`, schema)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	type table struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	var tables []table
	for rows.Next() {
		var t table
		if err := rows.Scan(&t.Name, &t.Type); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		tables = append(tables, t)
	}
	writeJSON(w, http.StatusOK, tables)
}

func (s *server) tableRows(w http.ResponseWriter, r *http.Request) {
	pool, ok := s.poolFromRequest(w, r)
	if !ok {
		return
	}

	schema := r.URL.Query().Get("schema")
	table := r.URL.Query().Get("table")
	if schema == "" || table == "" {
		writeError(w, http.StatusBadRequest, errors.New("schema and table are required"))
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"))
	sql := fmt.Sprintf("select * from %s.%s limit %d", quoteIdent(schema), quoteIdent(table), limit)
	resp, err := queryRows(r.Context(), pool, sql)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) tableInfo(w http.ResponseWriter, r *http.Request) {
	pool, ok := s.poolFromRequest(w, r)
	if !ok {
		return
	}

	schema := r.URL.Query().Get("schema")
	table := r.URL.Query().Get("table")
	if schema == "" || table == "" {
		writeError(w, http.StatusBadRequest, errors.New("schema and table are required"))
		return
	}

	var info tableInfoResponse
	info.Schema = schema
	info.Name = table
	if err := pool.QueryRow(r.Context(), `
		select
			coalesce(c.reltuples::bigint, 0) as estimated_rows,
			pg_size_pretty(pg_total_relation_size(c.oid)) as total_size
		from pg_class c
		join pg_namespace n on n.oid = c.relnamespace
		where n.nspname = $1
		  and c.relname = $2
	`, schema, table).Scan(&info.Rows, &info.Size); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	rows, err := pool.Query(r.Context(), `
		select
			column_name,
			case
				when data_type = 'character varying' then 'VARCHAR'
				when data_type = 'timestamp without time zone' then 'TIMESTAMP'
				when data_type = 'timestamp with time zone' then 'TIMESTAMPTZ'
				when data_type = 'integer' then 'INT'
				when data_type = 'bigint' then 'BIGINT'
				when data_type = 'numeric' then 'NUMERIC'
				when data_type = 'boolean' then 'BOOLEAN'
				else upper(data_type)
			end as data_type,
			is_nullable = 'YES' as nullable
		from information_schema.columns
		where table_schema = $1
		  and table_name = $2
		order by ordinal_position
	`, schema, table)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var column columnInfo
		if err := rows.Scan(&column.Name, &column.DataType, &column.Nullable); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		info.Columns = append(info.Columns, column)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, info)
}

func (s *server) query(w http.ResponseWriter, r *http.Request) {
	var req queryRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		writeError(w, http.StatusBadRequest, errors.New("sql is required"))
		return
	}

	pool, ok := s.sessions.get(req.SessionID)
	if !ok {
		writeError(w, http.StatusUnauthorized, errors.New("invalid or expired session"))
		return
	}

	resp, err := queryRows(r.Context(), pool, req.SQL, req.Params...)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) closeSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, errors.New("sessionId is required"))
		return
	}
	s.sessions.delete(sessionID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func (s *server) poolFromRequest(w http.ResponseWriter, r *http.Request) (*pgxpool.Pool, bool) {
	sessionID := r.URL.Query().Get("sessionId")
	pool, ok := s.sessions.get(sessionID)
	if !ok {
		writeError(w, http.StatusUnauthorized, errors.New("invalid or expired session"))
		return nil, false
	}
	return pool, true
}

func openPool(ctx context.Context, req connectionRequest) (*pgxpool.Pool, error) {
	if req.Host == "" {
		return nil, errors.New("host is required")
	}
	if req.Username == "" {
		return nil, errors.New("username is required")
	}
	if req.Port == 0 {
		req.Port = 5432
	}
	if req.Database == "" {
		req.Database = "postgres"
	}
	if req.SSLMode == "" {
		req.SSLMode = "prefer"
	}

	dsn := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(req.Username, req.Password),
		Host:   req.Host + ":" + strconv.Itoa(req.Port),
		Path:   req.Database,
	}).String()
	dsn += "?sslmode=" + url.QueryEscape(req.SSLMode)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 4
	cfg.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func queryRows(ctx context.Context, pool *pgxpool.Pool, sql string, args ...interface{}) (rowsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return rowsResponse{}, err
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, field := range fields {
		columns[i] = string(field.Name)
	}

	resp := rowsResponse{Columns: columns}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return rowsResponse{}, err
		}
		obj := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			obj[col] = values[i]
		}
		encoded, err := json.Marshal(obj)
		if err != nil {
			return rowsResponse{}, err
		}
		resp.Rows = append(resp.Rows, encoded)
	}
	if err := rows.Err(); err != nil {
		return rowsResponse{}, err
	}
	resp.Count = len(resp.Rows)
	return resp, nil
}

func (s *sessionStore) set(id string, pool *pgxpool.Pool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = pool
}

func (s *sessionStore) get(id string) (*pgxpool.Pool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pool, ok := s.sessions[id]
	return pool, ok
}

func (s *sessionStore) delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if pool, ok := s.sessions[id]; ok {
		pool.Close()
		delete(s.sessions, id)
	}
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func parseLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 {
		return 200
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func staticHandler() http.Handler {
	dist := filepath.Join("..", "frontend", "dist")
	if _, err := os.Stat(dist); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "frontend/dist not found; run npm run dev in frontend for development", http.StatusNotFound)
		})
	}

	fs := http.FileServer(http.Dir(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(dist, strings.TrimPrefix(filepath.Clean(r.URL.Path), string(os.PathSeparator)))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dist, "index.html"))
	})
}
