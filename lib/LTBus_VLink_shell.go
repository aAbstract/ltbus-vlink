package lib

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	lua "github.com/yuin/gopher-lua"
)

const VLink_Shell_Clients_Max = 32

var VLink_Shell_connections = make(map[string]net.Conn, VLink_Shell_Clients_Max)

func print_VLink_Shell_connections() {
	fmt.Printf("VLink Shell Connections: %d\n", len(VLink_Shell_connections))
	for _, conn := range VLink_Shell_connections {
		fmt.Printf("\t%s\n", conn.RemoteAddr())
	}
}

func init_lua_vm(shell_client net.Conn, lua_vm *lua.LState) {
	lua_vm.SetGlobal("VLink_print", lua_vm.NewFunction(func(L *lua.LState) int {
		text := L.ToString(1)
		shell_client.Write([]byte(text + "\n"))
		return 0
	}))
}

func shell_client_loop(shell_client net.Conn) {
	fmt.Printf("Spawning New VLink Shell: %s...\n", shell_client.RemoteAddr())
	Lua_VM := lua.NewState()
	init_lua_vm(shell_client, Lua_VM)
	reader := bufio.NewReader(shell_client)
	fmt.Printf("Spawning New VLink Shell: %s...OK\n", shell_client.RemoteAddr())

	for {
		shell_client.Write([]byte("vlink (local)> "))
		cmd, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		lua_err := Lua_VM.DoString(cmd[:len(cmd)-1]) // remove last \n from cmd
		if lua_err != nil {
			shell_client.Write([]byte(lua_err.Error()))
			continue
		}
		shell_client.Write([]byte(Lua_VM.Get(-1).String() + "\n"))
	}
	shell_client_close(shell_client)
	Lua_VM.Close()
	fmt.Printf("VLink Shell Client Loop Stopped: %s\n", shell_client.RemoteAddr())
}

func shell_client_close(shell_client net.Conn) {
	conn_addr := shell_client.RemoteAddr().String()
	shell_client.Close()
	delete(VLink_Shell_connections, conn_addr)
	fmt.Printf("Closed VLink Shell Connection: %s\n", conn_addr)
	if VLink_Debug {
		print_VLink_Shell_connections()
	}
}

func shell_server_loop(shell_server net.Listener) {
	for {
		conn, err := shell_server.Accept()
		if errors.Is(err, net.ErrClosed) {
			for _, conn := range VLink_Shell_connections {
				shell_client_close(conn)
			}
			return
		}

		fmt.Printf("Received VLink Shell Connection: %s\n", conn.RemoteAddr())
		VLink_Shell_connections[conn.RemoteAddr().String()] = conn
		go shell_client_loop(conn)

		if VLink_Debug {
			print_VLink_Shell_connections()
		}
	}
}

func LTBus_VLink_Shell_Init(wg *sync.WaitGroup) {
	shell_addr := "127.0.0.1:6401"
	fmt.Printf("Starting VLink Shell Server at %s...\n", shell_addr)
	shell_server, err := net.Listen("tcp", shell_addr)
	if err != nil {
		fmt.Printf("Starting VLink Shell Server at %s...ERR\n%s\n", shell_addr, err)
		os.Exit(1)
	}
	fmt.Printf("Starting VLink Shell Server at %s...OK\n", shell_addr)

	stop_signal := make(chan os.Signal, 1)
	signal.Notify(stop_signal, os.Interrupt, syscall.SIGINT)

	wg.Go(func() {
		<-stop_signal
		fmt.Printf("Closing VLink Shell Server...\n")
		err := shell_server.Close()
		if err != nil {
			fmt.Printf("Closing VLink Shell Server...ERR\n")
		} else {
			fmt.Printf("Closing VLink Shell Server...OK\n")
		}
	})

	wg.Go(func() { shell_server_loop(shell_server) })
}
