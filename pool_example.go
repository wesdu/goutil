package main

import (
	"./pool"
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

var fmt_print_lock = sync.RWMutex{}

func printRed(v interface{}) {
	fmt_print_lock.Lock()
	defer fmt_print_lock.Unlock()
	fmt.Printf("\033[41;37;5m%v\033[0m", v)
}

func printGreen(v interface{}) {
	fmt_print_lock.Lock()
	defer fmt_print_lock.Unlock()
	fmt.Printf("\033[42;37;5m%v\033[0m", v)
}

func printYellow(v interface{}) {
	fmt_print_lock.Lock()
	defer fmt_print_lock.Unlock()
	fmt.Printf("\033[43;37;5m%v\033[0m", v)
}

func printBlue(v interface{}) {
	fmt_print_lock.Lock()
	defer fmt_print_lock.Unlock()
	fmt.Printf("\033[44;37;5m%v\033[0m", v)
}

func tcp_testserver() {
	ln, err := net.Listen("tcp", "0.0.0.0:8888")
	if err != nil {
		fmt.Println("server error", err)
		// handle error
	} else {

	}
	for {
		conn, err := ln.Accept()

		if err != nil {
			printRed(0)
		} else {

			reader := bufio.NewReader(conn)
			writer := bufio.NewWriter(conn)
			go func() {
				for {
					_, _, err := reader.ReadLine()
					if err != nil {
						printRed(1)
						printRed(err)
						break
					}
					//fmt.Println("server got:", string(buf))
					//time.Sleep(50000 * time.Second)
					writer.WriteString("PONG\r\n")
					writer.Flush()
				}
			}()
		}
	}
}

func newPool() *pool.Pool {
	return &pool.Pool{
		MaxActive:   100,
		IdleTimeout: 15 * time.Minute,
		Dial: func() (pool.Conn, error) {
			printBlue(0)
			return pool.DialTimeout("tcp", "0.0.0.0:8888", 5*time.Second, 5*time.Second, 5*time.Second)
		},
		TestOnBorrow: nil,
		Wait:         true,
	}
}

func client_ping(pool *pool.Pool, s string) {
	n := 0
	for {
		if n > 1 {
			//return
		}
		if c, err := pool.Get(); err == nil {
			printGreen(0)
			err := c.WriteStringLine(s)
			if err != nil {
				printRed(3)
				printRed(err)
			}
			_, err = c.ReadBytesLine()
			if err == nil {
				printYellow(0)
			} else {
				printRed(4)
				printRed(err)
			}
			c.Close()
		} else {
			printRed(2)
			printRed(err)
		}
		n ++
		time.Sleep(time.Duration(rand.Intn(10)+1) * time.Second)
	}

}

func main() {
	done := make(chan bool)
	go tcp_testserver()
	time.Sleep(1 * time.Second)
	p := newPool()
	for i := 0; i < 100000; i++ {
		go client_ping(p, "ping"+strconv.Itoa(i))
		//client_ping(p, "PING2")
	}
	<-done
}
