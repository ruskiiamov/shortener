package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ruskiiamov/shortener/internal/url"
)

const defaultNextID = 1

type memURL struct {
	Original string   `json:"original"`
	Users    []string `json:"users"`
}

func (m *memURL) hasUser(userID string) bool {
	for _, user := range m.Users {
		if user == userID {
			return true
		}
	}

	return false
}

type URLData struct {
	URLs   map[int]memURL `json:"urls"`
	NextID int            `json:"next_id"`
}

type memKeeper struct {
	filePath string
	data     URLData
}

func newMemKeeper(filePath string) (url.DataKeeper, error) {
	if filePath == "" {
		return &memKeeper{
			data: URLData{
				URLs:   make(map[int]memURL),
				NextID: defaultNextID,
			},
		}, nil
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
		return &memKeeper{
			filePath: filePath,
			data: URLData{
				URLs:   make(map[int]memURL),
				NextID: defaultNextID,
			},
		}, nil
	}

	var data URLData
	err = json.Unmarshal(fileData, &data)
	if err != nil {
		return nil, fmt.Errorf("cannot parse file data: %w", err)
	}

	return &memKeeper{
		filePath: filePath,
		data:     data,
	}, nil
}

func (m *memKeeper) Add(userID, original string) (int, error) {
	matches := m.findMatches([]string{original})
	if len(matches) != 0 {
		id := matches[original]
		m.addUser(id, userID)
		err := m.saveFile()
		if err != nil {
			return 0, fmt.Errorf("cannot save file: %w", err)
		}
		return 0, url.NewErrURLDuplicate(id, original)
	}

	id := m.getNextID()

	m.data.URLs[id] = memURL{
		Original: original,
		Users:    []string{userID},
	}

	err := m.saveFile()
	if err != nil {
		return 0, fmt.Errorf("cannot save file: %w", err)
	}

	return id, nil
}

func (m *memKeeper) AddBatch(userID string, originals []string) (map[string]int, error) {
	added := make(map[string]int)

	matches := m.findMatches(originals)

	for _, original := range originals {
		if id, ok := matches[original]; ok {
			m.addUser(id, userID)
			added[original] = id
			continue
		}

		id := m.getNextID()
		m.data.URLs[id] = memURL{
			Original: original,
			Users:    []string{userID},
		}
		added[original] = id
	}

	err := m.saveFile()
	if err != nil {
		return nil, fmt.Errorf("cannot save file: %w", err)
	}

	return added, nil
}

func (m *memKeeper) Get(id int) (string, error) {
	memURL, ok := m.data.URLs[id]
	if !ok {
		return "", errors.New("wrong id")
	}

	return memURL.Original, nil
}

func (m *memKeeper) GetAllByUser(userID string) (map[string]int, error) {
	urls := make(map[string]int)

	for id, memURL := range m.data.URLs {
		if memURL.hasUser(userID) {
			urls[memURL.Original] = id
		}
	}

	return urls, nil
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

func (m *memKeeper) addUser(id int, userID string) {
	memURL := m.data.URLs[id]

	if memURL.hasUser(userID) {
		return
	}

	memURL.Users = append(memURL.Users, userID)
	m.data.URLs[id] = memURL
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

	file, err := os.OpenFile(m.filePath, os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	_, err = file.Write(fileData)
	if err != nil {
		return fmt.Errorf("cannot save file: %w", err)
	}

	return nil
}

func (m *memKeeper) findMatches(originals []string) map[string]int {
	matches := make(map[string]int)

	for id, memURL := range m.data.URLs {
		for _, original := range originals {
			if memURL.Original == original {
				matches[original] = id
				break
			}
		}
	}

	return matches
}
