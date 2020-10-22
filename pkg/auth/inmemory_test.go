package auth

import (
	"os"
	"testing"
	"time"
)

func Test_setTokenLifetime(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{"no environnement", args{"OTHER_ENV", "10"}, 24 * time.Hour},
		{"wrong type", args{"INMEMORY_TOKEN_LIFE_DAYS", "A_STRING"}, 24 * time.Hour},
		{"to small", args{"INMEMORY_TOKEN_LIFE_DAYS", "-1"}, 24 * time.Hour},
		{"to big", args{"INMEMORY_TOKEN_LIFE_DAYS", "11 000"}, 24 * time.Hour},
		{"ok", args{"INMEMORY_TOKEN_LIFE_DAYS", "3"}, 72 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.args.key, tt.args.value)
			if got := setTokenLifetime(); got != tt.want {
				t.Errorf("setTokenLifetime() = %v, want %v", got, tt.want)
			}
		})
	}
}
