package main

import (
	"./zksnowflake"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

func test1() {
	zksnowflake.Dialog()
	zksnowflake.Setup("127.0.0.1")
	g := zksnowflake.GetGenerator("test1")
	for i := 0; i < 10; i++ {
		go fmt.Println(g.Gen(), i)
	}
	time.Sleep(3 * time.Second)
}

func test2() {
	zksnowflake.Dialog()
	z, _, err := zk.Connect([]string{"127.0.0.1"}, time.Second) //*10)
	if err == nil {
		zksnowflake.Setup(z)
		g := zksnowflake.GetGenerator("test1")
		for i := 0; i < 10; i++ {
			go fmt.Println(g.Gen(), i)
		}
		time.Sleep(3 * time.Second)
	}
}

func main() {
	//test1()
	test2()
}
