package sequence

import "sync/atomic"

var seq atomic.Int64

func Next() int64 {
	return seq.Add(1)
}
