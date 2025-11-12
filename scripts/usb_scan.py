import re
import json
import subprocess
from glob import glob


def get_udev_id(udev_info: str) -> str:
    vend_id = re.findall(r'ATTRS{idVendor}=="([0-9a-f]{4})"', udev_info)[0]
    prod_id = re.findall(r'ATTRS{idProduct}=="([0-9a-f]{4})"', udev_info)[0]
    return f"{vend_id}:{prod_id}"


if __name__ == '__main__':
    ports_info = {}

    usbd_list = subprocess.check_output(['lsusb'], text=True).split('\n')
    for usbd in usbd_list:
        if 'STMicroelectronics' in usbd:
            usbd_parts = usbd.split()
            device_id = usbd_parts[5]
            ports_info[device_id] = {'desc': usbd.split(device_id)[1]}

    udev_list = glob('/dev/ttyACM*') + glob('/dev/ttyUSB*')
    for udev in udev_list:
        udev_info = subprocess.check_output(['udevadm', 'info', '-a', '-n', udev], text=True)
        udev_id = get_udev_id(udev_info)
        if udev_id in ports_info:
            ports_info[udev_id]['port_name'] = udev

    print(json.dumps(ports_info, indent=2))
