package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/ruskiiamov/shortener/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewShortenerClient(conn)

	var header metadata.MD

	resp1, _ := c.AddURL(context.Background(), &pb.AddURLRequest{Url: "http://yandex.ru"}, grpc.Header(&header))
	fmt.Printf("RESP1: %s\n", resp1.Id)

	token := header.Get("auth")[0]
	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("auth", token),
	)

	resp2, _ := c.AddURL(ctx, &pb.AddURLRequest{Url: "http://google.com"})
	fmt.Printf("RESP2: %s\n", resp2.Id)

	resp3, err := c.GetURL(context.Background(), &pb.GetURLRequest{Id: "5"})
	fmt.Printf("RESP3: %s\n", resp3)
	fmt.Println(err)

	resp4, _ := c.GetURL(context.Background(), &pb.GetURLRequest{Id: "1"})
	fmt.Printf("RESP4: %s\n", resp4)

	resp5, _ := c.AddURLBatch(ctx, &pb.AddURLBatchRequest{Urls: []*pb.AddURLBatchRequestItem{
		{CorrelationId: "123", Url: "http://some-url1.com"},
		{CorrelationId: "456", Url: "http://some-url2.com"},
	}})
	fmt.Printf("RESP5: %s\n", resp5)

	resp6, _ := c.GetAllURL(ctx, &pb.GetAllURLRequest{})
	fmt.Printf("RESP6: %s\n", resp6.Urls)

	resp7, _ := c.DeleteURLBatch(ctx, &pb.DeleteURLBatchRequest{Ids: []string{"1", "4"}})
	fmt.Printf("RESP7: %s\n", resp7)

	resp8, _ := c.GetStats(ctx, &pb.GetStatsRequest{})
	fmt.Printf("RESP8: %s\n", resp8)

	resp9, _ := c.PingDB(ctx, &pb.PingDBRequest{})
	fmt.Printf("RESP9: %s\n", resp9)

	time.Sleep(12 * time.Second)

	resp10, _ := c.GetAllURL(ctx, &pb.GetAllURLRequest{})
	fmt.Printf("RESP10: %s\n", resp10.Urls)

	fmt.Println("DONE!!!")
}
