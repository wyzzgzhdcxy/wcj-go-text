//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	modkernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procSetPriorityClass         = modkernel32.NewProc("SetPriorityClass")
	procGetForegroundWindow      = modkernel32.NewProc("GetForegroundWindow")
	procSetForegroundWindow      = modkernel32.NewProc("SetForegroundWindow")
	procGetWindowThreadProcessId = modkernel32.NewProc("GetWindowThreadProcessId")
)

const (
	IDLE_PRIORITY_CLASS         = 0x40
	HIGH_PRIORITY_CLASS         = 0x80
	ABOVE_NORMAL_PRIORITY_CLASS = 0x800
	NORMAL_PRIORITY_CLASS       = 0x20
)

func main() {
	// 假设我们要置顶的进程ID已知
	processId := uint32(12704) // 替换为实际的进程ID

	// 设置进程优先级
	handle, _ := syscall.OpenProcess(0xF000, false, processId)
	if handle == 0 {
		panic("无法打开进程")
	}
	defer func(handle syscall.Handle) {
		err := syscall.CloseHandle(handle)
		if err != nil {

		}
	}(handle)

	r, _, _ := procSetPriorityClass.Call(uintptr(handle), uintptr(HIGH_PRIORITY_CLASS))
	if r == 0 {
		panic("无法设置进程优先级")
	}

	// 置顶窗口
	var hwnd uintptr
	for {
		hwnd, _, _ = procGetForegroundWindow.Call()
		if hwnd != 0 {
			break
		}
	}

	var pid uint32
	tid, _, _ := procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == processId && tid != 0 {
		procSetForegroundWindow.Call(hwnd)
	}
}
