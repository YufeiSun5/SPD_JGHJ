package models

import (
	"testing"
	"time"
)

func TestApplyNumericSampleDebounce(t *testing.T) {
	tests := []struct {
		name      string
		sequence  []float64
		threshold float64
		want      []float64
	}{
		{
			name:      "instant false zero is discarded",
			sequence:  []float64{10, 0, 12},
			threshold: 5,
			want:      []float64{10, 12},
		},
		{
			name:      "instant real reset emits zero then new value",
			sequence:  []float64{10, 0, 1},
			threshold: 5,
			want:      []float64{10, 0, 1},
		},
		{
			name:      "long false zero window is collapsed",
			sequence:  []float64{10, 0, 0, 11},
			threshold: 5,
			want:      []float64{10, 11},
		},
		{
			name:      "long real reset emits one zero",
			sequence:  []float64{10, 0, 0, 1},
			threshold: 5,
			want:      []float64{10, 0, 1},
		},
		{
			name:      "small counter reset is allowed",
			sequence:  []float64{1, 0, 1},
			threshold: 5,
			want:      []float64{1, 0, 1},
		},
		{
			name:      "threshold boundary is allowed",
			sequence:  []float64{5, 0, 3},
			threshold: 5,
			want:      []float64{5, 0, 3},
		},
		{
			name:      "above threshold false zero is observed then discarded",
			sequence:  []float64{6, 0, 7},
			threshold: 5,
			want:      []float64{6, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suspicious := 0.0
			tag := &Tag{
				VarID:             1,
				VarName:           "counter",
				DataType:          "INT16",
				SuspiciousValue:   &suspicious,
				DebounceThreshold: tt.threshold,
				IsFirstUpdate:     true,
			}

			var got []float64
			baseTime := time.Date(2026, 4, 16, 8, 0, 0, 0, time.Local)
			for idx, sample := range tt.sequence {
				changes := tag.ApplyNumericSample(sample, baseTime.Add(time.Duration(idx)*time.Second), 1)
				if idx == 0 {
					got = append(got, tag.GetValue())
					continue
				}
				for _, change := range changes {
					got = append(got, change.NewValue)
				}
			}

			if len(got) != len(tt.want) {
				t.Fatalf("got sequence %v, want %v", got, tt.want)
			}
			for idx := range got {
				if got[idx] != tt.want[idx] {
					t.Fatalf("got sequence %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestShouldStartupSnapshot(t *testing.T) {
	enabled := 1
	disabled := 0

	tests := []struct {
		name   string
		tag    Tag
		expect bool
	}{
		{
			name:   "legacy change storage keeps first-frame snapshot",
			tag:    Tag{StoreMode: 1, StartupSnapshotEnable: nil},
			expect: true,
		},
		{
			name:   "explicit disable blocks first-frame snapshot",
			tag:    Tag{StoreMode: 1, StartupSnapshotEnable: &disabled},
			expect: false,
		},
		{
			name:   "explicit enable allows first-frame snapshot",
			tag:    Tag{StoreMode: 3, StartupSnapshotEnable: &enabled},
			expect: true,
		},
		{
			name:   "no-store mode never snapshots",
			tag:    Tag{StoreMode: 0, StartupSnapshotEnable: &enabled},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tag.ShouldStartupSnapshot(); got != tt.expect {
				t.Fatalf("ShouldStartupSnapshot() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestApplyNumericSampleStartupSuspiciousUsesHistoryReference(t *testing.T) {
	suspicious := 0.0
	historyValue := 77.0
	baseTime := time.Date(2026, 4, 16, 8, 0, 0, 0, time.Local)

	t.Run("startup false zero is held and next value becomes snapshot candidate", func(t *testing.T) {
		tag := &Tag{
			VarID:                  1,
			VarName:                "counter",
			DataType:               "INT16",
			SuspiciousValue:        &suspicious,
			DebounceThreshold:      5,
			DebounceLastValidValue: &historyValue,
			IsFirstUpdate:          true,
		}

		if changes := tag.ApplyNumericSample(0, baseTime, 1); len(changes) != 0 {
			t.Fatalf("first suspicious changes = %v, want none", changes)
		}
		if !tag.HasPendingDebounce() {
			t.Fatalf("startup suspicious value should stay pending")
		}
		if tag.GetLastUpdateTime().IsZero() == false {
			t.Fatalf("startup suspicious value must not finish first update")
		}

		if changes := tag.ApplyNumericSample(77, baseTime.Add(time.Second), 1); len(changes) != 0 {
			t.Fatalf("recovered startup value changes = %v, want none", changes)
		}
		if tag.HasPendingDebounce() {
			t.Fatalf("pending debounce should be cleared after recovery")
		}
		if got := tag.GetValue(); got != 77 {
			t.Fatalf("current value = %v, want 77", got)
		}
	})

	t.Run("startup real reset emits zero then new value", func(t *testing.T) {
		tag := &Tag{
			VarID:                  1,
			VarName:                "counter",
			DataType:               "INT16",
			SuspiciousValue:        &suspicious,
			DebounceThreshold:      5,
			DebounceLastValidValue: &historyValue,
			IsFirstUpdate:          true,
		}

		_ = tag.ApplyNumericSample(0, baseTime, 1)
		changes := tag.ApplyNumericSample(1, baseTime.Add(time.Second), 1)
		if len(changes) != 2 {
			t.Fatalf("changes len = %d, want 2 (%v)", len(changes), changes)
		}
		if changes[0].OldValue != 77 || changes[0].NewValue != 0 {
			t.Fatalf("first change = %+v, want 77 -> 0", changes[0])
		}
		if changes[1].OldValue != 0 || changes[1].NewValue != 1 {
			t.Fatalf("second change = %+v, want 0 -> 1", changes[1])
		}
	})

	t.Run("startup suspicious without history keeps legacy first frame behavior", func(t *testing.T) {
		tag := &Tag{
			VarID:             1,
			VarName:           "counter",
			DataType:          "INT16",
			SuspiciousValue:   &suspicious,
			DebounceThreshold: 5,
			IsFirstUpdate:     true,
		}

		if changes := tag.ApplyNumericSample(0, baseTime, 1); len(changes) != 0 {
			t.Fatalf("first frame changes = %v, want none", changes)
		}
		if tag.HasPendingDebounce() {
			t.Fatalf("without history reference, startup suspicious should not be held")
		}
		if got := tag.GetValue(); got != 0 {
			t.Fatalf("current value = %v, want 0", got)
		}
	})
}
