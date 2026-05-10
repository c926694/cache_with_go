package main

import (
	"cache"
	"context"
	"fmt"
	"time"
)

func main() {
	s:=cache.NewServer(1024)
	go s.Start()
	cli,err:=cache.NewClient("127.0.0.1:8080")
	if err!=nil {
		panic(err)
	}
	cli.Set(context.Background(),"k1",[]byte("v1"))
	cli.SetWithExpiration(context.Background(),"k2",[]byte("v2"),2*time.Second)
	time.Sleep(2*time.Second)
	res,_:=cli.Get(context.Background(),"k2")
	print(string(res))
	select {}
}

func print(a any) {
	fmt.Println(a)
}
