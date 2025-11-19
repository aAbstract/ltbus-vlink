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
	"time"
)

const VLink_Clients_Max = 32

type VLinkConnection struct {
	VLink_Speed_ms int
	VLink_Tick_ms  int64
	Conn           net.Conn
}

var VLink_TCP_connections = make(map[string]VLinkConnection, VLink_Clients_Max)

func print_VLink_TCP_connections() {
	fmt.Printf("VLink TCP Connections: %d\n", len(VLink_TCP_connections))
	for _, vlink_conn := range VLink_TCP_connections {
		fmt.Printf("\t%+v\n", vlink_conn)
	}
}

func tcp_server_loop(tcp_server net.Listener, vlink_speed_ms int) {
	for {
		conn, err := tcp_server.Accept()
		if errors.Is(err, net.ErrClosed) {
			for _, vlink_conn := range VLink_TCP_connections {
				tcp_client_close(vlink_conn)
			}
			return
		}

		vlink_conn := VLinkConnection{VLink_Speed_ms: vlink_speed_ms, VLink_Tick_ms: 0, Conn: conn}
		fmt.Printf("Received VLink TCP Connection: %+v\n", vlink_conn)
		VLink_TCP_connections[conn.RemoteAddr().String()] = vlink_conn
		go tcp_client_loop(vlink_conn)

		if VLink_Debug {
			print_VLink_TCP_connections()
		}
	}
}

func tcp_client_loop(vlink_conn VLinkConnection) {
	for {
		var rx_buffer [6]byte
		n, err := io.ReadFull(vlink_conn.Conn, rx_buffer[:])
		if err != nil || n != 6 {
			tcp_client_close(vlink_conn)
			break
		}
		VLink_WPkt_Chn <- rx_buffer
	}
	fmt.Printf("VLink TCP Client Loop Stopped: %s\n", vlink_conn.Conn.RemoteAddr())
}

func tcp_client_close(vlink_conn VLinkConnection) {
	conn_addr := vlink_conn.Conn.RemoteAddr().String()
	vlink_conn.Conn.Close()
	delete(VLink_TCP_connections, conn_addr)
	fmt.Printf("Closed VLink TCP Connection: %+v\n", vlink_conn)
	if VLink_Debug {
		print_VLink_TCP_connections()
	}
}

func LTBus_VLink_Broadcast(packet []byte) {
	for _, vlink_conn := range VLink_TCP_connections {
		if vlink_conn.VLink_Speed_ms == 0 {
			_, err := vlink_conn.Conn.Write(packet)
			if err != nil {
				tcp_client_close(vlink_conn)
			}
			continue
		}

		now_ms := time.Now().UnixMilli()
		if (now_ms - vlink_conn.VLink_Tick_ms) > int64(vlink_conn.VLink_Speed_ms) {
			_, err := vlink_conn.Conn.Write(packet)
			if err != nil {
				tcp_client_close(vlink_conn)
				continue
			}
			vlink_conn.VLink_Tick_ms = now_ms
		}
	}
}

func create_VLink_Listener(wg *sync.WaitGroup, addr string, vlink_speed_ms int) net.Listener {
	fmt.Printf("Starting VLink TCP Server at %s, Speed: %dms...\n", addr, vlink_speed_ms)
	vlink_listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Starting VLink TCP Server at %s, Speed: %dms...ERR\n%s\n", addr, vlink_speed_ms, err)
		os.Exit(1)
	}
	fmt.Printf("Starting VLink TCP Server at %s, Speed: %dms...OK\n", addr, vlink_speed_ms)
	wg.Go(func() { tcp_server_loop(vlink_listener, vlink_speed_ms) })
	return vlink_listener
}

func LTBus_VLink_TCP_Init(wg *sync.WaitGroup) {
	vlink_listener_0ms := create_VLink_Listener(wg, "127.0.0.1:6400", 0)
	vlink_listener_10ms := create_VLink_Listener(wg, "127.0.0.1:6401", 10)

	stop_signal := make(chan os.Signal, 1)
	signal.Notify(stop_signal, os.Interrupt, syscall.SIGINT)

	wg.Go(func() {
		<-stop_signal
		fmt.Printf("Closing VLink TCP Servers...\n")
		vlink_listener_0ms.Close()
		vlink_listener_10ms.Close()
		fmt.Printf("Closing VLink TCP Servers...OK\n")
	})
}
