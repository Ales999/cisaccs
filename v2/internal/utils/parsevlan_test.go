package utils

import (
	"reflect"
	"testing"
)

func TestParseVlan(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want VlanLineData
	}{
		// TODO: Add test cases.
		{
			name: "Test1",
			args: args{line: "1    default                          active"},
			want: VlanLineData{vlid: 1, vlname: "default"},
		},
		{
			name: "Test2",
			args: args{line: "1    default                          active "},
			want: VlanLineData{vlid: 1, vlname: "default"},
		},
		{
			name: "Test3",
			args: args{line: "1    default                          active   "},
			want: VlanLineData{vlid: 1, vlname: "default"},
		},
		{
			name: "Test4",
			args: args{line: "141  TD_*14.16/28                     active"},
			want: VlanLineData{vlid: 141, vlname: "TD_*14.16/28"},
		},
		{
			name: "Test5",
			args: args{line: "1002 fddi-default                     act/unsup    "},
			want: VlanLineData{vlid: 1002, vlname: "fddi-default"},
		},
		{
			name: "Test6",
			args: args{line: "2001 VLAN2001                         active"},
			want: VlanLineData{vlid: 2001, vlname: "VLAN2001"},
		},
		{ // 150  Test12                           active    Te1/0/20, Gi2/0/20
			name: "Test7",
			args: args{line: "150  Test12                           active    Te1/0/20, Gi2/0/20"},
			want: VlanLineData{vlid: 150, vlname: "Test12"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseVlan(tt.args.line); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVlan() = %v, want %v", got, tt.want)
			}
		})
	}
}
