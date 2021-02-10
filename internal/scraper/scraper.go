package scraper

import (
	"fmt"
	"sync"
)

type Saver interface {
	SaveTopStories(TopStoriesResponse) error
	SaveItem(*ItemResponse) error
}
type Client interface {
	TopStories() (TopStoriesResponse, error)
	Item(int) (*ItemResponse, error)
}

type TopStoriesResponse []int

type ItemResponse struct {
	By          string `json:"by,omitempty"`
	Descendants int    `json:"descendants,omitempty"`
	ID          int    `json:"id"`
	Kids        []int  `json:"kids,omitempty"`
	Score       int    `json:"score,omitempty"`
	Time        int    `json:"time,omitempty"`
	Title       string `json:"title,omitempty"`
	Type        string `json:"type,omitempty"`
	URL         string `json:"url,omitempty"`
	Text        string `json:"text,omitempty"`
	Parts       []int  `json:"parts,omitempty"`
	Poll        int    `json:"poll,omitempty"`
	Parent      int    `json:"parent,omitempty"`
	Deleted     bool   `json:"deleted,omitempty"`
	Dead        bool   `json:"dead,omitempty"`
}

type Scraper struct {
	saver   Saver
	client  Client
	workers int
}

type Option func(*Scraper)

func WithClient(client Client) Option {
	return func(c *Scraper) {
		c.client = client
	}
}

func WithSaver(saver Saver) Option {
	return func(c *Scraper) {
		c.saver = saver
	}
}

func WithWorkerCount(count int) Option {
	return func(c *Scraper) {
		c.workers = count
	}
}

func NewScraper(opts ...Option) *Scraper {
	scraper := &Scraper{
		workers: 1,
	}

	for _, opt := range opts {
		opt(scraper)
	}

	if scraper.saver == nil {
		panic(fmt.Errorf("scraper: option `WithSaver` must be passed to NewScraper"))
	}

	if scraper.client == nil {
		panic(fmt.Errorf("scraper: option `WithClient` must be passed to NewScraper"))
	}

	return scraper
}

func (s *Scraper) Scrape() (int, error) {
	topItems, err := s.client.TopStories()
	if err != nil {
		return 0, err
	}

	err = s.saver.SaveTopStories(topItems)
	if err != nil {
		return 0, err
	}

	err = s.workItems(topItems)
	if err != nil {
		return 0, err
	}

	return len(topItems), nil
}

func (s *Scraper) workItems(items []int) error {
	jobs := make(chan int)
	errs := make(chan error)

	var wg sync.WaitGroup

	for w := 1; w <= s.workers; w++ {
		go func(jobs <-chan int, errs chan<- error, wg *sync.WaitGroup) {
			for {
				j, open := <-jobs
				if !open {
					return
				}

				errs <- s.scrapeItem(j)
			}
		}(jobs, errs, &wg)
	}

	var receivedErrors []error
	go func(errs <-chan error, wg *sync.WaitGroup) {
		for {
			err, open := <-errs
			if !open {
				return
			}

			if err != nil {
				receivedErrors = append(receivedErrors, err)
			}
			wg.Done()
		}
	}(errs, &wg)

	for _, id := range items {
		wg.Add(1)
		jobs <- id
	}

	wg.Wait()
	close(jobs)
	close(errs)

	if len(receivedErrors) > 0 {
		return fmt.Errorf("scrape: worker: error(s) working items to scrape: %s", receivedErrors)
	}

	return nil
}

func (s *Scraper) scrapeItem(id int) error {
	item, err := s.client.Item(id)
	if err != nil {
		return err
	}

	if item.Deleted || item.Dead {
		return nil
	}

	err = s.saver.SaveItem(item)
	if err != nil {
		return err
	}

	nested := append(item.Kids, item.Parts...)

	for _, itemID := range nested {
		err := s.scrapeItem(itemID)
		if err != nil {
			return err
		}
	}

	return nil
}
