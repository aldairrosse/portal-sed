package scoring

import (
	"testing"
)

func TestEmployeeScore(t *testing.T) {
	tests := []struct {
		name       string
		categories []CategoryScore
		want       float64
	}{
		{
			name: "single category single goal complete",
			categories: []CategoryScore{
				{Weight: 100, Goals: []GoalScore{{Weight: 100, ProgressPercent: 100}}},
			},
			want: 100,
		},
		{
			name: "single category single goal partial",
			categories: []CategoryScore{
				{Weight: 100, Goals: []GoalScore{{Weight: 100, ProgressPercent: 50}}},
			},
			want: 50,
		},
		{
			name: "two categories equal weight",
			categories: []CategoryScore{
				{Weight: 50, Goals: []GoalScore{{Weight: 100, ProgressPercent: 100}}},
				{Weight: 50, Goals: []GoalScore{{Weight: 100, ProgressPercent: 0}}},
			},
			want: 50,
		},
		{
			name: "multiple goals within category",
			categories: []CategoryScore{
				{
					Weight: 100,
					Goals: []GoalScore{
						{Weight: 60, ProgressPercent: 100},
						{Weight: 40, ProgressPercent: 50},
					},
				},
			},
			want: 80, // (0.6*100)+(0.4*50) = 80
		},
		{
			name: "complex multi-category",
			categories: []CategoryScore{
				{
					Weight: 40,
					Goals: []GoalScore{
						{Weight: 50, ProgressPercent: 100},
						{Weight: 30, ProgressPercent: 80},
						{Weight: 20, ProgressPercent: 60},
					},
				},
				{
					Weight: 30,
					Goals: []GoalScore{
						{Weight: 100, ProgressPercent: 90},
					},
				},
				{
					Weight: 30,
					Goals: []GoalScore{
						{Weight: 50, ProgressPercent: 70},
						{Weight: 50, ProgressPercent: 50},
					},
				},
			},
			want: 79.4,
		},
		{
			name: "empty categories",
			categories: []CategoryScore{
				{Weight: 100, Goals: []GoalScore{}},
			},
			want: 0,
		},
		{
			name:       "no categories",
			categories: []CategoryScore{},
			want:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EmployeeScore(tt.categories)
			if !approxEqual(got, tt.want) {
				t.Errorf("EmployeeScore() = %v, want %v", got, tt.want)
			}
		})
	}
}
