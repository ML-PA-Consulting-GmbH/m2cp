package tools

import (
	"fmt"
	"os"
)

// GetMemoryUsageRss returns the physical memory usage of the current process in bytes
func GetMemoryUsageRss() (uint64, error) {
	// Read RSS from /proc/self/statm (Linux only)
	file, err := os.Open("/proc/self/statm")
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var size, resident, share int64
	_, err = fmt.Fscanf(file, "%d %d %d", &size, &resident, &share)
	if err != nil {
		return 0, err
	}

	pageSize := uint64(os.Getpagesize()) // Get system page size
	return uint64(resident) * pageSize, nil
}

// GetMemoryUsageRssHumanized returns the physical memory usage of the current process in a human-friendly format
func GetMemoryUsageRssHumanized() string {
	rss, err := GetMemoryUsageRss()
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}
	if rss < 1<<10 {
		return fmt.Sprintf("%d B", rss)
	}
	if rss < 1<<20 {
		return fmt.Sprintf("%d KB", rss>>10)
	}
	if rss < 1<<30 {
		return fmt.Sprintf("%d MB", rss>>20)
	}
	return fmt.Sprintf("%d GB", rss>>30)
}
