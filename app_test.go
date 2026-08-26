package main

import (
	"fmt"
	"testing"
)

func TestBuildTime(t *testing.T) {
	app := NewApp()
	got := app.GetBuildTime()
	fmt.Printf("GetBuildTime() = %q (len=%d)\n", got, len(got))
	if got == "" {
		t.Logf("BuildTime is empty. Was the binary built with -ldflags '-X main.BuildTime=...'?")
	} else {
		t.Logf("BuildTime looks good: %s", got)
	}
}