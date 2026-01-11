package health

import "sync/atomic"

var ready int32 = 0

func SetReady(value bool) {
	if value {
		atomic.StoreInt32(&ready, 1)
	} else {
		atomic.StoreInt32(&ready, 0)
	}
}

func IsReady() bool {
	return atomic.LoadInt32(&ready) == 1
}
