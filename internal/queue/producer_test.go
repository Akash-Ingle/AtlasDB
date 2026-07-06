package queue

import "testing"

func TestPartitionFor(t *testing.T) {
	cases := []struct {
		source string
		n      int
	}{
		{"payment-service", 4},
		{"user-service", 4},
		{"order-service", 4},
		{"auth-service", 8},
	}

	for _, c := range cases {
		p := partitionFor(c.source, c.n)
		if p < 0 || p >= c.n {
			t.Errorf("partitionFor(%q, %d) = %d, want 0..%d", c.source, c.n, p, c.n-1)
		}
	}

	// Deterministic: same source → same partition
	p1 := partitionFor("payment-service", 4)
	p2 := partitionFor("payment-service", 4)
	if p1 != p2 {
		t.Errorf("partitionFor not deterministic: %d != %d", p1, p2)
	}
}

func TestStreamKey(t *testing.T) {
	if got := StreamKey(0); got != "events:0" {
		t.Errorf("StreamKey(0) = %q, want events:0", got)
	}
	if got := StreamKey(3); got != "events:3" {
		t.Errorf("StreamKey(3) = %q, want events:3", got)
	}
}
