package formattools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatFloatTrimZero(t *testing.T) {

	tests := []struct {
		name string
		arg float64
		want string
	}{
		{
			name: "No zeros",
			arg: 10.1,
			want: "10.1",
		},
		{
			name: "Trailing zeros",
			arg: 10.1000,
			want: "10.1",
		},
		{
			name: "Long number",
			arg: 10.10000001,
			want: "10.10000001",
		},
		
		{
			name: "No decimal point",
			arg: 10,
			want: "10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatFloatTrimZero(tt.arg))
		})
	}
}
