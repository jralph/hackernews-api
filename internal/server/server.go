package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jralph/hackernews-api/internal/scraper"

	"github.com/labstack/echo/v4"
)

type AllItemsResponse []ItemListing

type ItemListing struct {
	ID       int    `json:"id,omitempty"`
	Location string `json:"location,omitempty"`
}

type ItemResponse struct {
	By          string         `json:"by,omitempty"`
	Descendants int            `json:"descendants,omitempty"`
	ID          int            `json:"id"`
	Kids        []*ItemListing `json:"kids,omitempty"`
	Score       int            `json:"score,omitempty"`
	Time        int            `json:"time,omitempty"`
	Title       string         `json:"title,omitempty"`
	Type        string         `json:"type,omitempty"`
	URL         string         `json:"url,omitempty"`
	Text        string         `json:"text,omitempty"`
	Parts       []*ItemListing `json:"parts,omitempty"`
	Poll        int            `json:"poll,omitempty"`
	Parent      *ItemListing   `json:"parent,omitempty"`
}

type Storage interface {
	GetAllItems() ([]int, error)
	GetItem(int) (*scraper.ItemResponse, error)
}

type Config struct {
	store Storage
}

type Option func(*Config)

func WithStorage(storage Storage) Option {
	return func(c *Config) {
		c.store = storage
	}
}

func CreateServer(opts ...Option) http.Handler {
	e := echo.New()

	conf := &Config{}

	for _, opt := range opts {
		opt(conf)
	}

	if conf.store == nil {
		panic(fmt.Errorf("server: error creating server, must pass `WithStorage` option to CreateServer"))
	}

	e.GET("/items", func(c echo.Context) error {
		items, _ := conf.store.GetAllItems()

		response := AllItemsResponse{}

		for _, id := range items {
			response = append(response, ItemListing{
				ID:       id,
				Location: fmt.Sprintf("/items/%d", id),
			})
		}

		return c.JSON(http.StatusOK, response)
	})

	e.GET("/items/:id", func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		savedItem, _ := conf.store.GetItem(id)

		if savedItem == nil {
			return c.JSON(http.StatusNotFound, nil)
		}

		item := &ItemResponse{
			By:          savedItem.By,
			Descendants: savedItem.Descendants,
			ID:          savedItem.ID,
			Score:       savedItem.Score,
			Time:        savedItem.Time,
			Title:       savedItem.Title,
			Type:        savedItem.Type,
			URL:         savedItem.URL,
			Text:        savedItem.Text,
			Poll:        savedItem.Poll,
		}

		for _, kidID := range savedItem.Kids {
			item.Kids = append(item.Kids, &ItemListing{
				ID:       kidID,
				Location: fmt.Sprintf("/items/%d", kidID),
			})
		}

		for _, PartID := range savedItem.Parts {
			item.Kids = append(item.Kids, &ItemListing{
				ID:       PartID,
				Location: fmt.Sprintf("/items/%d", PartID),
			})
		}

		if savedItem.Parent != 0 {
			item.Parent = &ItemListing{
				ID:       savedItem.Parent,
				Location: fmt.Sprintf("/items/%d", savedItem.Parent),
			}
		}

		return c.JSON(http.StatusOK, item)
	})

	return e
}
