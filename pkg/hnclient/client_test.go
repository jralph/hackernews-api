package hnclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/jralph/hackernews-api/pkg/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	exampleStory = scraper.ItemResponse{
		By:          "exampleuser",
		Descendants: 4,
		ID:          123,
		Kids:        []int{456, 789, 123, 456},
		Score:       123,
		Time:        1210981217,
		Title:       "My Story: Hello!",
		Type:        "story",
		URL:         "https://google.com",
	}
	exampleComment = scraper.ItemResponse{
		By:          "exampleuser",
		Descendants: 4,
		ID:          123,
		Parent:      555,
		Kids:        []int{456, 789, 123, 456},
		Time:        1210981217,
		Text:        "Comment: Hello!",
		Type:        "comment",
	}
	exampleAsk = scraper.ItemResponse{
		By:          "exampleuser",
		Descendants: 4,
		ID:          123,
		Parent:      555,
		Kids:        []int{456, 789, 123, 456},
		Score:       123,
		Time:        1210981217,
		Title:       "Ask: Ask me anything!",
		Text:        "Some example ask text!",
		Type:        "story",
	}
	exampleJob = scraper.ItemResponse{
		By:    "exampleuser",
		ID:    123,
		Score: 123,
		Time:  1210981217,
		Title: "My Job: Hello!",
		Text:  "Some example job text!",
		Type:  "job",
		URL:   "https://google.com",
	}
	examplePoll = scraper.ItemResponse{
		By:          "exampleuser",
		Descendants: 4,
		ID:          123,
		Kids:        []int{456, 789, 123, 456},
		Parts:       []int{777, 666, 888},
		Score:       123,
		Time:        1210981217,
		Title:       "My Poll: Hello!",
		Text:        "Some example poll text!",
		Type:        "poll",
	}
	examplePollPart = scraper.ItemResponse{
		By:    "exampleuser",
		ID:    123,
		Poll:  444,
		Score: 123,
		Text:  "Some example poll option text!",
		Type:  "pollopt",
	}
)

type MockHTTPClient struct {
	mock.Mock

	DoResponse struct {
		Response *http.Response
		Error    error
	}
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoResponse.Response, m.DoResponse.Error
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	t.Run("NewScraper returns instance of Client", func(t *testing.T) {
		require.IsType(t, &Client{}, client)
	})
}

func TestTopStories(t *testing.T) {
	type test struct {
		url              string
		expected         scraper.TopStoriesResponse
		err              error
		statusCode       int
		httpRequestError bool
		emptyBody        bool
		invalidBody      bool
	}

	tests := map[string]test{
		"Client returns list of item ids":            {url: DefaultURL, expected: []int{}, err: nil, statusCode: http.StatusOK},
		"Client returns expected list of item ids":   {url: DefaultURL, expected: []int{1234, 5678, 9012}, err: nil, statusCode: http.StatusOK},
		"Client handles http status error":           {url: DefaultURL, expected: []int{}, err: &IncorrectHTTPStatusCodeError{}, statusCode: http.StatusBadRequest},
		"Client handles http request error":          {url: DefaultURL, expected: []int{}, err: &HTTPResponseError{}, httpRequestError: true},
		"Client handles http request creation error": {url: " http://example.com", expected: []int{}, err: &HTTPRequestError{}},
		"Client handles body read error":             {url: DefaultURL, expected: []int{}, err: &ResponseParseError{}, statusCode: http.StatusOK, emptyBody: true},
		"Client handles body parse error":            {url: DefaultURL, expected: []int{}, err: &ResponseParseError{}, statusCode: http.StatusOK, invalidBody: true},
	}

	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			httpClient := &MockHTTPClient{}
			var body []byte
			body, _ = json.Marshal(opts.expected)
			if opts.emptyBody {
				body = []byte("")
			}
			if opts.invalidBody {
				body = []byte("invalid")
			}
			httpClient.DoResponse.Response = &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
				StatusCode: opts.statusCode,
			}

			if opts.httpRequestError {
				httpClient.DoResponse.Error = errors.New("mock: error")
			} else {
				httpClient.DoResponse.Error = nil
			}

			client := NewClient(
				WithHTTPClient(httpClient),
				WithAPIBaseURL(opts.url),
			)
			stories, err := client.TopStories()

			require.IsType(t, scraper.TopStoriesResponse{}, stories)
			if opts.err != nil {
				require.Error(t, err)
				require.IsType(t, opts.err, err)
				assert.Len(t, stories, 0)
			} else {
				require.NoError(t, err)
				assert.Len(t, stories, len(opts.expected))

				for _, id := range opts.expected {
					assert.Contains(t, stories, id)
				}
			}
		})
	}
}

func TestItem(t *testing.T) {
	type test struct {
		url              string
		id               int
		expected         *scraper.ItemResponse
		err              error
		statusCode       int
		httpRequestError bool
		emptyBody        bool
		invalidBody      bool
	}

	tests := map[string]test{
		"Client returns item":                        {url: DefaultURL, id: 123, expected: &scraper.ItemResponse{}, err: nil, statusCode: http.StatusOK},
		"Client returns expected item for given id":  {url: DefaultURL, id: 123, expected: &exampleStory, err: nil, statusCode: http.StatusOK},
		"Client handles http status error":           {url: DefaultURL, expected: &scraper.ItemResponse{}, err: &IncorrectHTTPStatusCodeError{}, statusCode: http.StatusBadRequest},
		"Client handles http request error":          {url: DefaultURL, expected: &scraper.ItemResponse{}, err: &HTTPResponseError{}, httpRequestError: true},
		"Client handles http request creation error": {url: " http://example.com", expected: &scraper.ItemResponse{}, err: &HTTPRequestError{}},
		"Client handles body read error":             {url: DefaultURL, expected: &scraper.ItemResponse{}, err: &ResponseParseError{}, statusCode: http.StatusOK, emptyBody: true},
		"Client handles body parse error":            {url: DefaultURL, expected: &scraper.ItemResponse{}, err: &ResponseParseError{}, statusCode: http.StatusOK, invalidBody: true},
	}

	httpClient := &MockHTTPClient{}

	for name, opts := range tests {
		t.Run(name, func(t *testing.T) {
			var body []byte
			body, _ = json.Marshal(opts.expected)
			if opts.emptyBody {
				body = []byte("")
			}
			if opts.invalidBody {
				body = []byte("invalid")
			}
			httpClient.DoResponse.Response = &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
				StatusCode: opts.statusCode,
			}

			if opts.httpRequestError {
				httpClient.DoResponse.Error = errors.New("mock: error")
			} else {
				httpClient.DoResponse.Error = nil
			}

			client := NewClient(
				WithHTTPClient(httpClient),
				WithAPIBaseURL(opts.url),
			)
			item, err := client.Item(opts.id)

			require.IsType(t, &scraper.ItemResponse{}, item)
			if opts.err != nil {
				require.Error(t, err)
				require.IsType(t, opts.err, err)
			} else {
				require.NoError(t, err)

				assert.Equal(t, opts.expected, item)
			}
		})
	}
}
