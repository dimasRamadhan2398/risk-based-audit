package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/repositories"
	svcCharter "audit-service/services/audit_charter"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type listChartersBody struct {
	Success bool `json:"success"`
	Data    struct {
		Charters   []models.AuditCharterResponse `json:"charters"`
		Pagination struct {
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			TotalCount int64 `json:"total_count"`
			TotalPages int   `json:"total_pages"`
		} `json:"pagination"`
	} `json:"data"`
}

// createCharterTable builds audit_charters from the GORM schema; AutoMigrate
// cannot run on sqlite because the id defaults to gen_random_uuid()
func createCharterTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&models.AuditCharter{}); err != nil {
		t.Fatal(err)
	}
	var cols []string
	for _, f := range stmt.Schema.Fields {
		if f.DBName == "" {
			continue
		}
		typ := "TEXT"
		switch dt := strings.ToLower(string(f.DataType)); {
		case strings.Contains(dt, "int"), dt == "bool":
			typ = "INTEGER"
		case dt == "time":
			typ = "DATETIME"
		}
		cols = append(cols, "`"+f.DBName+"` "+typ)
	}
	if err := db.Exec("CREATE TABLE `" + stmt.Schema.Table + "` (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
}

// seedCharters inserts versions 1.0, 1.1, ... the way the frontend assigns
// them (latest + 0.1), oldest first, with only the newest active
func seedCharters(t *testing.T, db *gorm.DB, n int) {
	t.Helper()
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < n; i++ {
		c := &models.AuditCharter{
			ID:       uuid.New(),
			Filename: fmt.Sprintf("Audit_Charter_v%d.pdf", i),
			Version:  fmt.Sprintf("%.1f", 1.0+float64(i)/10),
			Title:    fmt.Sprintf("Internal Audit Charter %d", i),
			IsActive: i == n-1,
		}
		c.CreatedAt = base.Add(time.Duration(i) * time.Hour)
		c.UpdatedAt = c.CreatedAt
		if err := db.Create(c).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func newChartersRouter(t *testing.T, n int) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	createCharterTable(t, db)
	seedCharters(t, db, n)

	repo := repositories.NewAuditCharterRepository(repositories.NewBaseRepository(db))
	ctrl := NewAuditCharterController(svcCharter.NewAuditCharterService(repo), nil)
	r := gin.New()
	r.GET("/api/v1/audit-charters", ctrl.ListCharters)
	return r, db
}

func getCharters(t *testing.T, r *gin.Engine, query string) (int, listChartersBody) {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/audit-charters"+query, nil))
	var body listChartersBody
	if w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("decode %q: %v", w.Body.String(), err)
		}
	}
	return w.Code, body
}

// 23 rows is what versions 1.0 .. 3.2 produce when every upload is kept
const chartersV32 = 23

func TestListCharters_DefaultPageReportsFullTotal(t *testing.T) {
	r, _ := newChartersRouter(t, chartersV32)

	code, body := getCharters(t, r, "")
	if code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	p := body.Data.Pagination
	if len(body.Data.Charters) != 10 || p.Page != 1 || p.PageSize != 10 {
		t.Fatalf("default page: got %d rows, page %d, page_size %d; want 10, 1, 10", len(body.Data.Charters), p.Page, p.PageSize)
	}
	if p.TotalCount != chartersV32 || p.TotalPages != 3 {
		t.Fatalf("meta: total_count %d total_pages %d; want %d, 3", p.TotalCount, p.TotalPages, chartersV32)
	}
	if got := body.Data.Charters[0].Version; got != "3.2" {
		t.Fatalf("newest first: first version %q, want 3.2", got)
	}
}

func TestListCharters_PageSizeAndAliases(t *testing.T) {
	r, _ := newChartersRouter(t, chartersV32)

	for _, q := range []string{"?page_size=100", "?limit=100", "?per_page=100"} {
		code, body := getCharters(t, r, q)
		if code != http.StatusOK {
			t.Fatalf("%s: status %d", q, code)
		}
		if len(body.Data.Charters) != chartersV32 || body.Data.Pagination.PageSize != 100 || body.Data.Pagination.TotalPages != 1 {
			t.Fatalf("%s: %d rows, meta %+v", q, len(body.Data.Charters), body.Data.Pagination)
		}
	}

	// page_size wins over the aliases when both are sent
	_, body := getCharters(t, r, "?page_size=5&limit=100")
	if len(body.Data.Charters) != 5 {
		t.Fatalf("page_size precedence: %d rows", len(body.Data.Charters))
	}

	// above the maximum is clamped to 100, and the meta says so
	_, body = getCharters(t, r, "?page_size=1000")
	if body.Data.Pagination.PageSize != 100 || len(body.Data.Charters) != chartersV32 {
		t.Fatalf("clamp: meta %+v rows %d", body.Data.Pagination, len(body.Data.Charters))
	}
}

func TestListCharters_WalkingPagesReturnsEveryRowOnce(t *testing.T) {
	r, db := newChartersRouter(t, chartersV32)
	// identical created_at on every row: order must still be total
	if err := db.Exec("UPDATE audit_charters SET created_at = ?", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)).Error; err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	for page := 1; page <= 5; page++ {
		_, body := getCharters(t, r, fmt.Sprintf("?page=%d&page_size=5", page))
		if body.Data.Pagination.TotalPages != 5 {
			t.Fatalf("total_pages %d", body.Data.Pagination.TotalPages)
		}
		for _, c := range body.Data.Charters {
			if seen[c.ID] {
				t.Fatalf("row %s returned twice", c.ID)
			}
			seen[c.ID] = true
		}
	}
	if len(seen) != chartersV32 {
		t.Fatalf("walked %d distinct rows, want %d", len(seen), chartersV32)
	}
}

func TestListCharters_FiltersAndSoftDelete(t *testing.T) {
	r, db := newChartersRouter(t, chartersV32)

	_, body := getCharters(t, r, "?is_active=false&page_size=100")
	if body.Data.Pagination.TotalCount != chartersV32-1 {
		t.Fatalf("is_active=false total %d", body.Data.Pagination.TotalCount)
	}

	// search is case-insensitive over title, version and filename
	code, body := getCharters(t, r, "?search=INTERNAL%20audit%20charter%201")
	if code != http.StatusOK {
		t.Fatalf("search status %d", code)
	}
	// "... Charter 1" plus "... Charter 10" .. "... Charter 19"
	if body.Data.Pagination.TotalCount != 11 || len(body.Data.Charters) != 10 {
		t.Fatalf("title search: total %d rows %d", body.Data.Pagination.TotalCount, len(body.Data.Charters))
	}
	_, body = getCharters(t, r, "?search=3.2")
	if body.Data.Pagination.TotalCount != 1 || body.Data.Charters[0].Version != "3.2" {
		t.Fatalf("version search: %+v", body.Data.Pagination)
	}

	// soft-deleted rows are excluded from both rows and total
	if err := db.Delete(&models.AuditCharter{}, "version = ?", "1.0").Error; err != nil {
		t.Fatal(err)
	}
	_, body = getCharters(t, r, "?page_size=100")
	if body.Data.Pagination.TotalCount != chartersV32-1 || len(body.Data.Charters) != chartersV32-1 {
		t.Fatalf("after soft delete: total %d rows %d", body.Data.Pagination.TotalCount, len(body.Data.Charters))
	}
}

// The search clause must be parenthesised so it cannot widen the is_active
// filter, and only reference real columns
func TestListCharters_SQLShape(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               gormlogger.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := repositories.NewAuditCharterRepository(repositories.NewBaseRepository(db))
	active := true

	var sqls []string
	capture := func(tx *gorm.DB) {
		sqls = append(sqls, tx.Statement.SQL.String())
		tx.Statement.SQL.Reset()
		tx.Statement.Vars = nil
	}
	_ = db.Callback().Query().After("gorm:query").Register("test:capture", capture)
	_ = db.Callback().Row().After("gorm:row").Register("test:capture_row", capture)

	if _, err := repo.FindMany(20, 10, "x", &active); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Count("x", &active); err != nil {
		t.Fatal(err)
	}
	where := `WHERE (LOWER("title") LIKE $1 OR LOWER("version") LIKE $2 OR LOWER("filename") LIKE $3) AND "is_active" = $4 AND "audit_charters"."deleted_at" IS NULL`
	want := []string{
		`SELECT * FROM "audit_charters" ` + where + ` ORDER BY "created_at" DESC,"id" DESC LIMIT $5 OFFSET $6`,
		`SELECT count(*) FROM "audit_charters" ` + where,
	}
	if len(sqls) != len(want) {
		t.Fatalf("got %d statements: %q", len(sqls), sqls)
	}
	for i := range want {
		if sqls[i] != want[i] {
			t.Fatalf("sql %d:\n got %s\nwant %s", i, sqls[i], want[i])
		}
	}
}
