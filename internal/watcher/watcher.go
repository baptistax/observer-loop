package watcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Config struct {
	PID   uint
	Title string
	Out   io.Writer
}

type ProcessInfo struct {
	PID  uint32
	Name string
}

func Run(ctx context.Context, cfg Config) error {
	if cfg.PID == 0 {
		return errors.New("invalid PID")
	}

	if cfg.Out == nil {
		cfg.Out = io.Discard
	}

	selfPID := uint32(CurrentPID())
	currentPID := uint32(cfg.PID)

	for {
		current, handle, err := Attach(currentPID)
		if err != nil {
			return fmt.Errorf("attach to PID %d: %w", currentPID, err)
		}

		logAttach(cfg.Out, current)

		err = WaitExitOrCancel(ctx, handle, 500*time.Millisecond)
		Close(handle)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("watch PID %d: %w", current.PID, err)
		}

		if err := ShowMessageBox(cfg.Title, BuildMessage(current.Name)); err != nil {
			return fmt.Errorf("show notification: %w", err)
		}

		next, err := FindNextProcess(current.PID, selfPID)
		for err != nil {
			if ctx.Err() != nil {
				return nil
			}

			if err := SleepContext(ctx, time.Second); err != nil {
				return nil
			}

			next, err = FindNextProcess(current.PID, selfPID)
		}
		currentPID = next.PID
	}
}

func Attach(pid uint32) (ProcessInfo, Handle, error) {
	name, err := ProcessName(pid)
	if err != nil {
		return ProcessInfo{}, 0, err
	}

	handle, err := OpenForWait(pid)
	if err != nil {
		return ProcessInfo{}, 0, err
	}

	return ProcessInfo{
		PID:  pid,
		Name: CleanProcessName(name),
	}, handle, nil
}

func FindNextProcess(afterPID, selfPID uint32) (ProcessInfo, error) {
	processes, err := ListProcesses()
	if err != nil {
		return ProcessInfo{}, err
	}

	candidates := make([]ProcessInfo, 0, len(processes))
	for _, process := range processes {
		if process.PID == 0 || process.PID == selfPID {
			continue
		}

		name := CleanProcessName(process.Name)
		if name == "process" {
			continue
		}

		handle, err := OpenForWait(process.PID)
		if err != nil {
			continue
		}
		Close(handle)

		process.Name = name
		candidates = append(candidates, process)
	}

	return SelectNextProcess(afterPID, candidates)
}

func SelectNextProcess(afterPID uint32, processes []ProcessInfo) (ProcessInfo, error) {
	if len(processes) == 0 {
		return ProcessInfo{}, errors.New("no accessible process found")
	}

	sorted := append([]ProcessInfo(nil), processes...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PID < sorted[j].PID
	})

	for _, process := range sorted {
		if process.PID > afterPID {
			return process, nil
		}
	}

	return sorted[0], nil
}

func BuildMessage(processName string) string {
	return fmt.Sprintf("The %s killed itself...", CleanProcessName(processName))
}

func CleanProcessName(processName string) string {
	name := strings.TrimSpace(processName)
	if name == "" {
		return "process"
	}

	name = strings.ReplaceAll(name, "\\", "/")
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "" {
		return "process"
	}

	return name
}

func CurrentPID() int {
	return currentPID()
}

func logAttach(out io.Writer, process ProcessInfo) {
	fmt.Fprintf(out, "PID %d | %s\n", process.PID, process.Name)
}
