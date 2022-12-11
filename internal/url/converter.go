package url

type OriginalURL struct {
	ID  string
	URL string
}

type DataKeeper interface {
	Add(OriginalURL) (id string, err error)
	Get(id string) (*OriginalURL, error)
}

type converter struct {
	dataKeeper DataKeeper
}

func NewConverter(d DataKeeper) *converter {
	return &converter{dataKeeper: d}
}

func (c *converter) Shorten(host, url string) (string, error) {
	originalURL := OriginalURL{URL: url}

	id, err := c.dataKeeper.Add(originalURL)
	if err != nil {
		return "", err
	}

	return host + "/" + id, nil
}

func (c *converter) GetOriginal(id string) (string, error) {
	originalURL, err := c.dataKeeper.Get(id)
	if err != nil {
		return "", err
	}

	return originalURL.URL, nil
}
