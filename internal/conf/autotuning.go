package conf

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const defaultRAMMB = 4096 // 4 GB fallback when total RAM cannot be determined

// sysRAMMB returns total physical RAM in megabytes.
// On Linux it reads /proc/meminfo; on other platforms it returns defaultRAMMB.
func sysRAMMB() int {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return defaultRAMMB
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.Atoi(fields[1])
				if err == nil && kb > 0 {
					return kb / 1024
				}
			}
		}
	}
	return defaultRAMMB
}

// sysCPUCount returns the number of logical CPUs available to the process.
func sysCPUCount() int {
	return runtime.NumCPU()
}

// clampInt clamps v to [lo, hi].
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// nextPowerOf2 returns the smallest power of 2 that is >= v.
// Returns 1 for v <= 0.
func nextPowerOf2(v int) int {
	if v <= 0 {
		return 1
	}
	p := 1
	for p < v {
		p <<= 1
	}
	return p
}
