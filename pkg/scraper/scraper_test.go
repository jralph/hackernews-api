package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockHNClient struct {
	mock.Mock

	TopStoriesResult struct {
		Response TopStoriesResponse
		Error    error
	}

	ItemResult struct {
		Response *ItemResponse
		Error    error
		MaxKids  int
	}

	returnedItemKids int
}

func (m *MockHNClient) TopStories() (TopStoriesResponse, error) {
	return m.TopStoriesResult.Response, m.TopStoriesResult.Error
}

func (m *MockHNClient) Item(id int) (*ItemResponse, error) {
	if m.ItemResult.Error != nil {
		return m.ItemResult.Response, m.ItemResult.Error
	}

	resp := m.ItemResult.Response
	resp.ID = id

	if m.returnedItemKids < m.ItemResult.MaxKids {
		resp.Kids = []int{rand.Int()}
		m.returnedItemKids++
	} else {
		resp.Kids = []int{}
	}

	return resp, m.ItemResult.Error
}

type MockSaver struct {
	mock.Mock

	memoryStore map[string]string

	SaveTopStoriesResult struct {
		Error error
	}

	SaveItemResult struct {
		Error error
	}
}

func (m *MockSaver) SaveTopStories(topStories TopStoriesResponse) error {
	data, _ := json.Marshal(topStories)
	m.memoryStore["topStories"] = string(data)

	return m.SaveTopStoriesResult.Error
}

func (m *MockSaver) SaveItem(item *ItemResponse) error {
	itemKey := fmt.Sprintf("item_%s_%d", item.Type, item.ID)
	data, _ := json.Marshal(item)
	m.memoryStore[itemKey] = string(data)

	return m.SaveItemResult.Error
}

func TestNewScraper(t *testing.T) {
	scraper := NewScraper()

	t.Run("NewScraper returns implementation of Scraper", func(t *testing.T) {
		require.IsType(t, &Scraper{}, scraper)
	})
}

func TestScrape(t *testing.T) {
	type test struct {
		topStoriesResponse *TopStoriesResponse
		topStoriesError    error
		itemsResponse      *ItemResponse
		itemError          error
		maxKids            int
		saverError         error
	}

	tests := map[string]test{
		"Scrape returns number of items scraped":               {topStoriesResponse: &TopStoriesResponse{}},
		"Scrape returns expected number of items":              {topStoriesResponse: &TopStoriesResponse{1, 2, 3, 4}, itemsResponse: &ItemResponse{}},
		"Scrape handles client http response error":            {topStoriesError: errors.New("mock: error")},
		"Scrape saves top items to saver":                      {topStoriesResponse: &TopStoriesResponse{1, 2, 3, 4}, itemsResponse: &ItemResponse{}},
		"Scrape saves items to saver":                          {topStoriesResponse: &TopStoriesResponse{1, 2, 3, 4}, itemsResponse: &ItemResponse{Type: "story"}},
		"Scrape handles saver error":                           {topStoriesResponse: &TopStoriesResponse{1, 2, 3, 4}, itemsResponse: &ItemResponse{}, saverError: errors.New("mock: error")},
		"Scrape saves items to saver and follows nested items": {topStoriesResponse: &TopStoriesResponse{1, 2, 3, 4}, itemsResponse: &ItemResponse{Type: "story"}, maxKids: 5},
	}

	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockHNClient{}
			mockSaver := &MockSaver{
				memoryStore: map[string]string{},
			}
			scraper := NewScraper(
				WithClient(mockClient),
				WithSaver(mockSaver),
			)

			if opts.topStoriesResponse != nil {
				mockClient.TopStoriesResult.Response = *opts.topStoriesResponse
			}

			mockClient.TopStoriesResult.Error = opts.topStoriesError
			mockClient.ItemResult.Response = opts.itemsResponse
			mockClient.ItemResult.Error = opts.itemError
			mockClient.ItemResult.MaxKids = opts.maxKids

			mockSaver.SaveTopStoriesResult.Error = opts.saverError
			mockSaver.SaveItemResult.Error = opts.saverError

			expectedSavedTopItems, _ := json.Marshal(opts.topStoriesResponse)

			result, err := scraper.Scrape()

			stored, ok := mockSaver.memoryStore["topStories"]

			if opts.topStoriesError != nil || opts.itemError != nil || opts.saverError != nil {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.True(t, ok)

				if opts.topStoriesResponse != nil {
					assert.Equal(t, len(*opts.topStoriesResponse), result)
					assert.Equal(t, string(expectedSavedTopItems), stored)
				}

				if opts.itemsResponse != nil {
					expectedKeys := 0
					if len(*opts.topStoriesResponse) > 0 {
						expectedKeys = 1
					}
					assert.Len(t, mockSaver.memoryStore, len(*opts.topStoriesResponse)+opts.maxKids+expectedKeys)
				}
			}
		})
	}
}
