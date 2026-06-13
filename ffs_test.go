// Copyright (c) 2026, go-filesystems
// SPDX-License-Identifier: BSD-3-Clause

package ffs

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func tool(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	for _, d := range []string{"/usr/local/sbin", "/usr/sbin", "/sbin", "/usr/bin"} {
		c := filepath.Join(d, name)
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

// TestInterop_MakefsFFS builds an FFS image with NetBSD's makefs and reads it
// back through this package (which re-exports the ufs driver), confirming the
// FFS import path works end to end. Skipped without makefs.
func TestInterop_MakefsFFS(t *testing.T) {
	makefs := tool("makefs")
	if makefs == "" {
		t.Skip("makefs not available")
	}
	src := t.TempDir()
	want := []byte("berkeley fast file system\n")
	if err := os.WriteFile(filepath.Join(src, "motd"), want, 0o644); err != nil {
		t.Fatal(err)
	}
	img := filepath.Join(t.TempDir(), "fs.ffs")
	if out, err := exec.Command(makefs, "-t", "ffs", "-B", "le", "-o", "version=2", "-s", "8m", img, src).CombinedOutput(); err != nil {
		t.Fatalf("makefs: %v\n%s", err, out)
	}
	fs, err := OpenFile(img)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer fs.Close()
	if got, err := fs.ReadFile("/motd"); err != nil || !bytes.Equal(got, want) {
		t.Fatalf("ReadFile(/motd): err=%v equal=%v", err, bytes.Equal(got, want))
	}
}
