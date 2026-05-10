package main

import (
	"cache"
	"fmt"
)

func main() {
	c := cache.NewCache(10)
	bw:=cache.NewByteView([]byte("111"))
	c.Set("k",bw)

	if v,ok:=c.Get("k");ok{
		print(string(v.ByteSlice()))
	}
}

func print(a any) {
	fmt.Println(a)
}
