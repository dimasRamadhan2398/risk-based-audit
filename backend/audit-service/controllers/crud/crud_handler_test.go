package crud

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type listTestModel struct {
	ID        string
	Name      string
	Status    string
	CreatedAt time.Time
}

// newDryRunDB returns a Postgres-dialect DB that only builds SQL, and a
// pointer to the statements it built
func newDryRunDB(t *testing.T) (*gorm.DB, *[]string) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var sqls []string
	capture := func(tx *gorm.DB) {
		sqls = append(sqls, tx.Statement.SQL.String())
		// A real run resets the built SQL after executing; DryRun does not,
		// which would make Find reuse the Count statement
		tx.Statement.SQL.Reset()
		tx.Statement.Vars = nil
	}
	_ = db.Callback().Query().After("gorm:query").Register("test:capture", capture)
	_ = db.Callback().Row().After("gorm:row").Register("test:capture_row", capture)
	return db, &sqls
}

func runList(t *testing.T, rawQuery string) (int, []string) {
	gin.SetMode(gin.TestMode)
	db, sqls := newDryRunDB(t)
	r := gin.New()
	r.GET("/items", List(db, "Item", func() interface{} { return &[]listTestModel{} }))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/items?"+rawQuery, nil))
	return w.Code, *sqls
}

func TestListRejectsInjectedOrder(t *testing.T) {
	for _, order := range []string{
		"name%3BDROP+TABLE+users",
		"name%3B+DROP+TABLE+users",
		"(SELECT+1)",
		"name+ASC,+pg_sleep(5)",
		"name+DESC+NULLS+FIRST",
		"unknown_column",
	} {
		code, sqls := runList(t, "order="+order)
		if code != http.StatusBadRequest {
			t.Errorf("order=%q: got %d, want 400", order, code)
		}
		if len(sqls) != 0 {
			t.Errorf("order=%q: query was built: %v", order, sqls)
		}
	}
}

func TestListQuotesOrderAndFilters(t *testing.T) {
	code, sqls := runList(t, "order=name+asc,createdAt+DESC&status=Aktif&search=kas&name%3D1+OR+1%3D1--=x")
	if code != http.StatusOK {
		t.Fatalf("got %d, want 200", code)
	}
	all := strings.Join(sqls, "\n")
	for _, want := range []string{
		`ORDER BY "name","created_at" DESC`,
		`"status" = $`,
		`("name" ILIKE $`,
	} {
		if !strings.Contains(all, want) {
			t.Errorf("expected %q in SQL:\n%s", want, all)
		}
	}
	// Filter keys that are not model columns must never reach the SQL
	if strings.Contains(all, "1=1") {
		t.Errorf("injected filter key reached SQL:\n%s", all)
	}
	// Model has no title column, so search must not reference it
	if strings.Contains(all, "title") {
		t.Errorf("search referenced a missing column:\n%s", all)
	}
}

func TestListDefaultOrder(t *testing.T) {
	code, sqls := runList(t, "")
	if code != http.StatusOK {
		t.Fatalf("got %d, want 200", code)
	}
	if !strings.Contains(strings.Join(sqls, "\n"), `ORDER BY "created_at" DESC`) {
		t.Errorf("default order missing:\n%s", strings.Join(sqls, "\n"))
	}
}
