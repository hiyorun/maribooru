package helpers_test

import (
	"maribooru/internal/helpers"
	"testing"
)

func TestSluggify(t *testing.T) {
	type args struct {
		slug string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				slug: "test",
			},
			want: "test",
		},
		{
			name: "test spaces",
			args: args{
				slug: "test spaces",
			},
			want: "test_spaces",
		},
		{
			name: "test spaces with underscores",
			args: args{
				slug: "test spaces with_underscores",
			},
			want: "test_spaces_with_underscores",
		},
		{
			name: "test spaces with double underscores",
			args: args{
				slug: "test spaces with__double_underscores",
			},
			want: "test_spaces_with_double_underscores",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := helpers.Sluggify(tt.args.slug); got != tt.want {
				t.Errorf("helpers.Sluggify() = %v, want %v", got, tt.want)
			}
		})
	}
}
