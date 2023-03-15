// URL is the core URL shortener logic.
package url

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	neturl "net/url"
)

const base62 = 62

// Error for trying tos shorten existing URL. Contains existing URL data.
type ErrURLDuplicate struct {
	// URL id in data storage.
	ID int

	// URL id for url.
	EncodedID string

	// Original URL.
	URL string
}

// Error implements error interface.
func (e *ErrURLDuplicate) Error() string {
	return fmt.Sprintf("URL duplicate %s id=%d", e.URL, e.ID)
}

// NewErrURLDuplicate returns new error object.
func NewErrURLDuplicate(id int, original string) *ErrURLDuplicate {
	return &ErrURLDuplicate{
		ID:  id,
		URL: original,
	}
}

// Error for trying to get deleted URL.
type ErrURLDeleted struct{}

// Error implements error interface.
func (e *ErrURLDeleted) Error() string {
	return "URL deleted"
}

// DataKeeper is data storage for URLs.
type DataKeeper interface {
	Add(ctx context.Context, userID, original string) (int, error)
	AddBatch(ctx context.Context, userID string, originals []string) (map[string]int, error)
	Get(ctx context.Context, id int) (string, error)
	GetAllByUser(ctx context.Context, userID string) (map[string]int, error)
	DeleteBatch(ctx context.Context, batch map[string][]int) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

// URL is the core entity for URL shortener.
type URL struct {
	// EncodedID is used in shortened URL.
	EncodedID string

	// Original is the original URL.
	Original string
}

// Converter is the core logic to operate with URL.
type Converter interface {
	Shorten(ctx context.Context, userID, original string) (*URL, error)
	ShortenBatch(ctx context.Context, userID string, originals []string) ([]URL, error)
	GetOriginal(ctx context.Context, encodedID string) (*URL, error)
	GetAllByUser(ctx context.Context, userID string) ([]URL, error)
	RemoveBatch(ctx context.Context, batch map[string][]string) error
	PingKeeper(ctx context.Context) error
}

type converter struct {
	dataKeeper DataKeeper
}

// Converter returns object that implements Converter interface.
func NewConverter(d DataKeeper) Converter {
	return &converter{dataKeeper: d}
}

// Shorten returns URL object with encoded id or ErrURLDuplicate in case of
// trying to shorten existing URL.
func (c *converter) Shorten(ctx context.Context, userID, original string) (*URL, error) {
	if _, err := neturl.ParseRequestURI(original); err != nil {
		return nil, fmt.Errorf("URL %s not valid: %w", original, err)
	}

	var errDupl *ErrURLDuplicate

	id, err := c.dataKeeper.Add(ctx, userID, original)
	if errors.As(err, &errDupl) {
		errDupl.EncodedID = encode(errDupl.ID)
		return nil, errDupl
	}
	if err != nil {
		return nil, fmt.Errorf("URL %s adding error: %w", original, err)
	}

	return &URL{EncodedID: encode(id), Original: original}, nil
}

// ShortenBatch returns a slice of URL objects with shortened IDs.
func (c *converter) ShortenBatch(ctx context.Context, userID string, originals []string) ([]URL, error) {
	if len(originals) == 0 {
		return nil, errors.New("empty originals")
	}

	originals = unique(originals)

	for _, original := range originals {
		if _, err := neturl.ParseRequestURI(original); err != nil {
			return nil, fmt.Errorf("URL %s not valid: %w", original, err)
		}
	}

	m, err := c.dataKeeper.AddBatch(ctx, userID, originals)
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

// GetOriginal returns URL object by shortened id.
func (c *converter) GetOriginal(ctx context.Context, encodedID string) (*URL, error) {
	id, err := decode(encodedID)
	if err != nil {
		return nil, fmt.Errorf("decoding error: %w", err)
	}

	original, err := c.dataKeeper.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("data keeper error: %w", err)
	}

	return &URL{
		EncodedID: encodedID,
		Original:  original,
	}, nil
}

// GetAllByUser returns a slice of URL objects with all user URLs.
func (c *converter) GetAllByUser(ctx context.Context, userID string) ([]URL, error) {
	m, err := c.dataKeeper.GetAllByUser(ctx, userID)
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

// RemoveBatch removes URL batch by encoded IDs.
func (c *converter) RemoveBatch(ctx context.Context, batch map[string][]string) error {
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

	return c.dataKeeper.DeleteBatch(ctx, decodedBatch)
}

// PingKeeper checks the data storage connection.
func (c *converter) PingKeeper(ctx context.Context) error {
	return c.dataKeeper.Ping(ctx)
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
