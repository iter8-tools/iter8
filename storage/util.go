package storage

import (
	"golang.org/x/sys/unix"
)

// GetVolumeUsage gets the available and total capacity of a volume, in that order
func GetVolumeUsage(path string) (uint64, uint64, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	// Available blocks * size per block = available space in bytes
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	// Total blocks * size per block = available space in bytes
	totalBytes := stat.Blocks * uint64(stat.Bsize)

	return availableBytes, totalBytes, nil
}
