import sys
import subprocess
from glob import glob


def exec_go_heap_diag() -> str:
    cmd = ['go', 'build', '-gcflags=-m'] + glob('lib/*')
    proc_ret = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    return proc_ret.stderr


def get_src_line(src_coord: str) -> str:
    coord_parts = src_coord.split(':')
    file_path = coord_parts[0]
    line_number = int(coord_parts[1]) - 1
    with open(file_path, 'r') as f:
        return f.readlines()[line_number].strip()


if __name__ == '__main__':
    go_heap_diag = exec_go_heap_diag()
    ghd_lines = go_heap_diag.split('\n')
    for l in ghd_lines:
        if 'escapes to heap' in l:
            src_coord = l.split(' ')[0]
            src_line = get_src_line(src_coord)
            if 'fmt.Print' not in src_line:
                print(f"{l} -> {src_line}")
