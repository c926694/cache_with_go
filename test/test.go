package main

import (
	"cache"
	"context"
	"fmt"
)

func main() {
	s:=cache.NewServer(1024)
	go s.Start()
	cli,err:=cache.NewClient("127.0.0.1:8080")
	if err!=nil {
		panic(err)
	}
	go func() {
		for i:=0;i<100;i++ {
			key:="key:1"
			value:=[]byte("value:1")
			cli.Set(context.Background(),key,value)
		}
	}()
	go func() {
		for i:=0;i<100;i++ {
			key:="key:1"
			cli.Delete(context.Background(),key)
			//print("deleted")
		}
	}()
	select {

	}
}

func print(a any) {
	fmt.Println(a)
}
