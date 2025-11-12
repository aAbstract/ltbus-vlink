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
		fmt.Printf("Delat_T: %dus\n", dt.Microseconds())

		time.Sleep(1 * time.Millisecond)
	}
}
// Delta_T: ~250us / 4KHz
```
