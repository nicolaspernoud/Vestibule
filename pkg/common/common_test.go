package common

import (
	"os"
	"testing"
)

func TestStringValueFromEnv(t *testing.T) {
	os.Setenv("MY_EV", "from_env")
	var rv string
	var err error
	type args struct {
		ev  string
		def string
	}
	tests := []struct {
		name     string
		args     args
		expected string
		wantErr  bool
	}{
		{"string_value_from_env", args{"MY_EV", "test"}, "from_env", false},
		{"string_value_from_def", args{"MY_DEF", "test"}, "test", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rv, err = StringValueFromEnv(tt.args.ev, tt.args.def); (err != nil) != tt.wantErr {
				t.Errorf("StringValueFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rv != tt.expected {
				t.Errorf("StringValueFromEnv() error ; got %v, expected %v", rv, tt.expected)
			}
		})
	}
}

func TestIntValueFromEnv(t *testing.T) {
	os.Setenv("MY_EV", "from_env")
	var rv int
	var err error
	type args struct {
		ev  string
		def int
	}
	tests := []struct {
		name     string
		args     args
		expected int
		wantErr  bool
	}{
		{"int_value_from_def", args{"MY_DEF", 1}, 1, false},
		{"string_on_int_from_env", args{"MY_EV", 1}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rv, err = IntValueFromEnv(tt.args.ev, tt.args.def); (err != nil) != tt.wantErr {
				t.Errorf("IntValueFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rv != tt.expected {
				t.Errorf("IntValueFromEnv() error ; got %v, expected %v", rv, tt.expected)
			}
		})
	}
}

func TestBoolValueFromEnv(t *testing.T) {
	os.Setenv("MY_EV", "from_env")
	var rv bool
	var err error
	type args struct {
		ev  string
		def bool
	}
	tests := []struct {
		name     string
		args     args
		expected bool
		wantErr  bool
	}{

		{"bool_value_from_def", args{"MY_DEF", true}, true, false},
		{"string_on_bool_from_def", args{"MY_EV", true}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rv, err = BoolValueFromEnv(tt.args.ev, tt.args.def); (err != nil) != tt.wantErr {
				t.Errorf("BoolValueFromEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rv != tt.expected {
				t.Errorf("BoolValueFromEnv() error ; got %v, expected %v", rv, tt.expected)
			}
		})
	}
}
