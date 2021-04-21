package common

import (
	"os"
	"testing"
)

func init() {
	disableLogFatal = true
}

func TestStringValueFromEnv(t *testing.T) {
	os.Setenv("MY_EV", "from_env")
	var rv string
	type args struct {
		ev  string
		def string
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{"string_value_from_env", args{"MY_EV", "test"}, "from_env"},
		{"string_value_from_def", args{"MY_DEF", "test"}, "test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv = StringValueFromEnv(tt.args.ev, tt.args.def)
			if rv != tt.expected {
				t.Errorf("StringValueFromEnv() error ; got %v, expected %v", rv, tt.expected)
			}
		})
	}
}

func TestIntValueFromEnv(t *testing.T) {
	os.Setenv("MY_EV", "from_env")
	var rv int
	type args struct {
		ev  string
		def int
	}
	tests := []struct {
		name     string
		args     args
		expected int
	}{
		{"int_value_from_def", args{"MY_DEF", 1}, 1},
		{"string_on_int_from_env", args{"MY_EV", 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv = IntValueFromEnv(tt.args.ev, tt.args.def)
			if rv != tt.expected {
				t.Errorf("IntValueFromEnv() error ; got %v, expected %v", rv, tt.expected)
			}
		})
	}
}

func TestBoolValueFromEnv(t *testing.T) {
	os.Setenv("MY_EV", "from_env")
	var rv bool
	type args struct {
		ev  string
		def bool
	}
	tests := []struct {
		name     string
		args     args
		expected bool
	}{

		{"bool_value_from_def", args{"MY_DEF", true}, true},
		{"string_on_bool_from_def", args{"MY_EV", true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rv = BoolValueFromEnv(tt.args.ev, tt.args.def)
			if rv != tt.expected {
				t.Errorf("BoolValueFromEnv() error ; got %v, expected %v", rv, tt.expected)
			}
		})
	}
}
