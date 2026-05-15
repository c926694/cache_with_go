package main

import (
	"cache"
	"context"
	"fmt"
	"log"
	"time"
)

func exampleClient() {
	cli, err := cache.NewClient("127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	cli.Set(context.Background(), "k1", []byte("v1"))
	cli.SetWithExpiration(context.Background(), "k2", []byte("v2"), 6*time.Second)
	//time.Sleep(2*time.Second)
	res, err := cli.Get(context.Background(), "k2", func(args ...any) (any, error) {
		return "v2", nil
	}, 0)
	if err != nil {
		log.Printf("get failed :%v", err)
	}
	print(string(res))
}

func print(a any) {
	fmt.Println(a)
}
