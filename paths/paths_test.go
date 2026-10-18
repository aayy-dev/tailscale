// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package paths

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestStateDirOverride(t *testing.T) {
	old := stateDirOverride
	defer func() { stateDirOverride = old }()
	stateDirOverride = "/data/local/lib/tailscale"

	if got, want := DefaultTailscaledStateFile(), filepath.Join(stateDirOverride, "tailscaled.state"); got != want {
		t.Errorf("DefaultTailscaledStateFile = %q; want %q", got, want)
	}
	if runtime.GOOS != "plan9" {
		if got, want := DefaultTailscaledStateDir(), filepath.Clean(stateDirOverride); got != want {
			t.Errorf("DefaultTailscaledStateDir = %q; want %q", got, want)
		}
	}
	if !MakeAutomaticStateDir() {
		t.Error("MakeAutomaticStateDir = false; want true")
	}
	if runtime.GOOS != "windows" {
		if got, want := DefaultTailscaledSocket(), filepath.Join(stateDirOverride, "tailscaled.sock"); got != want {
			t.Errorf("DefaultTailscaledSocket = %q; want %q", got, want)
		}
	}
}

func TestSocketPathOverride(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("socket is a named pipe on Windows")
	}
	oldDir, oldSock := stateDirOverride, socketPathOverride
	defer func() { stateDirOverride, socketPathOverride = oldDir, oldSock }()
	stateDirOverride = "/data/local/var/lib/tailscale"
	socketPathOverride = "/data/local/var/run/tailscale/tailscaled.sock"

	if got := DefaultTailscaledSocket(); got != socketPathOverride {
		t.Errorf("DefaultTailscaledSocket = %q; want %q", got, socketPathOverride)
	}
	if got, want := DefaultTailscaledStateFile(), filepath.Join(stateDirOverride, "tailscaled.state"); got != want {
		t.Errorf("DefaultTailscaledStateFile = %q; want %q", got, want)
	}
}
