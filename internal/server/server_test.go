package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jralph/hackernews-api/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) GetAllItems() ([]int, error) {
	return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil
}

func (m *MockStorage) GetAllPosts(postType *string) ([]int, error) {
	return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil
}

func (m *MockStorage) GetItem(id int) (*scraper.ItemResponse, error) {
	return &scraper.ItemResponse{}, nil
}

func (m *MockStorage) Cache(key string, expireAfter time.Duration, target interface{}, f func() interface{}) error {
	toCache := f()

	encoded, err := json.Marshal(toCache)
	if err != nil {
		return err
	}

	err = json.Unmarshal(encoded, target)
	return err
}

func TestHTTPServerItemsEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost/posts", nil)
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
