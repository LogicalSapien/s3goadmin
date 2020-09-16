package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHomePage(t *testing.T) {
	// Create a request to pass to handler
	r, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a new responserecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(ListBuckets)

	// Handlers satisfy http.Handler, so can call their ServeHTTP method directly to pass in the request and responserecorder.
	handler.ServeHTTP(rr, r)

	// assert status code of response recorder is OK
	// assert.Equal(t, rr.Code, http.StatusOK, "HandlerFunc(Home) returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
}
