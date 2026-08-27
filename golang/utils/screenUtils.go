package utils

import (
	"errors"
	"syscall"
)

func TurnOffDisplay() error {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessage := user32.NewProc("SendMessageW")

	// 广播消息关闭显示器
	hWnd := uintptr(0xFFFF)   // HWND_BROADCAST
	wMsg := uintptr(0x0112)   // WM_SYSCOMMAND
	wParam := uintptr(0xF170) // SC_MONITORPOWER
	lParam := uintptr(2)      // 关闭显示器

	_, _, err := sendMessage.Call(hWnd, wMsg, wParam, lParam)
	if err != nil && !errors.Is(syscall.Errno(0), err) {
		return err
	}
	return nil
}
