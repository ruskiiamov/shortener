package url

import neturl "net/url"

type OriginalURL struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type DataKeeper interface {
	Add(OriginalURL) (id string, err error)
	Get(id string) (*OriginalURL, error)
}

type converter struct {
	dataKeeper DataKeeper
	baseURL    string
}

func NewConverter(d DataKeeper, baseURL string) *converter {
	return &converter{
		dataKeeper: d,
		baseURL:    baseURL,
	}
}

func (c *converter) Shorten(url string) (string, error) {
	if _, err := neturl.ParseRequestURI(url); err != nil {
		return "", err
	}

	originalURL := OriginalURL{URL: url}

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
