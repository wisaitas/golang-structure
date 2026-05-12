package main

import (
	"context"
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"go/token"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/caarlos0/env/v11"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/wisaitas/golang-structure/pkg/db/sqlx"
)

//go:embed ignore_tables.txt
var ignoreTablesFile string

func main() {
	ctx := context.Background()

	var outDir string
	flag.StringVar(&outDir, "o", "", "required: output directory for generated <table>.go files")
	flag.StringVar(&outDir, "out", "", "same as -o")

	schema := flag.String("schema", "public", "Postgres schema to introspect")
	ignorePath := flag.String("ignore-file", "", "path to ignore tables file (default: embedded cmd/genentity/ignore_tables.txt)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: go run ./cmd/genentity -o <output-dir> [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags -o / -out are required (no default). Example:\n")
		fmt.Fprintf(os.Stderr, "  go run ./cmd/genentity -o internal/golangstructure/domain/entity/gen\n")
		fmt.Fprintf(os.Stderr, "  (package name is the last folder of -o, e.g. \"gen\" or \"entity\")\n\n")
		fmt.Fprintf(os.Stderr, "Reads SQLDB_* from .env (current dir or repo root) and introspects Postgres.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if strings.TrimSpace(outDir) == "" {
		fmt.Fprintln(os.Stderr, "genentity: missing required flag -o (or -out): output directory for generated .go files")
		flag.Usage()
		os.Exit(2)
	}

	for _, path := range []string{
		".env",
		filepath.Join("..", "..", ".env"),
	} {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}

	var sqlCfg sqlx.Config
	if err := env.Parse(&sqlCfg); err != nil {
		log.Fatalf("env: %v", err)
	}

	if !strings.EqualFold(sqlCfg.Driver, "postgres") && sqlCfg.Driver != "" {
		log.Fatalf("genentity only supports SQLDB_DRIVER=postgres (got %q)", sqlCfg.Driver)
	}

	dsn, err := postgresDSN(sqlCfg)
	if err != nil {
		log.Fatalf("dsn: %v", err)
	}

	rawIgnore := ignoreTablesFile
	if strings.TrimSpace(*ignorePath) != "" {
		b, err := os.ReadFile(*ignorePath)
		if err != nil {
			log.Fatalf("read ignore file: %v", err)
		}
		rawIgnore = string(b)
	}

	exclude := parseIgnoreTables(rawIgnore)
	if len(exclude) == 0 {
		log.Printf("genentity: warning: no ignored tables configured (ignore_tables.txt empty?)")
	}

	outAbs, err := filepath.Abs(outDir)
	if err != nil {
		log.Fatalf("genentity: output path: %v", err)
	}
	pkgName, err := goPackageNameFromOutDir(outAbs)
	if err != nil {
		log.Fatalf("genentity: %v", err)
	}
	if err := run(ctx, genConfig{
		DSN:           dsn,
		Schema:        *schema,
		OutDir:        outAbs,
		PackageName:   pkgName,
		ExcludeTables: exclude,
	}); err != nil {
		log.Fatalf("genentity: %v", err)
	}

	fmt.Fprintf(os.Stdout, "wrote %d table model file(s) under %s (package %s)\n", countTableModelGoFiles(outAbs), outAbs, pkgName)
}

// countTableModelGoFiles counts *.go model files written under dir.
func countTableModelGoFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".go") {
			n++
		}
	}
	return n
}

func postgresDSN(cfg sqlx.Config) (string, error) {
	if cfg.Host == "" || cfg.DBName == "" || cfg.User == "" {
		return "", fmt.Errorf("SQLDB_HOST, SQLDB_DB_NAME, and SQLDB_USER are required")
	}

	port := strings.TrimSpace(cfg.Port)
	if port == "" {
		port = "5432"
	}

	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, port),
		Path:   "/" + strings.TrimPrefix(cfg.DBName, "/"),
	}

	q := url.Values{}
	if strings.TrimSpace(cfg.SSLMode) == "" {
		q.Set("sslmode", "disable")
	} else {
		q.Set("sslmode", cfg.SSLMode)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func parseIgnoreTables(raw string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out[strings.ToLower(line)] = struct{}{}
	}
	return out
}

// goPackageNameFromOutDir returns a valid Go package name from the last path segment of absOut.
func goPackageNameFromOutDir(absOut string) (string, error) {
	base := filepath.Base(filepath.Clean(absOut))
	if base == "." || base == ".." || base == "" || base == string(filepath.Separator) {
		return "", fmt.Errorf("cannot derive Go package name from output path %q", absOut)
	}

	var b strings.Builder
	first := true
	for _, r := range base {
		lr := unicode.ToLower(r)
		if first {
			first = false
			if unicode.IsLetter(lr) || lr == '_' {
				b.WriteRune(lr)
				continue
			}
			if unicode.IsDigit(lr) {
				b.WriteString("pkg")
				b.WriteRune(lr)
				continue
			}
			if lr == '-' {
				b.WriteRune('_')
				continue
			}
			b.WriteRune('_')
			continue
		}
		if unicode.IsLetter(lr) || unicode.IsDigit(lr) || lr == '_' {
			b.WriteRune(lr)
		} else if lr == '-' {
			b.WriteRune('_')
		} else {
			b.WriteRune('_')
		}
	}

	s := b.String()
	if s == "" || strings.Trim(s, "_") == "" {
		return "", fmt.Errorf("folder name %q cannot be turned into a Go package name", base)
	}
	if token.IsKeyword(s) {
		s = s + "pkg"
	}
	if !token.IsIdentifier(s) {
		return "", fmt.Errorf("derived package name %q from folder %q is not a valid Go identifier", s, base)
	}
	return s, nil
}

type genConfig struct {
	DSN           string
	Schema        string
	OutDir        string
	PackageName   string
	ExcludeTables map[string]struct{}
}

type columnRow struct {
	TableName              string
	ColumnName             string
	OrdinalPosition        int
	DataType               string
	UdtName                string
	IsNullable             string
	ColumnDefault          sql.NullString
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
}

type tableModel struct {
	StructName string
	TableName  string
	Fields     []fieldModel
}

type fieldModel struct {
	Name   string
	GoType string
	Tag    string
}

const generatedHeader = `// Code generated by genentity (Postgres introspection). DO NOT EDIT.

`

var modelFileTemplate = template.Must(template.New("modelFile").Parse(`
package {{.PackageName}}
{{if or .NeedsJSON .NeedsTime .NeedsUUID .NeedsGorm}}

import (
{{- if .NeedsJSON}}
	"encoding/json"
{{- end}}
{{- if .NeedsTime}}
	"time"
{{- end}}
{{- if .NeedsUUID}}
	"github.com/google/uuid"
{{- end}}
{{- if .NeedsGorm}}
	"gorm.io/gorm"
{{- end}}
)
{{end}}
{{$t := .Table}}
type {{$t.StructName}} struct {
{{- range $t.Fields}}
	{{.Name}} {{.GoType}} ` + "`{{.Tag}}`" + `
{{- end}}
}

func ({{$t.StructName}}) TableName() string {
	return "{{$t.TableName}}"
}
`))

func run(ctx context.Context, cfg genConfig) error {
	if cfg.DSN == "" {
		return fmt.Errorf("DSN is empty")
	}
	if cfg.Schema == "" {
		cfg.Schema = "public"
	}
	if cfg.OutDir == "" {
		return fmt.Errorf("OutDir is empty")
	}
	if cfg.PackageName == "" {
		return fmt.Errorf("PackageName is empty")
	}
	if cfg.ExcludeTables == nil {
		cfg.ExcludeTables = map[string]struct{}{}
	}

	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	tables, err := listTables(ctx, db, cfg.Schema, cfg.ExcludeTables)
	if err != nil {
		return err
	}

	columns, err := listColumns(ctx, db, cfg.Schema, tables)
	if err != nil {
		return err
	}

	pk, err := listPrimaryKeys(ctx, db, cfg.Schema, tables)
	if err != nil {
		return err
	}

	models := buildModels(tables, columns, pk)
	sort.Slice(models, func(i, j int) bool {
		return models[i].TableName < models[j].TableName
	})

	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	entries, err := os.ReadDir(cfg.OutDir)
	if err != nil {
		return fmt.Errorf("readdir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		_ = os.Remove(filepath.Join(cfg.OutDir, name))
	}

	for _, m := range models {
		needsJSON, needsTime, needsUUID, needsGorm := collectImportsForTable(m)

		var buf strings.Builder
		buf.WriteString(generatedHeader)
		if err := modelFileTemplate.Execute(&buf, map[string]any{
			"PackageName": cfg.PackageName,
			"Table":       m,
			"NeedsJSON":   needsJSON,
			"NeedsTime":   needsTime,
			"NeedsUUID":   needsUUID,
			"NeedsGorm":   needsGorm,
		}); err != nil {
			return fmt.Errorf("template %s: %w", m.TableName, err)
		}

		outPath := filepath.Join(cfg.OutDir, m.TableName+".go")
		if err := os.WriteFile(outPath, []byte(buf.String()), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

func listTables(ctx context.Context, db *sql.DB, schema string, exclude map[string]struct{}) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = $1
  AND table_type = 'BASE TABLE'
ORDER BY table_name
`, schema)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if _, skip := exclude[strings.ToLower(name)]; skip {
			continue
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

func listColumns(ctx context.Context, db *sql.DB, schema string, tables []string) (map[string][]columnRow, error) {
	if len(tables) == 0 {
		return map[string][]columnRow{}, nil
	}

	allowed := map[string]struct{}{}
	for _, t := range tables {
		allowed[t] = struct{}{}
	}

	rows, err := db.QueryContext(ctx, `
SELECT table_name,
       column_name,
       ordinal_position,
       data_type,
       udt_name,
       is_nullable,
       column_default,
       character_maximum_length,
       numeric_precision,
       numeric_scale
FROM information_schema.columns
WHERE table_schema = $1
ORDER BY table_name, ordinal_position
`, schema)
	if err != nil {
		return nil, fmt.Errorf("list columns: %w", err)
	}
	defer rows.Close()

	out := map[string][]columnRow{}
	for rows.Next() {
		var c columnRow
		if err := rows.Scan(
			&c.TableName,
			&c.ColumnName,
			&c.OrdinalPosition,
			&c.DataType,
			&c.UdtName,
			&c.IsNullable,
			&c.ColumnDefault,
			&c.CharacterMaximumLength,
			&c.NumericPrecision,
			&c.NumericScale,
		); err != nil {
			return nil, err
		}
		if _, ok := allowed[c.TableName]; !ok {
			continue
		}
		out[c.TableName] = append(out[c.TableName], c)
	}
	return out, rows.Err()
}

func listPrimaryKeys(ctx context.Context, db *sql.DB, schema string, tables []string) (map[string]map[string]struct{}, error) {
	if len(tables) == 0 {
		return map[string]map[string]struct{}{}, nil
	}

	allowed := map[string]struct{}{}
	for _, t := range tables {
		allowed[t] = struct{}{}
	}

	rows, err := db.QueryContext(ctx, `
SELECT kcu.table_name,
       kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
  ON tc.constraint_name = kcu.constraint_name
 AND tc.table_schema = kcu.table_schema
WHERE tc.table_schema = $1
  AND tc.constraint_type = 'PRIMARY KEY'
`, schema)
	if err != nil {
		return nil, fmt.Errorf("list primary keys: %w", err)
	}
	defer rows.Close()

	out := map[string]map[string]struct{}{}
	for rows.Next() {
		var table, col string
		if err := rows.Scan(&table, &col); err != nil {
			return nil, err
		}
		if _, ok := allowed[table]; !ok {
			continue
		}
		if out[table] == nil {
			out[table] = map[string]struct{}{}
		}
		out[table][col] = struct{}{}
	}
	return out, rows.Err()
}

func buildModels(tables []string, columns map[string][]columnRow, pk map[string]map[string]struct{}) []tableModel {
	var models []tableModel
	for _, table := range tables {
		cols := columns[table]
		if len(cols) == 0 {
			continue
		}

		var fields []fieldModel
		for _, c := range cols {
			fields = append(fields, fieldModel{
				Name:   exportedFieldName(c.ColumnName),
				GoType: pgToGoType(c),
				Tag:    buildGormTag(c, pk[table]),
			})
		}

		models = append(models, tableModel{
			StructName: exportedTableTypeName(table),
			TableName:  table,
			Fields:     fields,
		})
	}
	return models
}

func isTimestampFamily(c columnRow) bool {
	dt := strings.ToLower(strings.TrimSpace(c.DataType))
	return strings.Contains(dt, "timestamp") || strings.EqualFold(c.UdtName, "timestamptz")
}

// collectImportsForTable decides imports for a single generated file.
func collectImportsForTable(m tableModel) (needsJSON, needsTime, needsUUID, needsGorm bool) {
	for _, f := range m.Fields {
		if strings.Contains(f.GoType, "json.") {
			needsJSON = true
		}
		if strings.Contains(f.GoType, "time.") {
			needsTime = true
		}
		if strings.Contains(f.GoType, "uuid.") {
			needsUUID = true
		}
		if strings.Contains(f.GoType, "gorm.") {
			needsGorm = true
		}
	}
	return needsJSON, needsTime, needsUUID, needsGorm
}

func exportedFieldName(col string) string {
	parts := strings.Split(col, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		if strings.EqualFold(p, "id") {
			parts[i] = "ID"
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

func exportedTableTypeName(table string) string {
	parts := strings.Split(table, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

func pgToGoType(c columnRow) string {
	udt := strings.ToLower(strings.TrimSpace(c.UdtName))
	dt := strings.ToLower(strings.TrimSpace(c.DataType))
	nullable := strings.ToUpper(strings.TrimSpace(c.IsNullable)) == "YES"

	colLower := strings.ToLower(c.ColumnName)
	if colLower == "deleted_at" && isTimestampFamily(c) {
		return "*gorm.DeletedAt"
	}

	switch udt {
	case "int2", "serial2":
		return nullableType("int16", nullable)
	case "int4", "serial", "serial4":
		return nullableType("int", nullable)
	case "int8", "serial8":
		return nullableType("int64", nullable)
	case "bool":
		return nullableType("bool", nullable)
	case "float4":
		return nullableType("float32", nullable)
	case "float8":
		return nullableType("float64", nullable)
	case "numeric", "decimal":
		return nullableType("float64", nullable)
	case "uuid":
		return nullableType("uuid.UUID", nullable)
	case "bytea":
		return nullableType("[]byte", nullable)
	case "json", "jsonb":
		return nullableType("json.RawMessage", nullable)
	case "text", "varchar", "bpchar", "name", "citext":
		return nullableType("string", nullable)
	case "date":
		return nullableType("time.Time", nullable)
	case "time", "timetz":
		return nullableType("string", nullable)
	case "timestamp", "timestamptz":
		return nullableType("time.Time", nullable)
	default:
		if strings.Contains(dt, "timestamp") {
			return nullableType("time.Time", nullable)
		}
		if strings.Contains(dt, "character varying") || dt == "character" {
			return nullableType("string", nullable)
		}
		return nullableType("string", nullable)
	}
}

func nullableType(base string, nullable bool) string {
	if !nullable {
		return base
	}
	if strings.HasPrefix(base, "*") {
		return base
	}
	return "*" + base
}

func buildGormTag(c columnRow, pkCols map[string]struct{}) string {
	parts := []string{fmt.Sprintf("column:%s", c.ColumnName)}

	if pkCols != nil {
		if _, ok := pkCols[c.ColumnName]; ok {
			parts = append(parts, "primaryKey")
		}
	}

	if strings.ToUpper(strings.TrimSpace(c.IsNullable)) == "NO" {
		parts = append(parts, "not null")
	}

	if c.ColumnDefault.Valid {
		def := strings.TrimSpace(c.ColumnDefault.String)
		if strings.Contains(strings.ToLower(def), "nextval(") {
			parts = append(parts, "autoIncrement")
		} else if def != "" {
			if len(def) <= 120 {
				parts = append(parts, fmt.Sprintf("default:%s", sqlDefaultForGorm(def)))
			}
		}
	}

	if maxLen := c.CharacterMaximumLength; maxLen.Valid && maxLen.Int64 > 0 {
		parts = append(parts, fmt.Sprintf("size:%d", maxLen.Int64))
	}

	return "gorm:\"" + strings.Join(parts, ";") + "\""
}

func sqlDefaultForGorm(def string) string {
	def = strings.TrimSpace(def)
	if strings.EqualFold(def, "now()") || strings.HasPrefix(strings.ToLower(def), "now()") {
		return "now()"
	}
	if strings.EqualFold(def, "true") || strings.EqualFold(def, "false") {
		return def
	}
	if len(def) >= 2 && def[0] == '\'' && def[len(def)-1] == '\'' {
		inner := strings.ReplaceAll(def[1:len(def)-1], "''", "'")
		return "'" + strings.ReplaceAll(inner, "'", "\\'") + "'"
	}
	return def
}
