import time
import socket
import struct


vlink_server_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)


def fmt_packet_hex(packet: bytes) -> str:
    return ' '.join([f"0x{x:02X}" for x in packet])


if __name__ == '__main__':
    vlink_server_sock.connect(('127.0.0.1', 6400))
    device_mmap = vlink_server_sock.recv(14)
    print(f"Received {len(device_mmap)} Bytes: {fmt_packet_hex(device_mmap)}")
    input()
