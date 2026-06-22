// Copyright (c) 2026, go-filesystems
// SPDX-License-Identifier: BSD-3-Clause

package ffs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	ufs "github.com/go-filesystems/ufs"
	"github.com/go-volumes/safeio"
)

// This package is a thin re-export of github.com/go-filesystems/ufs, so the
// security hardening lives entirely in ufs. This regression suite proves the
// inheritance: a malicious image opened through ffs.Open / ffs.OpenFile is
// turned into a graceful error by the hardened ufs driver rather than
// panicking or OOM-ing the host. The headline vector is the di_size=2^40
// over-allocation (ufs block.go ReadFileBody), reproduced here end to end.

// On-disk struct fs / dinode offsets we need to hand-build a minimal image.
// These mirror the (unexported) constants in the ufs package; we only need
// the handful that place the superblock and root inode.
const (
	sblockUFS2 = 65536 // ufs.SblockUFS2

	// struct fs field offsets (FreeBSD 14.x amd64 layout).
	offSblkno        = 8
	offCblkno        = 12
	offIblkno        = 16
	offDblkno        = 20
	offNcg           = 44
	offBsize         = 48
	offFsize         = 52
	offFrag          = 56
	offBshift        = 80
	offFshift        = 84
	offFsbtodb       = 100
	offSbsize        = 104
	offNindir        = 116
	offInopb         = 120
	offIpg           = 184
	offFpg           = 188
	offMaxsymlinklen = 1320
	offOldInodefmt   = 1324
	offMaxfilesize   = 1328
	offMagic         = 1372

	// ufs2_dinode field offsets.
	inoOffMode   = 0
	inoOffNlink  = 2
	inoOffSize   = 16
	inoOffBlocks = 24
	inoOffDirect = 112

	inodeSize = 256

	// geometry (matches the ufs test fixture)
	bsize   = 4096
	fsize   = 4096
	frag    = 1
	inopb   = bsize / inodeSize // 16
	ipg     = 32
	fpg     = 256
	ncg     = 1
	sblkno  = sblockUFS2 / fsize // 16
	cblkno  = sblkno + 2         // 18
	iblkno  = 24
	dblkno  = 26
	fsbtodb = 3
	bshift  = 12
	fshift  = 12
	nindir  = bsize / 8 // 512
	magicU2 = 0x19540119

	ifdir uint16 = 0o040000

	inoRoot     = 2
	fragRootDir = dblkno // root directory data block
)

// buildMinImage hand-builds a small but valid UFS2 image whose root inode is
// a directory. di_size for the root inode is set to `rootSize`, letting a
// caller poison it. The image is exactly 1 MiB (fpg*fsize), matching the
// geometry the superblock advertises, so the image-derived ceiling is 1 MiB.
func buildMinImage(rootSize uint64) []byte {
	le := binary.LittleEndian
	img := make([]byte, fpg*fsize) // 1 MiB

	sb := img[sblockUFS2 : sblockUFS2+1376]
	le.PutUint32(sb[offSblkno:], sblkno)
	le.PutUint32(sb[offCblkno:], cblkno)
	le.PutUint32(sb[offIblkno:], iblkno)
	le.PutUint32(sb[offDblkno:], dblkno)
	le.PutUint32(sb[offNcg:], ncg)
	le.PutUint32(sb[offBsize:], bsize)
	le.PutUint32(sb[offFsize:], fsize)
	le.PutUint32(sb[offFrag:], frag)
	le.PutUint32(sb[offBshift:], bshift)
	le.PutUint32(sb[offFshift:], fshift)
	le.PutUint32(sb[offFsbtodb:], fsbtodb)
	le.PutUint32(sb[offSbsize:], 1376)
	le.PutUint32(sb[offNindir:], nindir)
	le.PutUint32(sb[offInopb:], inopb)
	le.PutUint32(sb[offIpg:], ipg)
	le.PutUint32(sb[offFpg:], fpg)
	le.PutUint32(sb[offMaxsymlinklen:], 120)
	le.PutUint32(sb[offOldInodefmt:], 2)
	le.PutUint64(sb[offMaxfilesize:], 1<<48)
	le.PutUint32(sb[offMagic:], magicU2)

	// Root inode (ino 2): a directory pointing at one data block.
	off := iblkno*fsize + inoRoot*inodeSize
	in := img[off : off+inodeSize]
	le.PutUint16(in[inoOffMode:], ifdir|0o755)
	le.PutUint16(in[inoOffNlink:], 2)
	le.PutUint64(in[inoOffSize:], rootSize)
	le.PutUint64(in[inoOffBlocks:], 1)
	le.PutUint64(in[inoOffDirect:], fragRootDir)

	return img
}

// TestHardenInherited_DiSizeOOM is the inheritance proof: a root directory
// inode whose di_size is 2^40 must NOT drive a ~1 TiB allocation when read
// through ffs.Open. The hardened ufs driver rejects it with
// safeio.ErrTooLarge; no panic, no OOM.
func TestHardenInherited_DiSizeOOM(t *testing.T) {
	img := buildMinImage(1 << 40)

	fs, err := Open(bytes.NewReader(img), int64(len(img)))
	if err != nil {
		t.Fatalf("ffs.Open: %v", err)
	}
	defer fs.Close()

	_, err = fs.ListDir("/")
	if err == nil {
		t.Fatal("expected error for oversized root di_size, got nil")
	}
	if !errors.Is(err, safeio.ErrTooLarge) {
		t.Fatalf("expected safeio.ErrTooLarge, got %v", err)
	}
	// And it must wrap the safeio family sentinel too.
	if !errors.Is(err, safeio.ErrSafeIO) {
		t.Fatalf("error should wrap safeio.ErrSafeIO: %v", err)
	}
}

// TestHardenInherited_Sane confirms the same builder yields a working image
// at a sane di_size, so the test above proves a *behavioral* rejection rather
// than a builder that simply never opens. A zero-size root directory reads
// back as an empty listing (no entries, no error).
func TestHardenInherited_Sane(t *testing.T) {
	img := buildMinImage(0)
	fs, err := Open(bytes.NewReader(img), int64(len(img)))
	if err != nil {
		t.Fatalf("ffs.Open: %v", err)
	}
	defer fs.Close()
	ents, err := fs.ListDir("/")
	if err != nil {
		t.Fatalf("ListDir on sane image: %v", err)
	}
	if len(ents) != 0 {
		t.Fatalf("expected empty root listing, got %d entries", len(ents))
	}
}

// TestHardenInherited_FSAlias guards the re-export contract that makes the
// inheritance possible: ffs.FS must be the very same type as ufs.FS.
func TestHardenInherited_FSAlias(t *testing.T) {
	var _ *ufs.FS = (*FS)(nil)
}
