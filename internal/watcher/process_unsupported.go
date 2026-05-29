//go:build !windows

package watcher

import (
	"context"
	"errors"
	"time"
)

type Handle uintptr

func OpenForWait(pid uint32) (Handle, error) {
	return 0, errors.New("this project supports Windows only")
}

func Close(handle Handle) {}

func WaitExitOrCancel(ctx context.Context, handle Handle, step time.Duration) error {
	return errors.New("this project supports Windows only")
}

func ProcessName(pid uint32) (string, error) {
	return "", errors.New("this project supports Windows only")
}

func ListProcesses() ([]ProcessInfo, error) {
	return nil, errors.New("this project supports Windows only")
}

func ShowMessageBox(title, message string) error {
	return errors.New("this project supports Windows only")
}

func SleepContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func currentPID() int {
	return 0
}
