package main

import (
	"cache"
	"context"
	"encoding/json"
	"fmt"
)

func main() {
	cli, err := cache.NewClusterClient(1, "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	var getter= func (args ...any) (any, error) {
		return getUserByID(args[0].(int))
	}
	user,err:=cli.Get(context.Background(),"k1",getter,0,6)
	if err!=nil {
		panic(err)
	}
	print(user)
	user,err=cli.Get(context.Background(),"k1",getter,0,6)
	if err!=nil {
		panic(err)
	}
	var u User
	json.Unmarshal(user,&u)
	fmt.Printf("%v",u)
}

func getUserByID(id int) (User,error) {
	return  User{Id:id,Username:"admin",Password:"admin"},nil
}

type User struct {
	Id int
	Username string
	Password string
}