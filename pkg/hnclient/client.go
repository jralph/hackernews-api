package hnclient

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/jralph/hackernews-api/internal/scraper"
)

const (
	DefaultURL = "https://hacker-news.firebaseio.com/v0"
)

type IncorrectHTTPStatusCodeError struct {
	StatusCode int
	URL        string
}

func (e *IncorrectHTTPStatusCodeError) Error() string {
	return fmt.Sprintf("client: got http status code %d for url %s", e.StatusCode, e.URL)
}

type HTTPResponseError struct {
	PreviousError error
	URL           string
}

func (e *HTTPResponseError) Error() string {
	return fmt.Sprintf("client: error making http request: %s", e.PreviousError)
}

type HTTPRequestError struct {
	PreviousError error
	URL           string
}

func (e *HTTPRequestError) Error() string {
	return fmt.Sprintf("client: error creating request for url %s: %s", e.URL, e.PreviousError)
}

type ResponseParseError struct {
	PreviousError error
}

func (e *ResponseParseError) Error() string {
	return fmt.Sprintf("client: erorr parsing response: %s", e.PreviousError)
}

type Client struct {
	httpClient HTTPClient
	url        string
}

type Option func(*Client)

func WithHTTPClient(httpClient HTTPClient) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func WithAPIBaseURL(url string) Option {
	return func(c *Client) {
		c.url = url
	}
}

func NewClient(opts ...Option) *Client {
	defaultClient := &http.Client{
		Timeout: time.Second * 10,
	}

	client := &Client{
		httpClient: defaultClient,
		url:        DefaultURL,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) TopStories() (scraper.TopStoriesResponse, error) {
	resp, err := c.get("/topstories.json")
	if err != nil {
		return scraper.TopStoriesResponse{}, err
	}

	var topStories scraper.TopStoriesResponse
	err = c.parse(resp.Body, &topStories)
	if err != nil {
		return scraper.TopStoriesResponse{}, err
	}

	return topStories, nil
}

func (c *Client) Item(id int) (*scraper.ItemResponse, error) {
	resp, err := c.get(fmt.Sprintf("/item/%d.json", id))
	if err != nil {
		return &scraper.ItemResponse{}, err
	}

	var item scraper.ItemResponse
	err = c.parse(resp.Body, &item)
	if err != nil {
		return &scraper.ItemResponse{}, err
	}

	return &item, nil
}

func (c *Client) get(path string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.url, path)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &HTTPRequestError{PreviousError: err, URL: url}
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return resp, &HTTPResponseError{PreviousError: err, URL: url}
	}

	if resp.StatusCode != 200 {
		return resp, &IncorrectHTTPStatusCodeError{StatusCode: resp.StatusCode, URL: url}
	}

	return resp, nil
}

func (c *Client) parse(readCloser io.ReadCloser, target interface{}) error {
	body, err := ioutil.ReadAll(readCloser)
	if err != nil {
		return &ResponseParseError{PreviousError: err}
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		return &ResponseParseError{PreviousError: err}
	}

	return nil
}
