package swag

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	SearchDir = "./testdata/format_test"
	Excludes  = "./testdata/format_test/web"
	MainFile  = "main.go"
)

func Test_isBlankComment(t *testing.T) {
	type args struct {
		comment string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				comment: " ",
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				comment: " A",
			},
			want: false,
		},
		{
			name: "test3",
			args: args{
				comment: " \t",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBlankComment(tt.args.comment)
			if got != tt.want {
				t.Errorf("isBlankComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isSwagComment(t *testing.T) {
	type args struct {
		comment string
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				comment: "@Param some_id ",
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				comment: "@ ",
			},
			want: false,
		},
		{
			name: "test3",
			args: args{
				comment: "@Success {object} ",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSwagComment(tt.args.comment)
			if got != tt.want {
				t.Errorf("isSwagComment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_replaceRange(t *testing.T) {
	type args struct {
		s     []byte
		start int
		end   int
		new   byte
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test_replaceSuccess",
			args: args{
				s:     []byte("// @ID  get-ids"),
				start: 6,
				end:   8,
				new:   '\t',
			},
			want: []byte("// @ID\tget-ids"),
		},
		{
			name: "test1_replaceFail",
			args: args{
				s:     []byte("// @ID  A pet"),
				start: 6,
				end:   8,
				new:   '\t',
			},
			want: []byte("// @ID\tA pet"),
		},
		{
			name: "test1_replaceFail2",
			args: args{
				s:     []byte("// @ID  "),
				start: 6,
				end:   12,
				new:   '\t',
			},
			want: []byte("// @ID\t"),
		},
		{
			name: "test1_replaceFail3",
			args: args{
				s:     []byte("// @ID  "),
				start: 2,
				end:   1,
				new:   '\t',
			},
			want: []byte("// @ID  "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceRange(tt.args.s, tt.args.start, tt.args.end, tt.args.new); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_separatorFinder(t *testing.T) {
	type args struct {
		comment string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				comment: `// @Param   some_id  query int  "some id  data" Enums(1, 2, 3)`,
			},
			want: `// @Param|some_id|query|int|"some id  data"|Enums(1, 2, 3)`,
		},
		{
			name: "test2",
			args: args{
				comment: `// @Summary   A pet store. `,
			},
			want: `// @Summary|A pet store. `,
		},
		{
			name: "test3",
			args: args{
				comment: `// @Summary    `,
			},
			want: `// @Summary    `,
		},
		{
			name: "test4",
			args: args{
				comment: `// @Failure      400       {object}  web.APIError{data=web.D ,data2=web.D2}  "We need ID!!"`,
			},
			want: `// @Failure|400|{object}|web.APIError{data=web.D ,data2=web.D2}|"We need ID!!"`,
		},
		{
			name: "test5",
			args: args{
				comment: `// `,
			},
			want: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := separatorFinder(tt.args.comment, '|')
			assert.Equal(t, got, tt.want)
		})
	}
}
