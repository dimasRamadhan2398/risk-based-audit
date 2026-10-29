package masterclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListDepartmentsFollowsPagesAndForwardsToken(t *testing.T) {
	var auths, pages []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/departments" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		auths = append(auths, r.Header.Get("Authorization"))
		pages = append(pages, r.URL.Query().Get("page")+"/"+r.URL.Query().Get("page_size"))
		switch r.URL.Query().Get("page") {
		case "1":
			_, _ = io.WriteString(w, `{"success":true,"data":{"departments":[{"id":"d1","department_name":"Finance"},{"id":"d2","department_name":"IT"}],"pagination":{"page":1,"total_pages":2}}}`)
		case "2":
			_, _ = io.WriteString(w, `{"success":true,"data":{"departments":[{"id":"d3","department_name":"Ops"}],"pagination":{"page":2,"total_pages":2}}}`)
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
			_, _ = io.WriteString(w, `{"success":true,"data":{"departments":[],"pagination":{"total_pages":2}}}`)
		}
	}))
	defer srv.Close()

	got, err := New(srv.URL, srv.Client()).ListDepartments(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].ID != "d1" || got[2].DepartmentName != "Ops" {
		t.Fatalf("departments = %+v", got)
	}
	if len(pages) != 2 || pages[0] != "1/"+departmentPageSize || pages[1] != "2/"+departmentPageSize {
		t.Fatalf("pages requested = %v", pages)
	}
	for _, a := range auths {
		if a != token {
			t.Fatalf("Authorization = %q", a)
		}
	}
}

func TestListDepartmentsBareArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success":true,"data":[{"id":"d1","department_name":"Finance"}]}`)
	}))
	defer srv.Close()

	got, err := New(srv.URL, srv.Client()).ListDepartments(context.Background(), token)
	if err != nil || len(got) != 1 || got[0].DepartmentName != "Finance" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestListDepartmentsErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	c := New(srv.URL, srv.Client())

	if _, err := c.ListDepartments(context.Background(), token); err == nil {
		t.Fatal("401 accepted")
	}
	if _, err := c.ListDepartments(context.Background(), ""); !errors.Is(err, ErrNoAuthorization) {
		t.Fatalf("err = %v, want ErrNoAuthorization", err)
	}
	if _, err := New("", nil).ListDepartments(context.Background(), token); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
}
