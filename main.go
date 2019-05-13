package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
)

func connChannel(conn net.Conn) chan []byte {
	channel := make(chan []byte)

	go func() {
		buf := make([]byte, 1024)
		defer close(channel)

		for {
			read, err := conn.Read(buf)
			if err != nil {
				break
			}

			if read > 0 {
				data := make([]byte, read)
				copy(data, buf[0:read])
				channel <- data
			}
		}
	}()

	return channel
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	dest := GetConfiguration().Destination
	destConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", dest.Host, dest.Port))
	if err != nil {
		log.Printf("Cannot connect to destination: %v", err)
		return
	}
	defer destConn.Close()

	sourceChan := connChannel(conn)
	destChan := connChannel(destConn)

Loop:
	for {
		var data []byte
		open := true

		select {
		case data, open = <-sourceChan:
			if !open {
				break Loop
			}
			_, err = destConn.Write(data)

		case data, open = <-destChan:
			if !open {
				break Loop
			}
			_, err = conn.Write(data)
		}
	}
}

func Listener(source Address, signal chan bool) {

	addr := fmt.Sprintf("%s:%d", source.Host, source.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("Error binding to address %s: %v", addr, err)
		return
	}

	log.Printf("Listening on %s", addr)

	connChan := make(chan net.Conn, 0)
	defer close(connChan)
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			connChan <- conn
		}
	}()

	for {
		select {
		case conn := <-connChan:
			go HandleConnection(conn)
		case _ = <-signal:
			_ = listener.Close()
			return
		}
	}
}

func main() {
	config := GetConfiguration()

	notifiers := make([]*chan bool, 0)

	for _, host := range config.BindAddresses {
		signal := make(chan bool, 1)
		go Listener(host, signal)
		notifiers = append(notifiers, &signal)
	}

	systemSignal := make(chan os.Signal, 1)
	for true {
		signal := <-systemSignal
		if signal == syscall.SIGINT {
			break
		}
	}

	for _, notifier := range notifiers {
		*notifier <- true
	}
}
