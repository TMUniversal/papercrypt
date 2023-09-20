package internal

import "fmt"

func SprintBinarySize64(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.2f KiB", float64(size)/1024)
	}
	if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2f MiB", float64(size)/(1024*1024))
	}
	if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2f GiB", float64(size)/(1024*1024*1024))
	}
	return fmt.Sprintf("%.2f TiB", float64(size)/(1024*1024*1024*1024))
}

func SprintBinarySize(size int) string {
	return SprintBinarySize64(int64(size))
}
