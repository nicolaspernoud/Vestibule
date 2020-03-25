// +build !windows

package sysinfo

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/nicolaspernoud/vestibule/pkg/du"
	"golang.org/x/sys/unix"
)

// float64(1<<SI_LOAD_SHIFT) == 65536.0
const scale = 65536.0

var mtx = new(sync.RWMutex)

// Info returns a complete system information object on unix
func Info() (*SysInfo, error) {
	mtx.Lock()
	defer mtx.Unlock()
	sysinfo := &SysInfo{}
	// Get loads and memory usage
	rawsysinfo := &unix.Sysinfo_t{}
	if err := unix.Sysinfo(rawsysinfo); err != nil {
		return nil, err
	}
	// Get number of cores
	cores := float64(runtime.NumCPU())
	// Get executable path
	ex, err := os.Executable()
	if err != nil {
		return nil, err

	}
	exPath := filepath.Dir(ex)
	usage, err := du.NewDiskUsage(exPath)
	if err == nil {
		sysinfo.UsedGB = usage.Used() / du.GB
		sysinfo.TotalGB = usage.Size() / du.GB
	}
	sysinfo.Uptime = time.Duration(rawsysinfo.Uptime) * time.Second
	sysinfo.Load = (float64(rawsysinfo.Loads[0]) / scale) / cores
	unit := uint64(rawsysinfo.Unit) * 1024 // kB
	sysinfo.TotalRAM = uint64(rawsysinfo.Totalram) / unit
	sysinfo.FreeRAM = uint64(rawsysinfo.Freeram) / unit
	return sysinfo, nil
}
