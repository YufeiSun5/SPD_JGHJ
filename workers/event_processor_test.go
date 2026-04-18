package workers

import "testing"

func TestCorrectProductionCounterDelta(t *testing.T) {
	tests := []struct {
		name        string
		rawDelta    int
		oldValue    int
		newValue    int
		lastNonzero int
		want        int
	}{
		{
			name:        "reconnect recovery does not count restored total",
			rawDelta:    77,
			oldValue:    0,
			newValue:    77,
			lastNonzero: 77,
			want:        0,
		},
		{
			name:        "reconnect recovery counts only missed increment",
			rawDelta:    80,
			oldValue:    0,
			newValue:    80,
			lastNonzero: 77,
			want:        3,
		},
		{
			name:        "real reset starts a new baseline",
			rawDelta:    1,
			oldValue:    0,
			newValue:    1,
			lastNonzero: 77,
			want:        1,
		},
		{
			name:        "normal increase keeps raw delta",
			rawDelta:    1,
			oldValue:    76,
			newValue:    77,
			lastNonzero: 76,
			want:        1,
		},
		{
			name:        "counter reset remains negative for caller to ignore",
			rawDelta:    -77,
			oldValue:    77,
			newValue:    0,
			lastNonzero: 77,
			want:        -77,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := correctProductionCounterDelta(tt.rawDelta, tt.oldValue, tt.newValue, &tt.lastNonzero)
			if got != tt.want {
				t.Fatalf("correctProductionCounterDelta() = %d, want %d", got, tt.want)
			}
		})
	}
}
