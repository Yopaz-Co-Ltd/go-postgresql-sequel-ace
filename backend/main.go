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
	Offset    int           `json:"offset"`
}

type rowsResponse struct {
	Columns []string          `json:"columns"`
	Rows    []json.RawMessage `json:"rows"`
	Count   int               `json:"count"`
	Total   int64             `json:"total"`
}

type tableInfoResponse struct {
	Schema  string       `json:"schema"`
	Name    string       `json:"name"`
	Rows    int64        `json:"rows"`
	Size    string       `json:"size"`
	Columns []columnInfo `json:"columns"`
	Indexes []indexInfo  `json:"indexes"`
}

type columnInfo struct {
	Name      string `json:"name"`
	DataType  string `json:"dataType"`
	Length    string `json:"length"`
	Nullable  bool   `json:"nullable"`
	Key       string `json:"key"`
	Default   string `json:"default"`
	Extra     string `json:"extra"`
	Encoding  string `json:"encoding"`
	Collation string `json:"collation"`
	Comment   string `json:"comment"`
}

type indexInfo struct {
	NonUnique   int64  `json:"nonUnique"`
	KeyName     string `json:"keyName"`
	Sequence    int64  `json:"sequence"`
	ColumnName  string `json:"columnName"`
	Collation   string `json:"collation"`
	Cardinality int64  `json:"cardinality"`
	SubPart     string `json:"subPart"`
	Packed      string `json:"packed"`
	Comment     string `json:"comment"`
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

	names := []string{}
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

	schemas := []string{}
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
	tables := []table{}
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
	offset := parseOffset(r.URL.Query().Get("offset"))
	sql := fmt.Sprintf("select * from %s.%s", quoteIdent(schema), quoteIdent(table))
	resp, err := queryRows(r.Context(), pool, sql, limit, offset)
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
		with key_columns as (
			select
				tc.constraint_schema,
				tc.table_name,
				kcu.column_name,
				case
					when bool_or(tc.constraint_type = 'PRIMARY KEY') then 'PRI'
					when bool_or(tc.constraint_type = 'UNIQUE') then 'UNI'
					else ''
				end as key_type
			from information_schema.table_constraints tc
			join information_schema.key_column_usage kcu
			  on kcu.constraint_schema = tc.constraint_schema
			 and kcu.constraint_name = tc.constraint_name
			 and kcu.table_schema = tc.table_schema
			 and kcu.table_name = tc.table_name
			where tc.table_schema = $1
			  and tc.table_name = $2
			  and tc.constraint_type in ('PRIMARY KEY', 'UNIQUE')
			group by tc.constraint_schema, tc.table_name, kcu.column_name
		)
		select
			c.column_name,
			case
				when c.data_type = 'character varying' then 'VARCHAR'
				when c.data_type = 'timestamp without time zone' then 'TIMESTAMP'
				when c.data_type = 'timestamp with time zone' then 'TIMESTAMPTZ'
				when c.data_type = 'integer' then 'INT'
				when c.data_type = 'bigint' then 'BIGINT'
				when c.data_type = 'numeric' then 'NUMERIC'
				when c.data_type = 'boolean' then 'BOOLEAN'
				else upper(c.data_type)
			end as data_type,
			coalesce(
				case
					when c.data_type in ('character varying', 'character', 'character varying') then c.character_maximum_length::text
					else null
				end,
				case
					when c.data_type = 'numeric' and c.numeric_precision is not null and c.numeric_scale is not null
						then c.numeric_precision::text || ',' || c.numeric_scale::text
					when c.data_type = 'numeric' and c.numeric_precision is not null then c.numeric_precision::text
					else ''
				end
			) as length,
			c.is_nullable = 'YES' as nullable,
			coalesce(k.key_type, '') as key_type,
			coalesce(c.column_default, '') as column_default,
			trim(concat(
				case when c.is_identity = 'YES' then 'identity ' else '' end,
				case when c.is_generated <> 'NEVER' then lower(c.is_generated) else '' end
			)) as extra,
			'UTF-8' as encoding,
			coalesce(c.collation_name, '') as collation,
			coalesce(d.description, '') as comment
		from information_schema.columns c
		left join key_columns k
		  on k.constraint_schema = c.table_schema
		 and k.table_name = c.table_name
		 and k.column_name = c.column_name
		left join pg_namespace n
		  on n.nspname = c.table_schema
		left join pg_class cls
		  on cls.relnamespace = n.oid
		 and cls.relname = c.table_name
		left join pg_attribute a
		  on a.attrelid = cls.oid
		 and a.attname = c.column_name
		left join pg_description d
		  on d.objoid = cls.oid
		 and d.objsubid = a.attnum
		where c.table_schema = $1
		  and c.table_name = $2
		order by c.ordinal_position
	`, schema, table)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var column columnInfo
		if err := rows.Scan(
			&column.Name,
			&column.DataType,
			&column.Length,
			&column.Nullable,
			&column.Key,
			&column.Default,
			&column.Extra,
			&column.Encoding,
			&column.Collation,
			&column.Comment,
		); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		info.Columns = append(info.Columns, column)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	indexRows, err := pool.Query(r.Context(), `
		select
			case when i.indisunique then 0 else 1 end as non_unique,
			idx.relname as key_name,
			k.ordinality::bigint as sequence,
			coalesce(a.attname, pg_get_indexdef(i.indexrelid, k.ordinality::int, true)) as column_name,
			case when (i.indoption[k.ordinality::int - 1] & 1) = 1 then 'D' else 'A' end as collation,
			greatest(coalesce(tbl.reltuples::bigint, 0), 0) as cardinality,
			'NULL' as sub_part,
			'NULL' as packed,
			coalesce(obj_description(i.indexrelid, 'pg_class'), '') as comment
		from pg_index i
		join pg_class tbl on tbl.oid = i.indrelid
		join pg_namespace n on n.oid = tbl.relnamespace
		join pg_class idx on idx.oid = i.indexrelid
		cross join lateral unnest(i.indkey) with ordinality as k(attnum, ordinality)
		left join pg_attribute a
		  on a.attrelid = tbl.oid
		 and a.attnum = k.attnum
		where n.nspname = $1
		  and tbl.relname = $2
		order by i.indisprimary desc, idx.relname, k.ordinality
	`, schema, table)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer indexRows.Close()

	for indexRows.Next() {
		var index indexInfo
		if err := indexRows.Scan(
			&index.NonUnique,
			&index.KeyName,
			&index.Sequence,
			&index.ColumnName,
			&index.Collation,
			&index.Cardinality,
			&index.SubPart,
			&index.Packed,
			&index.Comment,
		); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		info.Indexes = append(info.Indexes, index)
	}
	if err := indexRows.Err(); err != nil {
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

	resp, err := queryRows(r.Context(), pool, req.SQL, req.Limit, req.Offset, req.Params...)
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

func queryRows(ctx context.Context, pool *pgxpool.Pool, sql string, limit int, offset int, args ...interface{}) (rowsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	baseSQL := normalizeSQL(sql)
	if baseSQL == "" {
		return rowsResponse{}, errors.New("sql is required")
	}

	total, err := countRows(ctx, pool, baseSQL, args...)
	if err != nil {
		return rowsResponse{}, err
	}

	dataSQL := baseSQL
	if limit > 0 {
		if offset < 0 {
			offset = 0
		}
		dataSQL = fmt.Sprintf("select * from (%s) as __codex_rows limit %d offset %d", baseSQL, limit, offset)
	}

	rows, err := pool.Query(ctx, dataSQL, args...)
	if err != nil {
		return rowsResponse{}, err
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, field := range fields {
		columns[i] = string(field.Name)
	}

	resp := rowsResponse{Columns: columns, Total: total}
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

func parseOffset(raw string) int {
	offset, err := strconv.Atoi(raw)
	if err != nil || offset < 0 {
		return 0
	}
	return offset
}

func normalizeSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	sql = strings.TrimSuffix(sql, ";")
	return strings.TrimSpace(sql)
}

func countRows(ctx context.Context, pool *pgxpool.Pool, sql string, args ...interface{}) (int64, error) {
	var total int64
	countSQL := fmt.Sprintf("select count(*) from (%s) as __codex_count", sql)
	if err := pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
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
