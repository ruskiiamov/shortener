package url

import (
	"errors"
	"fmt"
	"math/big"
	neturl "net/url"
)

const base62 = 62

type ErrURLDuplicate struct {
	ID        int
	EncodedID string
	URL       string
}

func (e *ErrURLDuplicate) Error() string {
	return fmt.Sprintf("URL duplicate %s id=%d", e.URL, e.ID)
}

func NewErrURLDuplicate(id int, original string) *ErrURLDuplicate {
	return &ErrURLDuplicate{
		ID:  id,
		URL: original,
	}
}

type ErrURLDeleted struct{}

func (e *ErrURLDeleted) Error() string {
	return "URL deleted"
}

type DataKeeper interface {
	Add(userID, original string) (int, error)
	AddBatch(userID string, originals []string) (map[string]int, error)
	Get(id int) (string, error)
	GetAllByUser(userID string) (map[string]int, error)
	// DeleteBatch(userID string, IDs []int) error
	DeleteBatch(batch map[string][]int) error
	Ping() error
	Close() error
}

type URL struct {
	EncodedID string
	Original  string
}

type Converter interface {
	Shorten(userID, original string) (*URL, error)
	ShortenBatch(userID string, originals []string) ([]URL, error)
	GetOriginal(encodedID string) (*URL, error)
	GetAllByUser(userID string) ([]URL, error)
	// RemoveBatch(userID string, encodedIDs []string) error
	RemoveBatch(batch map[string][]string) error
	PingKeeper() error
}

type converter struct {
	dataKeeper DataKeeper
}

func NewConverter(d DataKeeper) Converter {
	return &converter{dataKeeper: d}
}

func (c *converter) Shorten(userID, original string) (*URL, error) {
	if _, err := neturl.ParseRequestURI(original); err != nil {
		return nil, fmt.Errorf("URL %s not valid: %w", original, err)
	}

	var errDupl *ErrURLDuplicate

	id, err := c.dataKeeper.Add(userID, original)
	if errors.As(err, &errDupl) {
		errDupl.EncodedID = encode(errDupl.ID)
		return nil, errDupl
	}
	if err != nil {
		return nil, fmt.Errorf("URL %s adding error: %w", original, err)
	}

	return &URL{EncodedID: encode(id), Original: original}, nil
}

func (c *converter) ShortenBatch(userID string, originals []string) ([]URL, error) {
	if len(originals) == 0 {
		return nil, errors.New("empty originals")
	}

	originals = unique(originals)

	for _, original := range originals {
		if _, err := neturl.ParseRequestURI(original); err != nil {
			return nil, fmt.Errorf("URL %s not valid: %w", original, err)
		}
	}

	m, err := c.dataKeeper.AddBatch(userID, originals)
	if err != nil {
		return nil, fmt.Errorf("URLs adding error: %w", err)
	}

	var result []URL
	for original, id := range m {
		result = append(result, URL{
			EncodedID: encode(id),
			Original:  original,
		})
	}

	return result, nil
}

func (c *converter) GetOriginal(encodedID string) (*URL, error) {
	id, err := decode(encodedID)
	if err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	original, err := c.dataKeeper.Get(id)
	if err != nil {
		return nil, fmt.Errorf("data keeper error: %w", err)
	}

	return &URL{
		EncodedID: encodedID,
		Original:  original,
	}, nil
}

func (c *converter) GetAllByUser(userID string) ([]URL, error) {
	m, err := c.dataKeeper.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}

	var result []URL
	for original, id := range m {
		result = append(result, URL{
			EncodedID: encode(id),
			Original:  original,
		})
	}

	return result, nil
}

// func (c *converter) RemoveBatch(userID string, encodedIDs []string) error {
// 	if len(encodedIDs) == 0 {
// 		return errors.New("empty encodedIDs")
// 	}

// 	encodedIDs = unique(encodedIDs)

// 	IDs := make([]int, 0, len(encodedIDs))
// 	for _, encodedID := range encodedIDs {
// 		id, err := decode(encodedID)
// 		if err != nil {
// 			return fmt.Errorf("decoding error: %w", err)
// 		}
// 		IDs = append(IDs, id)
// 	}

// 	return c.dataKeeper.DeleteBatch(userID, IDs)
// }

func (c *converter) RemoveBatch(batch map[string][]string) error {
	if len(batch) == 0 {
		return errors.New("empty encodedIDs")
	}

	decodedBatch := make(map[string][]int)

	for userID, encodedIDs := range batch {
		for _, encodedID := range encodedIDs {
			id, err := decode(encodedID)
			if err != nil {
				return fmt.Errorf("decoding error: %w", err)
			}
			decodedBatch[userID] = append(decodedBatch[userID], id)
		}
	}

	return c.dataKeeper.DeleteBatch(decodedBatch)
}

func (c *converter) PingKeeper() error {
	return c.dataKeeper.Ping()
}

func encode(id int) string {
	var i big.Int
	i.SetInt64(int64(id))

	return i.Text(base62)
}

func decode(encodedID string) (int, error) {
	var i big.Int
	_, ok := i.SetString(encodedID, base62)
	if !ok {
		return 0, fmt.Errorf("encoded id not valid: %s", encodedID)
	}

	return int(i.Int64()), nil
}

func unique[T comparable](originals []T) []T {
	result := make([]T, 0)
	m := make(map[T]bool)
	for _, original := range originals {
		if _, ok := m[original]; !ok {
			m[original] = true
			result = append(result, original)
		}
	}

	return result
}
