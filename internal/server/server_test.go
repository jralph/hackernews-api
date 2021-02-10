package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/jralph/hackernews-api/internal/scraper"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetAllItems() ([]int, error) {
	return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil
}

func (m *MockStorage) GetItem(id int) (*scraper.ItemResponse, error) {
	return &scraper.ItemResponse{}, nil
}

func TestHTTPServerItemsEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost/items", nil)
	w := httptest.NewRecorder()

	storage := &MockStorage{}

	handler := CreateServer(
		WithStorage(storage),
	)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	var response AllItemsResponse
	err := json.Unmarshal(body, &response)

	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json; charset=UTF-8", resp.Header.Get("Content-Type"))
	require.NoError(t, err)
	assert.Len(t, response, 10)

	for _, item := range response {
		assert.IsType(t, string(""), item.Location)
		assert.IsType(t, int(0), item.ID)
	}
}
