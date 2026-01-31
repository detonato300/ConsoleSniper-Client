//go:build !amd64

package security

import "time"

// rdtsc implementation for non-amd64 architectures.
func rdtsc() uint64 {
	return uint64(time.Now().UnixNano())
}

// cpuid_check implementation for non-amd64 architectures.
func cpuid_check() bool {
	return false
}
