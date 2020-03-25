package sysinfo

import (
	"encoding/json"
	"net/http"
	"time"
)

// SysInfo represents global sustem information
type SysInfo struct {
	Uptime   time.Duration `json:"uptime,omitempty"`   // time since boot
	Load     float64       `json:"load,omitempty"`     // 1 minute load average divided by number of CPU Cores
	TotalRAM uint64        `json:"totalram,omitempty"` // total usable main memory size [kB]
	FreeRAM  uint64        `json:"freeram,omitempty"`  // available memory size [kB]
	UsedGB   uint64        `json:"usedgb,omitempty"`   // Used GB of HDD
	TotalGB  uint64        `json:"totalgb,omitempty"`  // Total GB of HDD
}

// GetInfo returns the system information
func GetInfo(w http.ResponseWriter, r *http.Request) {
	info, err := Info()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(info)
}
