package log

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCityAndCountryFromRequest(t *testing.T) {

	// Request types
	requestFromLocalHost := httptest.NewRequest("GET", "/test", strings.NewReader(""))
	requestFromLocalHost.RemoteAddr = "[::1]:1234"

	requestFromLondon := httptest.NewRequest("GET", "/test", strings.NewReader(""))
	requestFromLondon.RemoteAddr = "81.2.69.142:1234"

	requestIpv6FromFrance := httptest.NewRequest("GET", "/test", strings.NewReader(""))
	requestIpv6FromFrance.RemoteAddr = "[2a02:cfc0:cd4:6bd0:18da:635c:74b2:6109]:1234"

	requestWithEmptyIP := httptest.NewRequest("GET", "/test", strings.NewReader(""))

	requestWithNotAnIp := httptest.NewRequest("GET", "/test", strings.NewReader(""))
	requestWithNotAnIp.RemoteAddr = "definitely_not_an:_ip[!]"

	// Test case of non existing ip location database (must return an error)
	ipDbLocation = "<NO IP LOCATION DATABASE>"

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Ip database location doesn't exist",
			args: args{
				req: requestFromLondon,
			},
			want: "no ip location database",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCityAndCountryFromRequest(tt.args.req); got != tt.want {
				t.Errorf("GetCityAndCountryFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test case of existing ip location database

	ipDbLocation = "../../configs/ipgeodatabase/GeoLite2-City.mmdb"

	tests = []struct {
		name string
		args args
		want string
	}{
		{
			name: "Request from localhost",
			args: args{
				req: requestFromLocalHost,
			},
			want: "localhost",
		},
		{
			name: "Request from london",
			args: args{
				req: requestFromLondon,
			},
			want: "Londres, Royaume-Uni.",
		},
		{
			name: "Request from london, again",
			args: args{
				req: requestFromLondon,
			},
			want: "Londres, Royaume-Uni",
		},
		{
			name: "Request with ipv6 from france",
			args: args{
				req: requestIpv6FromFrance,
			},
			want: ", France.",
		},
		{
			name: "Request with ipv6 from france, again",
			args: args{
				req: requestIpv6FromFrance,
			},
			want: ", France",
		},
		{
			name: "Request with empty ip",
			args: args{
				req: requestWithEmptyIP,
			},
			want: "ip not found",
		},
		{
			name: "Request with not an ip",
			args: args{
				req: requestWithNotAnIp,
			},
			want: "ip could not be parsed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCityAndCountryFromRequest(tt.args.req); got != tt.want {
				t.Errorf("GetCityAndCountryFromRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}
