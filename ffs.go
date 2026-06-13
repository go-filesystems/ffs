// Copyright (c) 2026, go-filesystems
// SPDX-License-Identifier: BSD-3-Clause

// Package ffs reads the Berkeley Fast File System (FFS).
//
// FFS is the same on-disk format that FreeBSD calls UFS: FFSv1 = UFS1 and
// FFSv2 = UFS2, shared across FreeBSD, NetBSD and OpenBSD. Rather than
// duplicate the parser, this package is a thin re-export of
// github.com/go-filesystems/ufs, which reads (and writes) the format. It
// exists so callers that think in "FFS" terms have a matching import path.
//
//	fs, err := ffs.OpenFile("disk.ffs")
//	data, err := fs.ReadFile("/etc/rc.conf")
package ffs

import (
	"io"

	ufs "github.com/go-filesystems/ufs"
)

// FS is an opened FFS/UFS filesystem. It is an alias of the ufs driver's FS.
type FS = ufs.FS

// Open parses the FFS/UFS superblock from rs (FFSv1/UFS1 or FFSv2/UFS2) and
// returns a read-only handle. size is the backing image size, or -1 if unknown.
func Open(rs io.ReaderAt, size int64) (*FS, error) { return ufs.Open(rs, size) }

// OpenFile opens path read-only and parses its FFS/UFS superblock.
func OpenFile(path string) (*FS, error) { return ufs.OpenFile(path) }

// OpenRW opens an FFS/UFS image for read and write.
func OpenRW(rs io.ReaderAt, wa io.WriterAt, size int64) (*FS, error) {
	return ufs.OpenRW(rs, wa, size)
}
