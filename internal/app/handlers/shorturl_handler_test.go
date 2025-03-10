package handlers

import (
	"github.com/vlxdisluv/shortener/internal/app/storage"
	"net/http"
	"reflect"
	"testing"
)

func TestNewShortURLHandler(t *testing.T) {
	type args struct {
		repo storage.URLRepository
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewShortURLHandler(tt.args.repo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewShortURLHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}
