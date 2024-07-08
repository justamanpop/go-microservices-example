package main

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi"
)

func Test_routes_exist(t *testing.T) {
	testApp := Config{}

	testRoutes := testApp.routes()
	chiRoutes := testRoutes.(chi.Router)

	expectedRoutes := []string{"/authenticate"}
	for _, route := range expectedRoutes {
		routeExists(t, chiRoutes, route)
	}
}

func routeExists(t *testing.T, routes chi.Router, routeToTest string) {
	found := false
	_ = chi.Walk(routes, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if routeToTest == route {
			found = true
		}
		return nil
	})
	if !found {
		t.Errorf("Did not find route %s in registered routes", routeToTest)
	}
}
