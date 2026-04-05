package solparser

import "time"

func NowUs() int64 {
	return time.Now().UnixMicro()
}
