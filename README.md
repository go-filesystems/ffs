# ffs

Pure-Go access to the **Berkeley Fast File System (FFS)** — no root, no external tools, no CGO.

FFS is the same on-disk format that FreeBSD calls **UFS**: `FFSv1 = UFS1` and
`FFSv2 = UFS2`, shared across FreeBSD, NetBSD and OpenBSD. Rather than duplicate
the parser, this module is a thin re-export of
[`go-filesystems/ufs`](https://github.com/go-filesystems/ufs), which reads and
writes the format. It exists so callers that think in "FFS" terms have a
matching import path.

```go
import ffs "github.com/go-filesystems/ffs"

fs, err := ffs.OpenFile("disk.ffs")
if err != nil { /* ... */ }
defer fs.Close()

data, err := fs.ReadFile("/etc/rc.conf")
```

## Support

Everything the `ufs` driver provides:

| | Status |
|---|---|
| Read (FFSv1/UFS1, FFSv2/UFS2) | ✅ |
| Write + `Mkfs` (UFS2) | ✅ via `ufs` |
| NetBSD / OpenBSD FFS images | ✅ (same on-disk format; tested via `makefs`) |

See [`go-filesystems/ufs`](https://github.com/go-filesystems/ufs) for the full
capability list and details.

## Module

```
github.com/go-filesystems/ffs
```
