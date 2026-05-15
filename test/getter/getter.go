package main

import (
	"cache"
	"context"
	"encoding/json"
	"fmt"
)

func main() {
	endpoints:=[]string{"127.0.0.1:2379"}
	cli, err := cache.NewEtcdClusterClient(context.Background(),1,endpoints,"cache")
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