package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

const remoteAddr = "192.168.1.120:37777"

func proxyConn(conn *net.TCPConn) {
	rAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		panic(err)
	}

	rConn, err := net.DialTCP("tcp", nil, rAddr)
	if err != nil {
		panic(err)
	}
	defer rConn.Close()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024*100)
			n, err := conn.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}

			if n == 0 {
				continue
			}

			log.Println("phone->cam", string(buf[:n]))

			_, err = rConn.Write(buf[:n])
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			buf := make([]byte, 1024*100)
			n, err := rConn.Read(buf)
			if err != nil {
				log.Println(err)
				return
			}

			log.Println("cam->phone", string(buf[:n]))

			_, err = conn.Write(buf[:n])
			if err != nil {
				log.Println(err)
				return
			}
		}
	}()

	wg.Wait()
}

func handleConn(in <-chan *net.TCPConn, out chan<- *net.TCPConn) {
	for conn := range in {
		proxyConn(conn)
		out <- conn
	}
}

func closeConn(in <-chan *net.TCPConn) {
	for conn := range in {
		conn.Close()
	}
}

func main() {
	flag.Parse()

	fmt.Printf("Proxying: %v\n\n", remoteAddr)

	addr, err := net.ResolveTCPAddr("tcp", "192.168.1.2:37777")
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}

	pending, complete := make(chan *net.TCPConn), make(chan *net.TCPConn)

	for i := 0; i < 5; i++ {
		go handleConn(pending, complete)
	}
	go closeConn(complete)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}
		pending <- conn
	}
}
