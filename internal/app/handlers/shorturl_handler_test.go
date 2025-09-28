package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vlxdisluv/shortener/internal/app/shortener"
	"github.com/vlxdisluv/shortener/internal/app/storage"
)

type MockRepository struct {
	mock.Mock
}

func (mock *MockRepository) Get(hash string) (string, error) {
	args := mock.Called(hash)
	result := args.Get(0)
	return result.(string), args.Error(1)
}

func (m *MockRepository) Close() error { return nil }

func (mock *MockRepository) Save(hash, original string) error {
	args := mock.Called(hash, original)
	return args.Error(0)
}

type MockCounter struct{ mock.Mock }

func (m *MockCounter) Next() (uint64, error) {
	args := m.Called()
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockCounter) Close() error { return nil }

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
			mockReturnErr: storage.ErrNotFound,
			want: want{
				statusCode:  http.StatusNotFound,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			handler := &ShortURLHandler{repo: mockRepo, counter: nil}

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
				//respBody:    "http://example.com/1111113",
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
			mockSaveErr: storage.ErrConflict,
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
			mockCounter := &MockCounter{}
			handler := &ShortURLHandler{repo: mockRepo, counter: mockCounter}

			//mockRepo.On("NextID").Return(int64(2))
			//mockRepo.
			//	On("Save", mock.AnythingOfType("string"), tt.requestBody).
			//	Return(tt.mockSaveErr).
			//	Maybe()

			body := strings.NewReader(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Host = "example.com"

			if tt.requestBody != "" {
				mockCounter.On("Next").Return(uint64(2), nil).Once()
				expectedHash := shortener.Generate(2, 7)

				mockRepo.On("Save", expectedHash, tt.requestBody).
					Return(tt.mockSaveErr).
					Maybe()

				if tt.mockSaveErr == nil {
					tt.want.respBody = "http://example.com/" + expectedHash
				}
			}

			w := httptest.NewRecorder()
			handler.CreateShortURLFromRawBody(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			resp, err := io.ReadAll(result.Body)
			require.NoError(t, err)

			if result.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.respBody, string(resp))
			}

			mockRepo.AssertExpectations(t)
			mockCounter.AssertExpectations(t)
		})
	}
}

func TestCreateShortURLFromJSON(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		respBodyHas string
	}

	type reqBody struct {
		URL string `json:"url"`
	}

	tests := []struct {
		name        string
		requestURL  string
		mockSaveErr error
		want
	}{
		{
			name:        "create short url JSON success #1",
			requestURL:  "http://yandex.ru",
			mockSaveErr: nil,
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
				//respBody:    `{"result":"http://example.com/1111114"}`,
			},
		},
		{
			name:        "empty url field #2",
			requestURL:  "",
			mockSaveErr: nil,
			want: want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			mockCounter := &MockCounter{}
			handler := &ShortURLHandler{repo: mockRepo, counter: mockCounter}

			var b strings.Builder
			_ = json.NewEncoder(&b).Encode(reqBody{URL: tt.requestURL})

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(b.String()))
			req.Host = "example.com"

			if tt.requestURL != "" {
				mockCounter.On("Next").Return(uint64(2), nil).Once()
				expectedHash := shortener.Generate(2, 7)
				mockRepo.On("Save", expectedHash, tt.requestURL).Return(tt.mockSaveErr).Maybe()
				if tt.mockSaveErr == nil {
					tt.want.respBodyHas = "\"result\":\"http://example.com/" + expectedHash + "\""
				}
			}

			w := httptest.NewRecorder()
			handler.CreateShortURLFromJSON(w, req)

			result := w.Result()
			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			data, _ := io.ReadAll(result.Body)
			if result.StatusCode == http.StatusCreated {
				assert.Contains(t, string(data), tt.want.respBodyHas)
			}

			mockRepo.AssertExpectations(t)
			mockCounter.AssertExpectations(t)
		})
	}
}
