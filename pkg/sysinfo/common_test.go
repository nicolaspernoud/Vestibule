package sysinfo

import (
	"net/http"
	"testing"

	"github.com/nicolaspernoud/vestibule/pkg/tester"
)

func TestGetInfo(t *testing.T) {
	handler := http.HandlerFunc(GetInfo)
	tester.DoRequestOnHandler(t, handler, "GET", "/", nil, "", http.StatusOK, `{"uptime"`)
}
