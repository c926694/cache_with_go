package main

import (
	"cache"
	"context"
	"fmt"
)

func main() {

	s1:=cache.NewServer(1024,"127.0.0.1:3000")
	s2:=cache.NewServer(1024,"127.0.0.1:3001")
	go s1.Start()
	go s2.Start()
	nodes := []string{"127.0.0.1:3000","127.0.0.1:3001"}
	cli,err:=cache.NewClusterClient(60,nodes...)
	if err!=nil {
		fmt.Printf("failed to create cluster client: %v",err)
		return
	}
	ctx:=context.Background()
	cli.Set(ctx,"k1",[]byte("v1"))
	cli.Set(ctx,"k2",[]byte("v2"))
	res,err:=cli.Get(ctx,"k2")
	if err!=nil {
		fmt.Printf("get failed :%v",err)
	}
	print(string(res))
	select {}
}