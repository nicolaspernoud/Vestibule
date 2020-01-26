package du

import (
	"errors"
	"syscall"
	"time"
	"unsafe"
)

//DiskUsage is an object holding the disk usage
type DiskUsage struct {
	freeBytes  int64
	totalBytes int64
	availBytes int64
}

func statfs(path string, du *DiskUsage) error {
	h := syscall.MustLoadDLL("kernel32.dll")
	c := h.MustFindProc("GetDiskFreeSpaceExW")

	c.Call(
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		uintptr(unsafe.Pointer(&du.freeBytes)),
		uintptr(unsafe.Pointer(&du.totalBytes)),
		uintptr(unsafe.Pointer(&du.availBytes)))
	return nil
}

// NewDiskUsage returns an object holding the disk usage of volumePath assuming volumePath is a valid path
func NewDiskUsage(volumePath string) (*DiskUsage, error) {
	var du DiskUsage
	ch := make(chan error, 1)
	go func() {
		ch <- statfs(volumePath, &du)
	}()
	select {
	case err := <-ch:
		return &du, err
	case <-time.After(5 * time.Second):
		return &du, errors.New("timeout getting disk usage")
	}
}

// Free returns the total free bytes on file system
func (du *DiskUsage) Free() uint64 {
	return uint64(du.freeBytes)
}

// Available returns the total available bytes on file system to an unpriveleged user
func (du *DiskUsage) Available() uint64 {
	return uint64(du.availBytes)
}

// Size returns the total size of the file system
func (du *DiskUsage) Size() uint64 {
	return uint64(du.totalBytes)
}

// Used returns the total bytes used in file system
func (du *DiskUsage) Used() uint64 {
	return du.Size() - du.Free()
}

// Usage returns the percentage of use on the file system
func (du *DiskUsage) Usage() float32 {
	return float32(du.Used()) / float32(du.Size())
}
