package authclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContactsForwardsTokenAndFilters(t *testing.T) {
	var gotQuery, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/users/contacts" {
			t.Errorf("path = %s", r.URL.Path)
		}
		gotQuery = r.URL.RawQuery
		gotAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"success":true,"data":{"users":[{"id":"u1","full_name":"Budi","email":"budi@x.id","department":"Finance"}]}}`))
	}))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, HTTPClient: srv.Client()}

	users, err := c.Contacts(context.Background(), "Bearer t", []string{"u1", "u2"}, " Finance ")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer t" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if gotQuery != "department_head_of=Finance&ids=u1%2Cu2" {
		t.Fatalf("query = %q", gotQuery)
	}
	if len(users) != 1 || users[0].Email != "budi@x.id" {
		t.Fatalf("users = %+v", users)
	}
}

func TestContactsErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, HTTPClient: srv.Client()}
	if _, err := c.Contacts(context.Background(), "", []string{"u1"}, ""); err == nil {
		t.Fatal("missing token must fail before the call")
	}
	if _, err := c.Contacts(context.Background(), "Bearer t", []string{"u1"}, ""); err == nil {
		t.Fatal("403 from auth-service must be an error")
	}
}
