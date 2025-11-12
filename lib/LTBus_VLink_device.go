package lib

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tarm/serial"
)

var VLink_Debug bool

func CheckCRC(data []byte) bool {
	data_len := len(data)
	calc_crc16 := LTBus_Compute_CRC16(data, uint16(data_len-3))
	crc16_bytes := []byte{data[data_len-3], data[data_len-2]}
	packet_crc16 := binary.LittleEndian.Uint16(crc16_bytes)
	return calc_crc16 == packet_crc16
}

func LTBus_VLink_Device_Loop(serial_port *serial.Port, packet_size int, stop_signal <-chan os.Signal) {
	for {
		select {

		case <-stop_signal:
			fmt.Printf("Closing Device...\n")
			err := serial_port.Close()
			if err != nil {
				fmt.Printf("Closing Device...ERR\n")
			} else {
				fmt.Printf("Closing Device...OK\n")
			}
			return

		default:
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

			mmap := rx_buffer[7 : 7+data_size]
			LTBus_VLink_Broadcast(mmap)

			time.Sleep(1 * time.Millisecond)
		}
	}
}

func LTBus_VLink_Device_Init(wg *sync.WaitGroup) {
	device_id := flag.Int("device_id", 0x1000, "LTBus Device ID")
	device_port := flag.String("device_port", "/dev/ttyUSB0", "Device Serial Port Name")
	buadrate := flag.Int("buadrate", 115200, "Serial Link BaudRate")
	packet_size := flag.Int("packet_size", 0, "Serial Link Packet Size")
	debug := flag.Bool("debug", false, "Enable Debug Logs")
	flag.Parse()

	fmt.Printf("Connecting to Device: %s...\n", *device_port)
	serial_conf := &serial.Config{Name: *device_port, Baud: *buadrate, ReadTimeout: time.Second * 1}
	serial_port, err := serial.OpenPort(serial_conf)
	if err != nil {
		fmt.Printf("Connecting to Device: %s...ERR\n", *device_port)
		os.Exit(1)
	}

	ltbrr_device_id := LTBus_Read_Request(0xA000, 2)
	serial_port.Write(ltbrr_device_id)
	rx_buffer := make([]byte, 12)
	n_bytes, err := serial_port.Read(rx_buffer)
	if err != nil || n_bytes != 12 {
		fmt.Printf("Can not Read Device ID @ 0xA000: Invalid LTBus Packet Frame\n")
		os.Exit(1)
	}

	if !CheckCRC(rx_buffer) {
		fmt.Printf("Can not Read Device ID @ 0xA000: Invalid CRC\n")
		os.Exit(1)
	}

	_device_id := LTBus_Decode_U16(rx_buffer)
	sw_id := uint16(*device_id)
	if _device_id != sw_id {
		fmt.Printf("Device ID Mismatch HW_ID: 0x%04X, SW_ID: 0x%04X\n", _device_id, sw_id)
		os.Exit(1)
	}

	VLink_Debug = *debug
	fmt.Printf("Connecting to Device: %s...OK\n", *device_port)

	stop_signal := make(chan os.Signal, 1)
	signal.Notify(stop_signal, os.Interrupt, syscall.SIGINT)
	wg.Go(func() { LTBus_VLink_Device_Loop(serial_port, *packet_size, stop_signal) })
}
