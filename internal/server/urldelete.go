package server

import (
	"context"
	"log"
	"time"
)

const deletePeriod = 10 * time.Second

type delBatch struct {
	userID     string
	encodedIDs []string
}

func (h *handler) startDeleteURL(ctx context.Context) {
	defer close(h.delFinish)

	buf := make(map[string][]string)

	defer func() {
		onCloseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := h.urlConverter.RemoveBatch(onCloseCtx, buf)
		if err != nil {
			log.Printf("on close delete URL batch error: %v\n", err)
			return
		}
		log.Printf("URL batch deleted on close\n")
	}()

	t := time.NewTimer(deletePeriod)
	for {
		select {
		case batch, ok := <-h.delBuf:
			if !ok {
				return
			}
			if URLs, ok := buf[batch.userID]; ok {
				buf[batch.userID] = unique(URLs, batch.encodedIDs)
			} else {
				buf[batch.userID] = unique(batch.encodedIDs)
			}
		case <-t.C:
			err := h.urlConverter.RemoveBatch(ctx, buf)
			t.Reset(deletePeriod)
			if err != nil {
				log.Printf("delete URL batch error: %v\n", err)
				continue
			}
			log.Printf("URL batch deleted\n")

			buf = make(map[string][]string)
		}
	}
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
