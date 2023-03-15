package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ruskiiamov/shortener/internal/url"
)

const (
	defaultNextID  = 1
	fileSavePeriod = 10 * time.Second
)

type memURL struct {
	Original string `json:"original"`
	User     string `json:"user"`
	Deleted  bool   `json:"deleted"`
}

type urlData struct {
	URLs   map[int]memURL `json:"urls"`
	NextID int            `json:"next_id"`
}

type memKeeper struct {
	filePath string
	data     urlData
	mu       sync.RWMutex
}

func newMemKeeper(filePath string) (m *memKeeper, err error) {
	defer func() {
		startPeriodicFileSave(m)
	}()

	if filePath == "" {
		m = &memKeeper{
			data: urlData{
				URLs:   make(map[int]memURL),
				NextID: defaultNextID,
			},
		}
		return m, nil
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer func() {
		e := file.Close()
		if e != nil {
			log.Println(e)
		}
	}()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	if len(fileData) == 0 {
		m = &memKeeper{
			filePath: filePath,
			data: urlData{
				URLs:   make(map[int]memURL),
				NextID: defaultNextID,
			},
		}
		return m, nil
	}

	var data urlData
	err = json.Unmarshal(fileData, &data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse file data: %w", err)
	}

	m = &memKeeper{
		filePath: filePath,
		data:     data,
	}
	return m, nil
}

func startPeriodicFileSave(m *memKeeper) {
	if m == nil || m.filePath == "" {
		return
	}

	t := time.NewTimer(fileSavePeriod)

	go func() {
		for range t.C {
			err := m.saveFile()
			if err != nil {
				log.Println("keeper file save error", err)
			}
			t.Reset(fileSavePeriod)
		}
	}()
}

// Add saves URL for user in memory storage and returns URL id.
func (m *memKeeper) Add(ctx context.Context, userID, original string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	default:
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	matches := m.findMatches([]string{original})
	if len(matches) != 0 {
		id := matches[original]
		return 0, url.NewErrURLDuplicate(id, original)
	}

	id := m.getNextID()

	m.data.URLs[id] = memURL{
		Original: original,
		User:     userID,
	}

	return id, nil
}

// AddBatch saves URL batch for user in memory storage and returns URL IDs.
func (m *memKeeper) AddBatch(ctx context.Context, userID string, originals []string) (map[string]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	added := make(map[string]int, len(originals))

	matches := m.findMatches(originals)

	for _, original := range originals {
		if id, ok := matches[original]; ok {
			added[original] = id
			continue
		}

		id := m.getNextID()
		m.data.URLs[id] = memURL{
			Original: original,
			User:     userID,
		}
		added[original] = id
	}

	return added, nil
}

// Get returns URL by id from memory storage.
func (m *memKeeper) Get(ctx context.Context, id int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	default:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	mURL, ok := m.data.URLs[id]
	if !ok {
		return "", errors.New("wrong id")
	}

	if mURL.Deleted {
		return "", new(url.ErrURLDeleted)
	}

	return mURL.Original, nil
}

// GetAllByUser returns all URL IDs for user from memory storage.
func (m *memKeeper) GetAllByUser(ctx context.Context, userID string) (map[string]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	default:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	urls := make(map[string]int)

	for id, mURL := range m.data.URLs {
		if mURL.User == userID && !mURL.Deleted {
			urls[mURL.Original] = id
		}
	}

	return urls, nil
}

// DeleteBatch deletes URL batch from memory storage.
func (m *memKeeper) DeleteBatch(ctx context.Context, batch map[string][]int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	for userID, IDs := range batch {
		for _, id := range IDs {
			mURL, ok := m.data.URLs[id]
			if !ok {
				continue
			}

			if mURL.User == userID {
				mURL.Deleted = true
				m.data.URLs[id] = mURL
			}
		}
	}

	return nil
}

// Ping always returns error because it is not a DB connection.
func (m *memKeeper) Ping(ctx context.Context) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	return errors.New("memory data keeper is used")
}

// Close dumps all data to the file.
func (m *memKeeper) Close(ctx context.Context) error {
	closed := make(chan error)

	go func() {
		closed <- m.saveFile()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-closed:
			if err != nil {
				return fmt.Errorf("cannot save file: %w", err)
			}
			return nil
		}
	}
}

func (m *memKeeper) getNextID() int {
	id := m.data.NextID
	m.data.NextID++

	return id
}

func (m *memKeeper) saveFile() error {
	if m.filePath == "" {
		return nil
	}

	fileData, err := json.Marshal(m.data)
	if err != nil {
		return fmt.Errorf("JSON encoding error: %w", err)
	}

	file, err := os.OpenFile(m.filePath, os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer func() {
		e:= file.Close()
		if e != nil {
			log.Println(e)
		}
	}()

	_, err = file.Write(fileData)
	if err != nil {
		return fmt.Errorf("cannot save file: %w", err)
	}

	log.Println("keeper file saved")

	return nil
}

func (m *memKeeper) findMatches(originals []string) map[string]int {
	matches := make(map[string]int, len(originals))

	for id, mURL := range m.data.URLs {
		for _, original := range originals {
			if mURL.Original == original {
				matches[original] = id
				break
			}
		}
	}

	return matches
}
