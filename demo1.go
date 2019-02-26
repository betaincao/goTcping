package main

import "fmt"

/*
 * @Description: In User Settings Edit
 * @Author: your name
 * @LastEditors: Please set LastEditors
 * @Date: 2019-02-25 13:58:08
 * @LastEditTime: 2019-02-26 15:50:20
 */
func Add(x, y int) {
	z := x + y
	fmt.Println(z)
}

func main() {
	chs := make([]chan int, 10)

	for i := 0; i < 10; i++ {
		chs[i] = make(chan int)
		go Add(i, i)
	}
}
