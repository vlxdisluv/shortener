package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockRepository struct {
	mock.Mock
}

func (mock *MockRepository) Get(hash string) (string, error) {
	args := mock.Called(hash)
	result := args.Get(0)
	return result.(string), args.Error(1)
}

func (mock *MockRepository) Save(hash, original string) error {
	args := mock.Called(hash, original)
	return args.Error(0)
}

func TestGetShortURL(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name          string
		hash          string
		mockReturnURL string
		mockReturnErr error
		want
	}{
		{
			name:          "get redirect link success #1",
			hash:          "EwHXdJfB",
			mockReturnURL: "http://google.com",
			mockReturnErr: nil,
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name:          "get redirect not found error #2",
			hash:          "EwHXdJfB",
			mockReturnURL: "",
			mockReturnErr: errors.New("short url does not exist for EwHXdJfB"),
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			handler := &ShortURLHandler{
				repo: mockRepo,
			}

			mockRepo.
				On("Get", tt.hash).
				Return(tt.mockReturnURL, tt.mockReturnErr).
				Once()

			req := httptest.NewRequest(http.MethodGet, "/"+tt.hash, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("hash", tt.hash)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()

			handler.GetShortURL(w, req)
			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateShortURLFromRawBody(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		respBody    string
	}
	tests := []struct {
		name        string
		requestBody string
		mockSaveErr error
		want
	}{
		{
			name:        "create short url success #1",
			requestBody: "http://google.com",
			mockSaveErr: nil,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
				respBody:    "http://example.com/EwHXdJfB",
			},
		},
		{
			name:        "empty body error #2",
			requestBody: "",
			mockSaveErr: nil,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				respBody:    "",
			},
		},
		{
			name:        "hash already exists err #3",
			requestBody: "http://google.com",
			mockSaveErr: errors.New("not found"),
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				respBody:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			handler := &ShortURLHandler{
				repo: mockRepo,
			}

			mockRepo.
				On("Save", "EwHXdJfB", tt.requestBody).
				Return(tt.mockSaveErr).
				Maybe()

			body := strings.NewReader(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			w := httptest.NewRecorder()

			handler.CreateShortURLFromRawBody(w, req)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			shortURLResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if result.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.respBody, string(shortURLResult))
			}
		})
	}
}

func TestCreateShortURLFromJSON(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		respBody    string
	}

	tests := []struct {
		name        string
		requestBody string
		mockSaveErr error
		want
	}{
		{
			name:        "create short url success #1",
			requestBody: `{"url":"http://google.com"}`,
			mockSaveErr: nil,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				respBody:    `{"result":"http://example.com/EwHXdJfB"}`,
			},
		},
		{
			name:        "empty body error #2",
			requestBody: "",
			mockSaveErr: nil,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				respBody:    "",
			},
		},
		{
			name:        "hash already exists err #3",
			requestBody: `{"url":"http://google.com"}`,
			mockSaveErr: errors.New("not found"),
			want: want{
				statusCode:  http.StatusInternalServerError,
				contentType: "text/plain; charset=utf-8",
				respBody:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			handler := &ShortURLHandler{
				repo: mockRepo,
			}

			var parsedBody struct {
				URL string `json:"url"`
			}
			_ = json.Unmarshal([]byte(tt.requestBody), &parsedBody)

			mockRepo.
				On("Save", "EwHXdJfB", parsedBody.URL).
				Return(tt.mockSaveErr).
				Maybe()

			body := strings.NewReader(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateShortURLFromJSON(w, req)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			shortURLResult, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			if result.StatusCode == http.StatusCreated {
				assert.JSONEq(t, tt.want.respBody, string(shortURLResult))
			}
		})
	}
}
