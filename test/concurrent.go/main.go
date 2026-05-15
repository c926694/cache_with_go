package main

import (
	"cache"
	"context"
	"fmt"
	"log"
)

func main() {
	endpoints := []string{"127.0.0.1:2379"}
	cli,err := cache.NewEtcdClusterClient(context.Background(), 1, endpoints, "cache")
	if err != nil {
		panic(err)
	}
	go func () {
		for i:=0;i<100;i++ {
			key:=fmt.Sprintf("user:%d",i)
			value:=[]byte(fmt.Sprintf("value:%d",i))
			cli.Set(context.Background(),key,value)
		}
	}()
	go func () {
		for i:=0;i<100;i++ {
			key:=fmt.Sprintf("user:%d",i)
			data,err:=cli.Get(context.Background(),key,nil,0)
			if err!= nil {
				log.Printf("get failed : %v",err)
			}
			fmt.Println(string(data))
		}
	}()
	select{}
}