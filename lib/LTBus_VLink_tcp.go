package lib

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const VLink_Clients_Max = 32

var VLink_TCP_connections = make(map[string]net.Conn, VLink_Clients_Max)

func print_VLink_TCP_connections() {
	fmt.Printf("VLink TCP Connections: %d\n", len(VLink_TCP_connections))
	for _, conn := range VLink_TCP_connections {
		fmt.Printf("\t%s\n", conn.RemoteAddr())
	}
}

func tcp_server_loop(tcp_server net.Listener) {
	for {
		conn, err := tcp_server.Accept()
		if errors.Is(err, net.ErrClosed) {
			for _, conn := range VLink_TCP_connections {
				tcp_client_close(conn)
			}
			return
		}

		fmt.Printf("Received VLink TCP Connection: %s\n", conn.RemoteAddr())
		VLink_TCP_connections[conn.RemoteAddr().String()] = conn
		go tcp_client_loop(conn)

		if VLink_Debug {
			print_VLink_TCP_connections()
		}
	}
}

func tcp_client_loop(tcp_client net.Conn) {
	for {
		var rx_buffer [6]byte
		n, err := io.ReadFull(tcp_client, rx_buffer[:])
		if err != nil || n != 6 {
			tcp_client_close(tcp_client)
			break
		}
		VLink_WPkt_Chn <- rx_buffer
	}
	fmt.Printf("VLink TCP Client Loop Stopped: %s\n", tcp_client.RemoteAddr())
}

func tcp_client_close(tcp_client net.Conn) {
	conn_addr := tcp_client.RemoteAddr().String()
	tcp_client.Close()
	delete(VLink_TCP_connections, conn_addr)
	fmt.Printf("Closed VLink TCP Connection: %s\n", conn_addr)
	if VLink_Debug {
		print_VLink_TCP_connections()
	}
}

func LTBus_VLink_Broadcast(packet []byte) {
	for _, conn := range VLink_TCP_connections {
		_, err := conn.Write(packet)
		if err != nil {
			tcp_client_close(conn)
		}
	}
}

func LTBus_VLink_TCP_Init(wg *sync.WaitGroup) {
	vlink_addr := "127.0.0.1:6400"
	fmt.Printf("Starting VLink TCP Server at %s...\n", vlink_addr)
	tcp_server, err := net.Listen("tcp", vlink_addr)
	if err != nil {
		fmt.Printf("Starting VLink TCP Server at %s...ERR\n%s\n", vlink_addr, err)
		os.Exit(1)
	}
	fmt.Printf("Starting VLink TCP Server at %s...OK\n", vlink_addr)

	stop_signal := make(chan os.Signal, 1)
	signal.Notify(stop_signal, os.Interrupt, syscall.SIGINT)

	wg.Go(func() {
		<-stop_signal
		fmt.Printf("Closing VLink TCP Server...\n")
		err := tcp_server.Close()
		if err != nil {
			fmt.Printf("Closing VLink TCP Server...ERR\n")
		} else {
			fmt.Printf("Closing VLink TCP Server...OK\n")
		}
	})

	wg.Go(func() { tcp_server_loop(tcp_server) })
}
