package crud

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type pagingChild struct {
	ID       uint `gorm:"primaryKey"`
	ParentID string
}

type pagingModel struct {
	ID        string `gorm:"primaryKey" json:"id"`
	Status    string `json:"status"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Children  []pagingChild  `gorm:"foreignKey:ParentID" json:"children"`
}

type pagingResponse struct {
	Data struct {
		Items      []pagingModel `json:"items"`
		Pagination struct {
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			Total      int64 `json:"total"`
			TotalPages int64 `json:"total_pages"`
		} `json:"pagination"`
	} `json:"data"`
}

// newPagingDB seeds 45 rows that all share one created_at (as a batch Create
// does), 5 of them soft-deleted, 15 of the live ones with status DONE
func newPagingDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&pagingModel{}, &pagingChild{}); err != nil {
		t.Fatal(err)
	}
	same := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rows := make([]pagingModel, 0, 45)
	for i := 0; i < 45; i++ {
		status := "DRAFT"
		if i%3 == 0 {
			status = "DONE"
		}
		// IDs inserted out of order so the physical order differs from id order
		rows = append(rows, pagingModel{ID: fmt.Sprintf("id-%02d", (i*7)%45), Status: status, CreatedAt: same})
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, r := range rows[:5] {
		if err := db.Delete(&pagingModel{}, "id = ?", r.ID).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func pageThrough(t *testing.T, db *gorm.DB, filter string, preloads ...string) (seen map[string]int, total int64) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/items", List(db, "Item", func() interface{} { return &[]pagingModel{} }, preloads...))

	seen = map[string]int{}
	for page := 1; ; page++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/items?page_size=7&page=%d%s", page, filter), nil))
		if w.Code != http.StatusOK {
			t.Fatalf("page %d: got %d: %s", page, w.Code, w.Body.String())
		}
		var resp pagingResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		p := resp.Data.Pagination
		if page == 1 {
			total = p.Total
		} else if p.Total != total {
			t.Fatalf("total changed between pages: %d then %d", total, p.Total)
		}
		for _, it := range resp.Data.Items {
			seen[it.ID]++
		}
		if int64(page) >= p.TotalPages {
			break
		}
	}
	return seen, total
}

func TestListPagingCoversEveryRowOnce(t *testing.T) {
	for _, tc := range []struct {
		name     string
		filter   string
		preloads []string
		want     int64
	}{
		{"all", "", nil, 40},
		{"filtered", "&status=DONE", nil, 13},
		{"with preload", "", []string{"Children"}, 40},
		{"explicit non-unique order", "&order=status+ASC", nil, 40},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := newPagingDB(t)
			var live int64
			q := db.Model(&pagingModel{})
			if tc.filter == "&status=DONE" {
				q = q.Where("status = ?", "DONE")
			}
			q.Count(&live)
			if live != tc.want {
				t.Fatalf("fixture: %d live rows, want %d", live, tc.want)
			}

			seen, total := pageThrough(t, db, tc.filter, tc.preloads...)
			if total != tc.want {
				t.Errorf("total = %d, want %d", total, tc.want)
			}
			if int64(len(seen)) != tc.want {
				t.Errorf("paged through %d distinct rows, want %d", len(seen), tc.want)
			}
			for id, n := range seen {
				if n != 1 {
					t.Errorf("row %s returned %d times", id, n)
				}
			}
		})
	}
}
