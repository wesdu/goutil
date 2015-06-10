package main

import (
	"fmt"
	"net"
	"time"
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
			for {
				buf := make([]byte, 2048)
				_, err := conn.Read(buf)
				if err != nil {
					fmt.Println("read from client err:", err)
					break
				}
				fmt.Println("server got:", string(buf))
				time.Sleep(50000 * time.Second)
				fmt.Fprintf(conn, "PONG\r\n")
			}
		}()
	}
}

func newPool() *pool.Pool {
	return &pool.Pool{
		MaxIdle:     3,
		IdleTimeout: 15 * time.Minute,
		Dial: func() (pool.Conn, error) {
			return pool.DialTimeout("tcp", ":9999", 5*time.Second, 5*time.Second, 5*time.Second)
		},
		TestOnBorrow: func(c pool.Conn, t time.Time) error {
			return nil
		},
		Wait: true,
	}
}

func client_ping(pool *pool.Pool, s string) {
	if c, err := pool.Get(); err == nil {
		fmt.Println("pool.Get()", c)
		err := c.WriteStringLine(s)
		if err != nil {
			fmt.Print("write to server error:", err)
		}
		fmt.Println("write to server:", s)
		bb, err := c.ReadBytesLine()
		if err == nil {
			fmt.Println("get from server:", string(bb))
		}
		c.Close()
	} else {
		fmt.Println("pool.Get() err:", err)
	}
}


func main() {
	done := make(chan bool)
	go tcp_testserver()
	time.Sleep(1 * time.Second)
	p := newPool()
	client_ping(p, "PING1")
	client_ping(p, "PING2")
	<-done
}
