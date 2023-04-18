package url

import (
	"context"
	"log"
	"time"
)

const deletePeriod = 10 * time.Second

// DelBatch is the item for delete buffer.
type DelBatch struct {
	UserID     string
	EncodedIDs []string
}

// StartDeleteURL starts goroutine to periodic deleting URLs and returns
// the channel to receive buffer items. To stop deleting it is needed to
// close the channel.
func StartDeleteURL(ctx context.Context, c Converter) chan *DelBatch {
	delBuf := make(chan *DelBatch)

	go deleteURL(ctx, delBuf, c)

	return delBuf
}

func deleteURL(ctx context.Context, delBuf chan *DelBatch, c Converter) {
	buf := make(map[string][]string)

	defer func() {
		onCloseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := c.RemoveBatch(onCloseCtx, buf)
		if err != nil {
			log.Printf("on close delete URL batch error: %v\n", err)
			return
		}
		log.Printf("URL batch deleted on close\n")
	}()

	t := time.NewTimer(deletePeriod)
	for {
		select {
		case batch, ok := <-delBuf:
			if !ok {
				return
			}
			if URLs, ok := buf[batch.UserID]; ok {
				buf[batch.UserID] = unq(URLs, batch.EncodedIDs)
			} else {
				buf[batch.UserID] = unq(batch.EncodedIDs)
			}
		case <-t.C:
			err := c.RemoveBatch(ctx, buf)
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

func unq(URLs ...[]string) []string {
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
