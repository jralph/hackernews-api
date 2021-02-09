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
		Error error
	}

	ItemResult struct {
		Response *ItemResponse
		Error error
		MaxKids int
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
		Save bool
	}

	SaveItemResult struct {
		Error error
		Save bool
	}
}

func (m *MockSaver) Clear() {
	m.memoryStore = map[string]string{}
}

func (m *MockSaver) SaveTopStories(topStories TopStoriesResponse) error {
	if m.SaveTopStoriesResult.Save {
		data, _ := json.Marshal(topStories)
		m.memoryStore["topStories"] = string(data)
	}

	return m.SaveTopStoriesResult.Error
}

func (m *MockSaver) SaveItem(item *ItemResponse) error {
	if m.SaveItemResult.Save {
		itemKey := fmt.Sprintf("item_%s_%d", item.Type, item.ID)
		data, _ := json.Marshal(item)
		m.memoryStore[itemKey] = string(data)
	}

	return m.SaveItemResult.Error
}

func TestNewScraper(t *testing.T) {
	scraper := NewScraper()

	t.Run("NewScraper returns implementation of Scraper", func(t *testing.T) {
		require.IsType(t, &Scraper{}, scraper)
	})
}

func TestScrape(t *testing.T) {
	mockClient := &MockHNClient{}
	mockSaver := &MockSaver{
		memoryStore: map[string]string{},
	}
	scraper := NewScraper(
		WithClient(mockClient),
		WithSaver(mockSaver),
	)

	t.Run("Scrape returns number of items scraped", func(t *testing.T) {
		mockSaver.Clear()

		mockClient.TopStoriesResult.Response = TopStoriesResponse{}
		mockClient.TopStoriesResult.Error = nil

		mockClient.ItemResult.Response = &ItemResponse{}
		mockClient.ItemResult.Error = nil
		mockClient.ItemResult.MaxKids = 0

		mockSaver.SaveTopStoriesResult.Save = false
		mockSaver.SaveTopStoriesResult.Error = nil
		mockSaver.SaveItemResult.Save = false
		mockSaver.SaveItemResult.Error = nil

		result, err := scraper.Scrape()
		require.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("Scrape returns expected number of items", func(t *testing.T) {
		expected := 4

		mockSaver.Clear()

		mockClient.TopStoriesResult.Response = TopStoriesResponse{1, 2, 3, 4}
		mockClient.TopStoriesResult.Error = nil

		mockClient.ItemResult.Response = &ItemResponse{}
		mockClient.ItemResult.Error = nil
		mockClient.ItemResult.MaxKids = 0

		mockSaver.SaveTopStoriesResult.Save = false
		mockSaver.SaveTopStoriesResult.Error = nil
		mockSaver.SaveItemResult.Save = false
		mockSaver.SaveItemResult.Error = nil

		result, err := scraper.Scrape()
		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("Scrape handles client http response error", func(t *testing.T) {
		mockSaver.Clear()

		mockClient.TopStoriesResult.Response = TopStoriesResponse{}
		mockClient.TopStoriesResult.Error = errors.New("some mock error")

		mockClient.ItemResult.Response = &ItemResponse{}
		mockClient.ItemResult.Error = nil
		mockClient.ItemResult.MaxKids = 0

		mockSaver.SaveTopStoriesResult.Save = false
		mockSaver.SaveTopStoriesResult.Error = nil
		mockSaver.SaveItemResult.Save = false
		mockSaver.SaveItemResult.Error = nil

		result, err := scraper.Scrape()
		require.Error(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("Scrape saves top items to saver", func(t *testing.T) {
		expected := 4

		topStories := TopStoriesResponse{1, 2, 3, 4}

		mockSaver.Clear()

		mockClient.TopStoriesResult.Response = topStories
		mockClient.TopStoriesResult.Error = nil

		mockClient.ItemResult.Response = &ItemResponse{
			Type: "story",
		}
		mockClient.ItemResult.Error = nil
		mockClient.ItemResult.MaxKids = 0

		mockSaver.SaveTopStoriesResult.Save = true
		mockSaver.SaveTopStoriesResult.Error = nil
		mockSaver.SaveItemResult.Save = false
		mockSaver.SaveItemResult.Error = nil

		expectedSavedTopItems, _ := json.Marshal(topStories)

		result, err := scraper.Scrape()
		stored, ok := mockSaver.memoryStore["topStories"]
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, expected, result)
		assert.Equal(t, string(expectedSavedTopItems), stored)
	})

	t.Run("Scrape saves items to saver", func(t *testing.T) {
		expected := 4

		topStories := TopStoriesResponse{1, 2, 3, 4}

		mockSaver.Clear()

		mockClient.TopStoriesResult.Response = topStories
		mockClient.TopStoriesResult.Error = nil

		mockClient.ItemResult.Response = &ItemResponse{
			Type: "story",
		}
		mockClient.ItemResult.Error = nil
		mockClient.ItemResult.MaxKids = 0

		mockSaver.SaveTopStoriesResult.Save = true
		mockSaver.SaveTopStoriesResult.Error = nil
		mockSaver.SaveItemResult.Save = true
		mockSaver.SaveItemResult.Error = nil

		result, err := scraper.Scrape()
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, mockSaver.memoryStore, 5)
	})

	t.Run("Scrape handles saver error", func(t *testing.T) {
		topStories := TopStoriesResponse{1, 2, 3, 4}

		mockSaver.Clear()

		mockClient.TopStoriesResult.Response = topStories
		mockClient.TopStoriesResult.Error = nil

		mockClient.ItemResult.Response = &ItemResponse{
			Type: "story",
		}
		mockClient.ItemResult.Error = nil
		mockClient.ItemResult.MaxKids = 0

		mockSaver.SaveTopStoriesResult.Save = false
		mockSaver.SaveTopStoriesResult.Error = errors.New("mock error")
		mockSaver.SaveItemResult.Save = false
		mockSaver.SaveItemResult.Error = nil

		_, err := scraper.Scrape()
		require.Error(t, err)
	})

	t.Run("Scrape saves items to saver and follows nested items", func(t *testing.T) {
		expected := 4

		topStories := TopStoriesResponse{1, 2, 3, 4}

		mockSaver.Clear()

		mockClient.TopStoriesResult.Response = topStories
		mockClient.TopStoriesResult.Error = nil

		mockClient.ItemResult.Response = &ItemResponse{
			Type: "story",
		}
		mockClient.ItemResult.Error = nil
		mockClient.ItemResult.MaxKids = 5

		mockSaver.SaveTopStoriesResult.Save = true
		mockSaver.SaveTopStoriesResult.Error = nil
		mockSaver.SaveItemResult.Save = true
		mockSaver.SaveItemResult.Error = nil

		result, err := scraper.Scrape()
		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, mockSaver.memoryStore, 5 + mockClient.ItemResult.MaxKids)
	})
}