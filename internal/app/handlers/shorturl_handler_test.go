package handlers

import (
	"errors"
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
			w := httptest.NewRecorder()

			handler.getShortURL(w, req)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestCreateShortURL(t *testing.T) {
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

			handler.createShortURL(w, req)

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
