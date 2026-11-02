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
		{
			name: "Static-on-physical-port-is-tracked",
			args: args{line: " 582    00bc.6092.fe89    STATIC      Gi1/0/18"},
			want: MacLineData{vlan: "582", mac: "00bc.6092.fe89", iface: "Gi1/0/18"},
		},
		{
			name: "Static-on-VLAN-is-SVI-gateway-ignored",
			args: args{line: " 40    c4ab.4dff.39f0    STATIC      Vl40"},
			want: MacLineData{},
		},
		{
			name: "Static-on-VLAN-long-name-is-SVI-gateway-ignored",
			args: args{line: " 40    c4ab.4dff.39f0    STATIC      Vlan40"},
			want: MacLineData{},
		},
		{
			name: "Static-on-Switch-port-is-local-MAC-ignored-Cat4500",
			args: args{line: " 24      0008.e3ff.fc44    static ip,ipx,assigned,other Switch"},
			want: MacLineData{},
		},
		{
			name: "Static-on-Local-port-is-local-MAC-ignored",
			args: args{line: " 24      0008.e3ff.fc44    static ip,ipx,assigned,other Local"},
			want: MacLineData{},
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
