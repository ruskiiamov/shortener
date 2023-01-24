package server

import (
	"log"
	"sync"
	"time"
)

const deletePeriod = 10 * time.Second

type delBatch struct {
	userID     string
	encodedIDs []string
}

func (h *handler) startDeleteURL(w *sync.WaitGroup) {
	w.Add(1)

	delCh := make(chan struct{})
	go func() {
		for {
			time.Sleep(deletePeriod)
			delCh <- struct{}{}
		}
	}()

	go func() {
		buf := make(map[string][]string)

	loop:
		for {
			select {
			case batch, ok := <-h.delBuf:
				if !ok {
					break loop
				}
				if URLs, ok := buf[batch.userID]; ok {
					buf[batch.userID] = unique(URLs, batch.encodedIDs)
				} else {
					buf[batch.userID] = unique(batch.encodedIDs)
				}
			case <-delCh:
				err := h.urlConverter.RemoveBatch(buf)
				if err != nil {
					log.Printf("delete URL batch error: %v\n", err)
					continue
				}
				log.Printf("URL batch deleted\n")

				buf = make(map[string][]string)
			}
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
