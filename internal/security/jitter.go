package security

// MeasureJitter performs a tight loop to measure CPU cycle variance.
func MeasureJitter() uint64 {
	var maxDiff uint64
	for i := 0; i < 100; i++ {
		t1 := rdtsc()
		// Tight loop to measure overhead
		_ = i * i
		t2 := rdtsc()
		diff := t2 - t1
		if i > 0 && diff > maxDiff {
			maxDiff = diff
		}
	}
	return maxDiff
}

// IsHighJitter checks if the measured jitter exceeds a safe threshold.
func IsHighJitter(measured, threshold uint64) bool {
	return measured > threshold
}
