// +build !windows

package du

import (
	"testing"
)

func TestNewDiskUsage(t *testing.T) {
	type args struct {
		volumePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"must_be_getting_usage", args{volumePath: "/"}, false},
		{"must_error", args{volumePath: "not_a_path"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDiskUsage(tt.args.volumePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDiskUsage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && (got.Size() <= 0 || got.Free() <= 0) {
				t.Errorf("DiskUsage sizes must be > 0 ; got size : %v, free : %v", got.Size(), got.Free())
			}
		})
	}
}
