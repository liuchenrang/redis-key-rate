package redis

import (
	"fmt"
	"regexp"
	"testing"
)

func TestStateKey_CalcRate2(t *testing.T) {
	compile, _ := regexp.Compile("OpenSearchCmsTbkIte::getCmsItems")
	compile.Match([]byte("OpenSearchCmsTbkIte::getCmsItems"))
}
func TestStateKey_CalcRate(t1 *testing.T) {
	type fields struct {
		queryCount uint64
		hitCount   uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name:"xx",
			fields: struct {
				queryCount uint64
				hitCount   uint64
			}{queryCount: 2, hitCount: 1},
			want:fmt.Sprintf("redis queryCount %d hitCount %d rate %s", 2, 1, "0.50"),
		},
		{
			name:"xx",
			fields: struct {
				queryCount uint64
				hitCount   uint64
			}{queryCount: 5, hitCount: 3},
			want:fmt.Sprintf("redis queryCount %d hitCount %d rate %s", 5, 3, "0.60"),
			
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &StateKey{
				queryCount: tt.fields.queryCount,
				hitCount:   tt.fields.hitCount,
			}
			if got := t.CalcRate(); got != tt.want {
				t1.Errorf("CalcRate() = %v, want %v", got, tt.want)
			}
		})
	}
}