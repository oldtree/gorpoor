package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancelf := context.WithTimeout(context.Background(), time.Second*5)
	defer cancelf()
	go func(ctx context.Context) {
		dead, ok := ctx.Deadline()
		fmt.Println(dead, ok)
	}(ctx)
	select {
	case <-ctx.Done():
		fmt.Println("task done")
	case <-time.After(time.Second * 5):
		fmt.Println("task timeout")
	}
	fmt.Println("over....")
}
