package main

import (
	"cache"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	cli,err:=cache.NewClient("127.0.0.1:3000")
	if err!=nil {
		panic(err)
	}
	cli.Set(context.Background(),"k1",[]byte("v1"))
	cli.SetWithExpiration(context.Background(),"k2",nil,6*time.Second)
	//time.Sleep(2*time.Second)
	res,err:=cli.Get(context.Background(),"k2")
	if err!=nil {
		log.Printf("get failed :%v",err)
	}
	print(string(res))
	select {}
}

func print(a any) {
	fmt.Println(a)
}
