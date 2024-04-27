package main

import "net/http"

func isAuthenticated(r *http.Request) bool {
	// auth implementation
	return r.Header.Get("Authorization") == "Bearer "
}
