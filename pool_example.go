package main

import (
	"net"
	"time"
	"fmt"
	//"errors"
	"./pool"
)

func tcp_testserver() {
	ln, err := net.Listen("tcp", ":9999")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		go func() {
			fmt.Println("server got:", conn)
			fmt.Fprintf(conn, "hi\r\n")
		}()
	}
}

func newPool() *pool.Pool {
	return &pool.Pool{
		MaxIdle: 3,
		IdleTimeout: 15 * time.Minute,
		Dial: func() (pool.Conn, error) {
			return pool.Dial("tcp", ":9999")
		},
		TestOnBorrow: func(c pool.Conn, t time.Time) error {
			fmt.Println("test on borrow")
			return nil
		},
		Wait: true,
	}
}

func main() {
	done := make(chan bool)

	go tcp_testserver()
	time.Sleep(1 * time.Second)

	pool := newPool()
	if c, err := pool.Get(); err == nil {
		fmt.Println("c", c)
		c.WriteStringLine("hello world")
		bb, err := c.ReadBytesLine()
		if err == nil {
			fmt.Println(string(bb))
		}
		c.Close()
		fmt.Println(c)
	} else {
		fmt.Println(err)
	}
	if c, err := pool.Get(); err == nil {
		fmt.Println("c", c)
		c.WriteStringLine("hello world2")
		bb, err := c.ReadBytesLine()
		if err == nil {
			fmt.Println(string(bb))
		}
		c.Close()
	}else{
		fmt.Println(err)
	}
	fmt.Println(pool)

	<-done
}

