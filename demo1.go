package main

import "fmt"

/*
 * @Description: In User Settings Edit
 * @Author: your name
 * @LastEditors: Please set LastEditors
 * @Date: 2019-02-25 13:58:08
 * @LastEditTime: 2019-02-25 15:46:47
 */
func Count(ch chan int) {
	ch <- 1
	fmt.Println(ch)
}

func main() {
	chs := make([]chan int, 10)
	for i := 0; i < 10; i++ {
		chs[i] = make(chan int)
		// fmt.Println(chs[i])
		go Count(chs[i])
	}
	for _, ch := range chs {
		<-ch
	}
}
