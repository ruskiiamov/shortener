package server

import (
	"log"
	"sync"
)

const threshold = 5

type delBatch struct {
	userID     string
	encodedIDs []string
}

func (h *handler) startDeleteURL(w *sync.WaitGroup) {
	w.Add(1)

	go func() {
		buf := make(map[string][]string)
		count := 0

		for batch := range h.delBuf {
			if URLs, ok := buf[batch.userID]; ok {
				buf[batch.userID] = unique(URLs, batch.encodedIDs)
			} else {
				buf[batch.userID] = unique(batch.encodedIDs)
			}

			count += len(batch.encodedIDs)
			if count < threshold {
				continue
			}

			err := h.urlConverter.RemoveBatch(buf)
			if err != nil {
				log.Printf("delete URL batch error: %v\n", err)
				continue
			}
			log.Printf("URL batch deleted\n")

			buf = make(map[string][]string)
			count = 0
		}

		err := h.urlConverter.RemoveBatch(buf)
		if err != nil {
			log.Printf("delete URL batch error: %v\n", err)
		}
		log.Printf("URL batch deleted on close\n")

		w.Done()
	}()
}

func unique(URLs ...[]string) []string {
	m := make(map[string]bool)

	unique := make([]string, 0)

	for _, u := range URLs {
		for _, url := range u {
			if _, ok := m[url]; !ok {
				m[url] = true
				unique = append(unique, url)
			}
		}
	}

	return unique
}
