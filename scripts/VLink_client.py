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
    vlink_server_sock.connect(('127.0.0.1', 6401))
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
