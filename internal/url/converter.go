package url

import neturl "net/url"

type OriginalURL struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	UserID string `json:"user_id"`
}

type URL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type DataKeeper interface {
	Add(OriginalURL) (id string, err error)
	Get(id string) (*OriginalURL, error)
	GetAllByUser(userID string) ([]OriginalURL, error)
	PingDB() error
	Close()
}

type Converter interface {
	Shorten(url, userID string) (string, error)
	GetOriginal(id string) (string, error)
	GetAll(userID string) ([]URL, error)
	PingDB() error
}

type converter struct {
	dataKeeper DataKeeper
	baseURL    string
}

func NewConverter(d DataKeeper, baseURL string) Converter {
	return &converter{
		dataKeeper: d,
		baseURL:    baseURL,
	}
}

func (c *converter) Shorten(url, userID string) (string, error) {
	if _, err := neturl.ParseRequestURI(url); err != nil {
		return "", err
	}

	originalURL := OriginalURL{URL: url, UserID: userID}

	id, err := c.dataKeeper.Add(originalURL)
	if err != nil {
		return "", err
	}

	return c.baseURL + "/" + id, nil
}

func (c *converter) GetOriginal(id string) (string, error) {
	originalURL, err := c.dataKeeper.Get(id)
	if err != nil {
		return "", err
	}

	return originalURL.URL, nil
}

func (c *converter) GetAll(userID string) ([]URL, error) {
	var res []URL

	urls, err := c.dataKeeper.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	for _, url := range urls {
		res = append(res, URL{
			ShortURL:    c.baseURL + "/" + url.ID,
			OriginalURL: url.URL,
		})
	}

	return res, nil
}

func (c *converter) PingDB() error {
	return c.dataKeeper.PingDB()
}
