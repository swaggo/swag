package swag

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFieldsByAnySpace(t *testing.T) {
	type args struct {
		s string
		n int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"test1",
			args{
				"	aa	bb	cc dd 		ff",
				2,
			},
			[]string{"aa", "bb\tcc dd \t\tff"},
		},
		{"test2",
			args{
				`	aa	"bb	cc dd 		ff"`,
				2,
			},
			[]string{"aa", `"bb	cc dd 		ff"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, FieldsByAnySpace(tt.args.s, tt.args.n), "FieldsByAnySpace(%v,  %v)", tt.args.s, tt.args.n)
		})
	}
}
