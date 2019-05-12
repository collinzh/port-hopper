package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"syscall"
)

func HandleConnection(conn *net.Conn) {
	dest := GetConfiguration().Destination

	destConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", dest.Host, dest.Port))
	defer destConn.Close()

	if err != nil {

	}

}

func Listener(source Address, signal chan bool) {

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", source.Host, source.Port))
	if err != nil {
		log.Printf("Error binding to address %s:%d", source.Host, source.Port)
		return
	}

	for true {
		connChan := make(chan *net.Conn, 0)
		go func() {
			for true {
				conn, err := listener.Accept()
				if err != nil {
					continue
				}
				connChan <- &conn
			}
		}()

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
