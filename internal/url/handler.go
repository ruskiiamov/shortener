package url

type OriginalURL struct {
	ID  string
	URL string
}

type Storage interface {
	Add(OriginalURL) (id string, err error)
	Get(id string) (*OriginalURL, error)
}

type handler struct {
	storage Storage
}

func NewHandler(s Storage) *handler {
	return &handler{storage: s}
}

func (h *handler) Shorten(host, url string) (string, error) {
	originalURL := OriginalURL{URL: url}

	id, err := h.storage.Add(originalURL)
	if err != nil {
		return "", err
	}

	return host + "/" + id, nil
}

func (h *handler) GetOriginal(id string) (string, error) {
	originalURL, err := h.storage.Get(id)
	if err != nil {
		return "", err
	}

	return originalURL.URL, nil
}
