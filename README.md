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

```python
import time
import socket
import struct


vlink_server_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)


def fmt_packet_hex(packet: bytes) -> str:
    return ' '.join([f"0x{x:02X}" for x in packet])


def mmap_decode_f32(mmap: bytes, offset: int) -> float:
    data_bytes = mmap[offset:offset + 4]
    f32_val = struct.unpack('<f', data_bytes)[0]
    return f32_val


def get_latest_mmap(mmap_size: int) -> bytes:
    total_mmap = vlink_server_sock.recv(4096)
    # print(f"Received Total {len(total_mmap)} Bytes")
    return total_mmap[-mmap_size:]


if __name__ == '__main__':
    vlink_server_sock.connect(('127.0.0.1', 6400))
    dts = []

    for i in range(1, 100_0, 10):
        v = i / 10
        wpkt = b'\x00\xD0' + struct.pack('<f', v)
        vlink_server_sock.send(wpkt)

        print('Testing f32:', v)
        t = time.time()
        while True:
            mmap = get_latest_mmap(14)
            echo_v = struct.unpack('<f', mmap[4:8])[0]
            if round(v, 1) == round(echo_v, 1):
                dt = time.time() - t
                dt = round(dt * 1E6, 2)
                print(f"\tRIGHT\t{v} -> {echo_v}\t{dt}us")
                dts.append(dt)
                break
            else:
                wrong_dt = time.time() - t
                wrong_dt = round(wrong_dt * 1E6, 2)
                print(f"\tWRONG\t{v} -> {echo_v}\t{wrong_dt}us")

        time.sleep(1E-2)

    avg_dt = sum(dts) / len(dts)
    print(f"Avg Echo Delay: {avg_dt} us")

# Avg Echo Delay: 350 us
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

### Analyze CPU Utilization
```go
func main() {
	f, _ := os.Create("trace.out")
	trace.Start(f)
	defer trace.Stop()
	
	...
}
```

### View Trace Report
```go
$ go tool trace trace.out
```
