package syncgo

import (
	"reflect"
	"testing"
)

func TestDetectRAGCycle(t *testing.T) {
	t.Run("invalid or empty table", func(t *testing.T) {
		for _, dp := range []*DiningPhilosophers{nil, {}, {Count: 5}} {
			report := DetectRAGCycle(dp)
			if report.Deadlock || report.Cycle == nil || len(report.Cycle) != 0 {
				t.Fatalf("expected an empty non-deadlock report, got %+v", report)
			}
		}
	})
	t.Run("circular wait and released fork", func(t *testing.T) {
		dp := NewDiningPhilosophers("deadlock")
		if report := DetectRAGCycle(dp); report.Deadlock {
			t.Fatal("new table must not be deadlocked")
		}
		for i := 0; i < dp.Count; i++ {
			dp.ForksOwner[i] = i
			dp.States[i] = StateDeadlocked
		}
		report := DetectRAGCycle(dp)
		if !report.Deadlock || !reflect.DeepEqual(report.Cycle, []int{1, 2, 3, 4, 5}) {
			t.Fatalf("expected circular wait, got %+v", report)
		}
		dp.ForksOwner[0] = -1
		if report := DetectRAGCycle(dp); report.Deadlock || len(report.Cycle) != 0 {
			t.Fatalf("released fork breaks the cycle, got %+v", report)
		}
	})
}
