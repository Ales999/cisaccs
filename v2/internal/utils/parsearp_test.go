package utils

import (
	"reflect"
	"testing"
)

func TestParseArpLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want ArpLineData
	}{
		// TODO: Add test cases.
		{name: "Test-1",
			args: args{line: "Internet  10.23.1.2               -   aabb.cc00.2030  ARPA   Ethernet0/3  "},
			want: ArpLineData{ip: "10.23.1.2", mac: "aabb.cc00.2030", iface: "Ethernet0/3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseArpLine(tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseArpLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
