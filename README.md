# LTBus_VLink

LTBus High Speed Data Link

### Benchmarks
```go
func LTBus_VLink_Device_Loop() {
	for {
		start := time.Now()

		data_size := packet_size - 10
		ltbrr_mmap := LTBus_Read_Request(0xD000, uint16(data_size))
		serial_port.Write(ltbrr_mmap)

		rx_buffer := make([]byte, packet_size)
		n_bytes, err := serial_port.Read(rx_buffer)
		if err != nil || n_bytes != packet_size {
			continue
		}

		if rx_buffer[0] != 0x7B || rx_buffer[packet_size-1] != 0x7D {
			fmt.Printf("Invalid LTBus Packet Frame\n")
			continue
		}

		if !CheckCRC(rx_buffer) {
			fmt.Printf("Invalid LTBus Packet CRC\n")
			continue
		}

		dt := time.Since(start)
		fmt.Printf("Delta_T: %dus\n", dt.Microseconds())

		time.Sleep(1 * time.Millisecond)
	}
}
// Delta_T: ~250us / 4KHz
```

### Analyze Heap Allocations
```bash
$ python scripts/heap_diag.py

lib/LTBus_VLink_tcp.go:16:33: make(map[string]net.Conn, 32) escapes to heap -> var VLink_TCP_connections = make(map[string]net.Conn, VLink_Clients_Max)
lib/LTBus_VLink_device.go:19:23: make([][6]byte, 32768) escapes to heap -> var VLink_WPool = make([][6]byte, VLink_Clients_Max*1024)
lib/LTBus_VLink_tcp.go:90:8: func literal escapes to heap -> wg.Go(func() {
lib/LTBus_VLink_tcp.go:101:8: func literal escapes to heap -> wg.Go(func() { tcp_server_loop(tcp_server) })
lib/LTBus_VLink_device.go:23:23: append escapes to heap -> VLink_WPool = append(VLink_WPool, WPkt)
lib/LTBus_VLink_device.go:71:19: make([]byte, packet_size) escapes to heap -> rx_buffer := make([]byte, packet_size)
lib/LTBus_VLink_device.go:136:8: func literal escapes to heap -> wg.Go(func() { LTBus_VLink_Device_Loop(serial_port, *packet_size, stop_signal) })
```
