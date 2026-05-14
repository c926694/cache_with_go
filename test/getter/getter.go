package main

import (
	"cache"
	"context"
)

func main() {
	cli, err := cache.NewClusterClient(1, "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	var getter= func (args ...any) (any, error) {
		return getUserByID(args[0].(int))
	}
	user,err:=cli.Get(context.Background(),"k1",getter,6)
	if err!=nil {
		panic(err)
	}
	print(user)
	user,err=cli.Get(context.Background(),"k1",getter,6)
	if err!=nil {
		panic(err)
	}
	print(user)
}

func getUserByID(id int) (User,error) {
	return  User{id:id,Username:"admin",Password:"admin"},nil
}

type User struct {
	id int
	Username string
	Password string
}