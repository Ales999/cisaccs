package utils

import (
	"reflect"
	"testing"
)

func TestParseMacLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want MacLineData
	}{
		// TODO: Add test cases.
		{name: "Test-1",
			args: args{line: "   1    548a.ba01.50b3    DYNAMIC     Gi0/43"},
			want: MacLineData{vlan: "1", mac: "548a.ba01.50b3", iface: "Gi0/43"},
		},
		{name: "Test-2",
			args: args{line: "   1      0e55.6c89.a819   dynamic ip,ipx,assigned,other Port-channel20             "},
			want: MacLineData{vlan: "1", mac: "0e55.6c89.a819", iface: "Port-channel20"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseMacLine(tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMacLine() = %v, want %v", got, tt.want)
			}
		})
	}
}
