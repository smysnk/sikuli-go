package main

import (
	"reflect"
	"testing"
)

func TestNormalizeServerFlagArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "bare listen uses default grpc address",
			in:   []string{"-listen"},
			want: []string{"-listen", defaultGRPCListenAddr},
		},
		{
			name: "explicit listen value preserved",
			in:   []string{"-listen", "127.0.0.1:6000"},
			want: []string{"-listen", "127.0.0.1:6000"},
		},
		{
			name: "bare listen before another flag gets default",
			in:   []string{"-listen", "-admin-listen", ":9000"},
			want: []string{"-listen", defaultGRPCListenAddr, "-admin-listen", ":9000"},
		},
		{
			name: "bare admin listen uses default admin address",
			in:   []string{"-listen", "-admin-listen"},
			want: []string{"-listen", defaultGRPCListenAddr, "-admin-listen", defaultAdminListenAddr},
		},
		{
			name: "utility command unchanged",
			in:   []string{"init:js-examples"},
			want: []string{"init:js-examples"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := normalizeServerFlagArgs(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("normalizeServerFlagArgs(%v)=%v want=%v", tc.in, got, tc.want)
			}
		})
	}
}
