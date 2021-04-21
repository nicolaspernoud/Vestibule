package sysinfo

import (
	"net/http"
	"runtime"
	"testing"

	"github.com/nicolaspernoud/vestibule/pkg/tester"
)

func TestGetInfo(t *testing.T) {
	handler := http.HandlerFunc(GetInfo)
	if runtime.GOOS == "windows" {
		tester.DoRequestOnHandler(t, handler, "GET", "/", nil, "", http.StatusOK, `{"usedgb"`)
	} else {
		tester.DoRequestOnHandler(t, handler, "GET", "/", nil, "", http.StatusOK, `{"uptime"`)
	}
}
