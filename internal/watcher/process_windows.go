//go:build windows

package watcher

import (
	"context"
	"fmt"
	"syscall"
	"time"
	"unsafe"
)

type Handle uintptr

type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [syscall.MAX_PATH]uint16
}

const (
	synchronize                    = 0x00100000
	processQueryLimitedInformation = 0x1000
	th32csSnapProcess              = 0x00000002
	waitObject0                    = 0x00000000
	waitTimeout                    = 0x00000102
	waitFailed                     = 0xFFFFFFFF
	mbOK                           = 0x00000000
	mbIconInformation              = 0x00000040
	invalidHandleValue             = ^uintptr(0)
)

var (
	modKernel32               = syscall.NewLazyDLL("kernel32.dll")
	modUser32                 = syscall.NewLazyDLL("user32.dll")
	procOpenProcess           = modKernel32.NewProc("OpenProcess")
	procCloseHandle           = modKernel32.NewProc("CloseHandle")
	procWaitForSingleObject   = modKernel32.NewProc("WaitForSingleObject")
	procQueryFullProcessImage = modKernel32.NewProc("QueryFullProcessImageNameW")
	procCreateSnapshot        = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First        = modKernel32.NewProc("Process32FirstW")
	procProcess32Next         = modKernel32.NewProc("Process32NextW")
	procMessageBoxW           = modUser32.NewProc("MessageBoxW")
)

func OpenForWait(pid uint32) (Handle, error) {
	access := uintptr(synchronize | processQueryLimitedInformation)

	ret, _, err := procOpenProcess.Call(access, 0, uintptr(pid))
	if ret == 0 {
		return 0, syscallError("OpenProcess", err)
	}

	return Handle(ret), nil
}

func Close(handle Handle) {
	if handle == 0 {
		return
	}

	_, _, _ = procCloseHandle.Call(uintptr(handle))
}

func WaitExitOrCancel(ctx context.Context, handle Handle, step time.Duration) error {
	timeout := uint32(step / time.Millisecond)
	if timeout == 0 {
		timeout = 500
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		ret, _, err := procWaitForSingleObject.Call(uintptr(handle), uintptr(timeout))
		switch ret {
		case waitObject0:
			return nil
		case waitTimeout:
			continue
		case waitFailed:
			return syscallError("WaitForSingleObject", err)
		default:
			return fmt.Errorf("unexpected wait state: %d", ret)
		}
	}
}

func ProcessName(pid uint32) (string, error) {
	handle, err := OpenForWait(pid)
	if err != nil {
		return "", err
	}
	defer Close(handle)

	buf := make([]uint16, syscall.MAX_PATH)
	size := uint32(len(buf))

	ret, _, callErr := procQueryFullProcessImage.Call(
		uintptr(handle),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return "", syscallError("QueryFullProcessImageNameW", callErr)
	}

	return syscall.UTF16ToString(buf[:size]), nil
}

func ListProcesses() ([]ProcessInfo, error) {
	snapshot, _, err := procCreateSnapshot.Call(th32csSnapProcess, 0)
	if snapshot == invalidHandleValue {
		return nil, syscallError("CreateToolhelp32Snapshot", err)
	}
	defer Close(Handle(snapshot))

	entry := processEntry32{
		Size: uint32(unsafe.Sizeof(processEntry32{})),
	}

	ret, _, callErr := procProcess32First.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, syscallError("Process32FirstW", callErr)
	}

	processes := make([]ProcessInfo, 0, 128)
	for {
		processes = append(processes, ProcessInfo{
			PID:  entry.ProcessID,
			Name: syscall.UTF16ToString(entry.ExeFile[:]),
		})

		ret, _, callErr = procProcess32Next.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			errno, ok := callErr.(syscall.Errno)
			if ok && errno == syscall.ERROR_NO_MORE_FILES {
				break
			}
			if !ok && callErr == syscall.Errno(0) {
				break
			}
			return nil, syscallError("Process32NextW", callErr)
		}
	}

	return processes, nil
}

func ShowMessageBox(title, message string) error {
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return err
	}

	messagePtr, err := syscall.UTF16PtrFromString(message)
	if err != nil {
		return err
	}

	ret, _, callErr := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(mbOK|mbIconInformation),
	)
	if ret == 0 {
		return syscallError("MessageBoxW", callErr)
	}

	return nil
}

func SleepContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func currentPID() int {
	return syscall.Getpid()
}

func syscallError(op string, err error) error {
	if err == nil || err == syscall.Errno(0) {
		return fmt.Errorf("%s failed", op)
	}

	return fmt.Errorf("%s failed: %w", op, err)
}
