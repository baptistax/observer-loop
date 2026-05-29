package watcher

import "testing"

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	got := BuildMessage("notepad.exe")
	want := "The notepad.exe killed itself..."

	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCleanProcessNameUsesBaseName(t *testing.T) {
	t.Parallel()

	got := CleanProcessName(`C:\Windows\System32\notepad.exe`)
	want := "notepad.exe"

	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSelectNextProcessChoosesNextHigherPID(t *testing.T) {
	t.Parallel()

	processes := []ProcessInfo{
		{PID: 100, Name: "alpha.exe"},
		{PID: 200, Name: "beta.exe"},
		{PID: 300, Name: "gamma.exe"},
	}

	got, err := SelectNextProcess(150, processes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.PID != 200 || got.Name != "beta.exe" {
		t.Fatalf("got %+v, want PID 200 beta.exe", got)
	}
}

func TestSelectNextProcessWrapsAround(t *testing.T) {
	t.Parallel()

	processes := []ProcessInfo{
		{PID: 100, Name: "alpha.exe"},
		{PID: 200, Name: "beta.exe"},
	}

	got, err := SelectNextProcess(250, processes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.PID != 100 || got.Name != "alpha.exe" {
		t.Fatalf("got %+v, want PID 100 alpha.exe", got)
	}
}
