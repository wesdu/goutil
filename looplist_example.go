package main
import (
	"./looplist"
	"fmt"
)

func main() {
	list := looplist.NewLoopList(5)
	list.Append(1)
	list.Append(2)
	list.Append(3)
	list.Append(4)
	list.Append(5)
	list.Append(6)
	list.Append(7)
	list.Append(8)
	list.Append(9)
	list.Append(10)
	list.Append(11)
	list.Append(12)
	for e := list.Back(); e!=nil ; e = e.Prev() {
		fmt.Println(e)
	}
	fmt.Println("---")
	for e := list.Front(); e!=nil ; e = e.Next() {
		fmt.Println(e)
	}
}