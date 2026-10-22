package base

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetEmployeeID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bc := &BaseController{}

	cases := []struct {
		name string
		set  func(c *gin.Context)
		want string
	}{
		{"claim present", func(c *gin.Context) { c.Set("employee_id", "emp-uuid") }, "emp-uuid"},
		{"claim empty", func(c *gin.Context) { c.Set("employee_id", "") }, ""},
		{"claim missing", func(c *gin.Context) {}, ""},
		{"wrong type", func(c *gin.Context) { c.Set("employee_id", 42) }, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tc.set(c)
			if got := bc.GetEmployeeID(c); got != tc.want {
				t.Fatalf("GetEmployeeID = %q, want %q", got, tc.want)
			}
		})
	}
}
