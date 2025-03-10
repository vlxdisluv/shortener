package main

import (
	"testing"
)

func Test_shortURLHandler(t *testing.T) {
	type want struct {
		statusCode int
	}
	tests := []struct {
		name    string
		method  string
		request string
		want    want
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}
