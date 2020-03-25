package sysinfo

import (
	"os"
	"path/filepath"

	"github.com/nicolaspernoud/vestibule/pkg/du"
)

// Info returns only the disk usage on windows
func Info() (*SysInfo, error) {
	sysinfo := &SysInfo{}
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
	return sysinfo, nil
}
