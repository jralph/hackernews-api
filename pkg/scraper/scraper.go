package scraper

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
	By string `json:"by,omitempty"`
	Descendants int `json:"descendants,omitempty"`
	ID int `json:"id"`
	Kids []int `json:"kids,omitempty"`
	Score int `json:"score,omitempty"`
	Time int `json:"time,omitempty"`
	Title string `json:"title,omitempty"`
	Type string `json:"type,omitempty"`
	URL string `json:"url,omitempty"`
	Text string `json:"text,omitempty"`
	Parts []int `json:"parts,omitempty"`
	Poll int `json:"poll,omitempty"`
	Parent int `json:"parent,omitempty"`
	Deleted bool `json:"deleted,omitempty"`
	Dead bool `json:"dead,omitempty"`
}

type Scraper struct {
	saver Saver
	client Client
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

func NewScraper(opts ...Option) *Scraper {
	scraper := &Scraper{}

	for _, opt := range opts {
		opt(scraper)
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

	for _, itemID := range topItems {
		err = s.ScrapeItem(itemID)
		if err != nil {
			return 0, err
		}
	}

	return len(topItems), nil
}

func (s *Scraper) ScrapeItem(id int) error {
	item, err := s.client.Item(id)
	if err != nil {
		return err
	}

	err = s.saver.SaveItem(item)
	if err != nil {
		return err
	}

	for _, itemID := range item.Kids {
		err := s.ScrapeItem(itemID)
		if err != nil {
			return err
		}
	}

	for _, itemID := range item.Parts {
		err := s.ScrapeItem(itemID)
		if err != nil {
			return err
		}
	}

	return nil
}


