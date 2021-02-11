package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	GetAllPosts(*string) ([]int, error)
	GetItem(int) (*scraper.ItemResponse, error)
	Cache(string, time.Duration, interface{}, func() interface{}) error
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

	e.GET("/", func(c echo.Context) error {
		response := map[string]string{
			"items":   "/items",
			"posts":   "/posts",
			"stories": "/stories",
			"jobs":    "/jobs",
		}

		return c.JSON(http.StatusOK, response)
	})

	e.GET("/items", func(c echo.Context) error {
		data := AllItemsResponse{}
		err := conf.store.Cache("items", time.Minute*5, &data, func() interface{} {
			response := AllItemsResponse{}
			items, _ := conf.store.GetAllItems()

			for _, id := range items {
				response = append(response, ItemListing{
					ID:       id,
					Location: fmt.Sprintf("/items/%d", id),
				})
			}

			return &response
		})

		if err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "")
		}

		return c.JSON(http.StatusOK, data)
	})

	e.GET("/posts", func(c echo.Context) error {
		data := AllItemsResponse{}
		err := conf.store.Cache("posts", time.Minute*5, &data, func() interface{} {
			response := AllItemsResponse{}
			items, _ := conf.store.GetAllPosts(nil)

			for _, id := range items {
				response = append(response, ItemListing{
					ID:       id,
					Location: fmt.Sprintf("/items/%d", id),
				})
			}

			return &response
		})

		if err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "")
		}

		return c.JSON(http.StatusOK, data)
	})

	e.GET("/items/:id", func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))

		data := &ItemResponse{}
		err := conf.store.Cache(fmt.Sprintf("item/%d", id), time.Minute*5, data, func() interface{} {
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

			return &item
		})

		if err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "")
		}

		return c.JSON(http.StatusOK, data)
	})

	e.GET("/stories", func(c echo.Context) error {
		data := AllItemsResponse{}
		err := conf.store.Cache("stories", time.Minute*5, &data, func() interface{} {
			response := AllItemsResponse{}
			postType := "story"
			items, _ := conf.store.GetAllPosts(&postType)

			for _, id := range items {
				response = append(response, ItemListing{
					ID:       id,
					Location: fmt.Sprintf("/items/%d", id),
				})
			}

			return &response
		})

		if err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "")
		}

		return c.JSON(http.StatusOK, data)
	})

	e.GET("/jobs", func(c echo.Context) error {
		data := AllItemsResponse{}
		err := conf.store.Cache("jobs", time.Minute*5, &data, func() interface{} {
			response := AllItemsResponse{}
			postType := "job"
			items, _ := conf.store.GetAllPosts(&postType)

			for _, id := range items {
				response = append(response, ItemListing{
					ID:       id,
					Location: fmt.Sprintf("/items/%d", id),
				})
			}

			return &response
		})

		if err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "")
		}

		return c.JSON(http.StatusOK, data)
	})

	return e
}
