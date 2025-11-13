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

type MockShortRepo struct{ mock.Mock }

func (m *MockShortRepo) Save(ctx context.Context, hash, original string) error {
	args := m.Called(ctx, hash, original)
	return args.Error(0)
}
func (m *MockShortRepo) Get(ctx context.Context, hash string) (string, error) {
	args := m.Called(ctx, hash)
	return args.String(0), args.Error(1)
}
func (m *MockShortRepo) GetByOriginal(ctx context.Context, url string) (string, error) {
	args := m.Called(ctx, url)
	return args.String(0), args.Error(1)
}

func (m *MockShortRepo) Close() error                                   { return nil }
func (m *MockShortRepo) WithTx(_ storage.Tx) storage.ShortURLRepository { return m }

type MockCounterRepo struct{ mock.Mock }

func (m *MockCounterRepo) Next(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockCounterRepo) Close() error                                  { return nil }
func (m *MockCounterRepo) WithTx(_ storage.Tx) storage.CounterRepository { return m }

type MockStorage struct {
	short   storage.ShortURLRepository
	counter storage.CounterRepository
	uow     storage.UnitOfWork
}

func (m *MockStorage) ShortURLs() storage.ShortURLRepository { return m.short }
func (m *MockStorage) Counters() storage.CounterRepository   { return m.counter }
func (m *MockStorage) UnitOfWork() storage.UnitOfWork        { return m.uow }

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
			mockShort := &MockShortRepo{}
			mockCounter := &MockCounterRepo{}
			ms := &MockStorage{short: mockShort, counter: mockCounter}
			handler := NewShortURLHandler(ms)

			mockShort.
				On("Get", mock.Anything, tt.hash).
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

			mockShort.AssertExpectations(t)
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
				statusCode:  http.StatusConflict,
				contentType: "text/plain",
				// respBody set dynamically
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockShortRepo{}
			mockCounter := &MockCounterRepo{}
			ms := &MockStorage{short: mockRepo, counter: mockCounter}
			handler := NewShortURLHandler(ms)

			body := strings.NewReader(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/", body)
			req.Host = "example.com"

			if tt.requestBody != "" {
				mockCounter.On("Next", mock.Anything).Return(uint64(2), nil).Once()
				expectedHash := shortener.Generate(2, 7)
				
				mockRepo.On("Save", mock.Anything, expectedHash, tt.requestBody).
					Return(tt.mockSaveErr).
					Maybe()
				
				if tt.mockSaveErr == nil {
					tt.want.respBody = "http://example.com/" + expectedHash
				} else if tt.mockSaveErr == storage.ErrConflict {
					existingHash := "EwHXdJfB"
					mockRepo.On("GetByOriginal", mock.Anything, tt.requestBody).
						Return(existingHash, nil).
						Once()
					tt.want.respBody = "http://example.com/" + existingHash
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
			
			if result.StatusCode == http.StatusCreated || result.StatusCode == http.StatusConflict {
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
			name:        "hash already exists err #2",
			requestURL:  "http://yandex.ru",
			mockSaveErr: storage.ErrConflict,
			want: want{
				statusCode:  http.StatusConflict,
				contentType: "application/json",
			},
		},
		{
			name:        "empty url field #3",
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
			mockShort := &MockShortRepo{}
			mockCounter := &MockCounterRepo{}
			ms := &MockStorage{short: mockShort, counter: mockCounter}
			handler := NewShortURLHandler(ms)

			var b strings.Builder
			_ = json.NewEncoder(&b).Encode(reqBody{URL: tt.requestURL})

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(b.String()))
			req.Host = "example.com"
			
			if tt.requestURL != "" {
				mockCounter.On("Next", mock.Anything).Return(uint64(2), nil).Once()
				expectedHash := shortener.Generate(2, 7)
				mockShort.On("Save", mock.Anything, expectedHash, tt.requestURL).Return(tt.mockSaveErr).Maybe()
				if tt.mockSaveErr == nil {
					tt.want.respBodyHas = "\"result\":\"http://example.com/" + expectedHash + "\""
				} else if tt.mockSaveErr == storage.ErrConflict {
					existingHash := "EwHXdJfB"
					mockShort.On("GetByOriginal", mock.Anything, tt.requestURL).
						Return(existingHash, nil).
						Once()
					tt.want.respBodyHas = "\"result\":\"http://example.com/" + existingHash + "\""
				}
			}
			
			w := httptest.NewRecorder()
			handler.CreateShortURLFromJSON(w, req)
			
			result := w.Result()
			defer result.Body.Close()
			
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			
			data, _ := io.ReadAll(result.Body)
			if result.StatusCode == http.StatusCreated || result.StatusCode == http.StatusConflict {
				assert.Contains(t, string(data), tt.want.respBodyHas)
			}
			
			mockShort.AssertExpectations(t)
			mockShort.AssertExpectations(t)
			mockCounter.AssertExpectations(t)
		})
	}
}
