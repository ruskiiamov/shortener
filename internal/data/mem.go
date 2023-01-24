package data

import (
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

type URLData struct {
	URLs   map[int]memURL `json:"urls"`
	NextID int            `json:"next_id"`
}

type memKeeper struct {
	filePath string
	data     URLData
	mu       sync.RWMutex
}

func newMemKeeper(filePath string) (m *memKeeper, err error) {
	defer func() {
		startPeriodicFileSave(m)
	}()

	if filePath == "" {
		m = &memKeeper{
			data: URLData{
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
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	if len(fileData) == 0 {
		m = &memKeeper{
			filePath: filePath,
			data: URLData{
				URLs:   make(map[int]memURL),
				NextID: defaultNextID,
			},
		}
		return m, nil
	}

	var data URLData
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

	go func() {
		for {
			time.Sleep(fileSavePeriod)
			err := m.saveFile()
			if err != nil {
				log.Println("keeper file save error", err)
			}
		}
	}()
}

func (m *memKeeper) Add(userID, original string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *memKeeper) AddBatch(userID string, originals []string) (map[string]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	added := make(map[string]int)

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

func (m *memKeeper) Get(id int) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mURL, ok := m.data.URLs[id]
	if !ok {
		return "", errors.New("wrong id")
	}

	if mURL.Deleted {
		return "", new(url.ErrURLDeleted)
	}

	return mURL.Original, nil
}

func (m *memKeeper) GetAllByUser(userID string) (map[string]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	urls := make(map[string]int)

	for id, mURL := range m.data.URLs {
		if mURL.User == userID && !mURL.Deleted {
			urls[mURL.Original] = id
		}
	}

	return urls, nil
}

func (m *memKeeper) DeleteBatch(batch map[string][]int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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

func (m *memKeeper) Ping() error {
	return errors.New("memory data keeper is used")
}

func (m *memKeeper) Close() error {
	err := m.saveFile()
	if err != nil {
		return fmt.Errorf("cannot save file: %w", err)
	}
	return nil
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
	defer file.Close()

	_, err = file.Write(fileData)
	if err != nil {
		return fmt.Errorf("cannot save file: %w", err)
	}

	log.Println("keeper file saved")

	return nil
}

func (m *memKeeper) findMatches(originals []string) map[string]int {
	matches := make(map[string]int)

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
