package cisaccs

import (
	"reflect"
	"testing"

	"github.com/ales999/cisaccs/internal/utils"
)

func TestCisFindVlan(t *testing.T) {
	type args struct {
		vlanlines  []string
		fndvlandid int
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 utils.VlanLineData
	}{
		// TODO: Add test cases.
		{
			name: "Test-1",
			args: args{vlanlines: []string{
				"1    default                          active   ",
				"141  TD_*14.16/28                     active   ",
				"1002 fddi-default                     act/unsup    ",
			}, fndvlandid: 141,
			},
			want:  true,
			want1: *utils.NewVlanLineData(141, "TD_*14.16/28"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CisFindVlan(tt.args.vlanlines, tt.args.fndvlandid)
			if got != tt.want {
				t.Errorf("CisFindVlan() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CisFindVlan() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
