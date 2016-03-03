package main
import (
	"./zksnowflake"
	"time"
	"fmt"
)

func main() {
	zksnowflake.Dialog()
	zksnowflake.Setup("127.0.0.1")
	g := zksnowflake.GetGenerator("test1")
	for i:=0;i<1000000;i++ {
		go fmt.Println(g.Gen(), i)
	}
	time.Sleep(1*time.Second)
}