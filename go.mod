module github.com/go-filesystems/ffs

go 1.26.4

require github.com/go-filesystems/ufs v0.0.0

require (
	github.com/go-filesystems/interface v0.0.0 // indirect
	github.com/go-volumes/safeio v0.0.0-20260622072324-7f8eb19f6f8c // indirect
)

replace github.com/go-filesystems/ufs => ../ufs

replace github.com/go-filesystems/interface => ../interface
