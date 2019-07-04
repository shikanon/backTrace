package main

import (
	"fmt"
	"github.com/shikanon/backTrace"
)

func main() {
	//backTrace.RunBacktrace()

	a := []string{"123"}

	for _, val := range a[0:0] {
		fmt.Println(val)
	}

	t1 := backTrace.Task{Code: "001", CodeIndex: 1, BuyIndex: 1, SellIndex: 1}
	t2 := backTrace.Task{Code: "002", CodeIndex: 1, BuyIndex: 1, SellIndex: 2}
	t3 := backTrace.Task{Code: "003", CodeIndex: 1, BuyIndex: 2, SellIndex: 1}
	i1 := backTrace.IndexNode{Key: "001", T: &t1}
	i2 := backTrace.IndexNode{Key: "002", T: &t2}
	i3 := backTrace.IndexNode{Key: "003", T: &t3}

	queue := backTrace.IndexQueue{}

	queue.Insert(&i3)
	queue.Insert(&i1)
	queue.Insert(&i2)

	head := queue.Head

	for {
		if head == nil {
			break
		}
		fmt.Println(head.Key + "\n")
		head = head.Next

	}

}
