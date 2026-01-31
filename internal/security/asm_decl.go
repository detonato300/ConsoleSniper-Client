//go:build amd64

package security

func rdtsc() uint64
func cpuid_check() bool
