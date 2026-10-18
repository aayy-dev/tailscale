// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package osuser

import (
	"context"
	"errors"
	"os/user"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

func TestProbeGetentDoubleDashSupport(t *testing.T) {
	origExecGetent := execGetent
	t.Cleanup(func() {
		execGetent = origExecGetent
	})

	tests := []struct {
		name string
		out  []byte
		err  error
		want bool
	}{
		{
			name: "supported",
			out:  []byte("root:x:0:0:root:/root:/bin/sh\n"),
			want: true,
		},
		{
			name: "error",
			err:  errors.New("unsupported"),
			want: false,
		},
		{
			name: "wrong-user",
			out:  []byte("daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\n"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execGetent = func(_ context.Context, args ...string) ([]byte, error) {
				return tt.out, tt.err
			}
			if got := probeGetentDoubleDashSupport(); got != tt.want {
				t.Fatalf("probeGetentDashDashSupport() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserLookupGetentUsesProbeResult(t *testing.T) {
	origExecGetent := execGetent
	origProbe := getentDoubleDashSupported
	t.Cleanup(func() {
		execGetent = origExecGetent
		getentDoubleDashSupported = origProbe
	})

	tests := []struct {
		name      string
		supported bool
		wantArgs  []string
	}{
		{
			name:      "supported",
			supported: true,
			wantArgs:  []string{"passwd", "--", "alice"},
		},
		{
			name:      "unsupported",
			supported: false,
			wantArgs:  []string{"passwd", "alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getentDoubleDashSupported = sync.OnceValue(func() bool { return tt.supported })
			execGetent = func(_ context.Context, args ...string) ([]byte, error) {
				if !reflect.DeepEqual(args, tt.wantArgs) {
					t.Fatalf("args = %q, want %q", args, tt.wantArgs)
				}
				return []byte("alice:x:1001:1001:Alice:/home/alice:/bin/sh\n"), nil
			}
			std := func(string) (*user.User, error) {
				t.Fatal("std lookup should not be called")
				return nil, nil
			}
			u, shell, err := userLookupGetent("alice", std)
			if err != nil {
				t.Fatalf("userLookupGetent error: %v", err)
			}
			if u.Username != "alice" || shell != "/bin/sh" {
				t.Fatalf("got user=%+v shell=%q", u, shell)
			}
		})
	}
}

func TestAndroidRootFallback(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Android fallback only applies to GOOS=linux binaries")
	}
	origIsAndroid, origHome, origShell, origExecGetent := isAndroid, androidRootHome, androidRootShell, execGetent
	t.Cleanup(func() {
		isAndroid, androidRootHome, androidRootShell, execGetent = origIsAndroid, origHome, origShell, origExecGetent
	})
	isAndroid = func() bool { return true }
	androidRootHome, androidRootShell = "/data/local/home", "/data/local/bin/zsh"
	execGetent = func(context.Context, ...string) ([]byte, error) { return nil, errors.New("no getent") }
	notFound := func(string) (*user.User, error) { return nil, errors.New("not found") }

	for _, name := range []string{"root", "0"} {
		u, shell, err := lookup(name, notFound, true)
		if err != nil {
			t.Fatalf("lookup(%q): %v", name, err)
		}
		if u.Uid != "0" || u.Username != "root" || u.HomeDir != androidRootHome || shell != androidRootShell {
			t.Errorf("lookup(%q) = %+v, %q; want built-in root with home %q, shell %q", name, u, shell, androidRootHome, androidRootShell)
		}
	}
	if _, _, err := lookup("nobody", notFound, true); err == nil {
		t.Error("lookup(nobody) succeeded; want error for non-root users")
	}

	realRoot := &user.User{Uid: "0", Gid: "0", Username: "root", HomeDir: "/root"}
	found := func(string) (*user.User, error) { return realRoot, nil }
	if u, _, err := lookup("root", found, false); err != nil || u != realRoot {
		t.Errorf("lookup(root) with a working user database = %+v, %v; want it unchanged", u, err)
	}
}
