// Package snowflake implements a distributed unique ID generator based on
// Twitter's Snowflake algorithm, suitable for use across multiple services.
package snowflake

import (
	"sync"
	"time"
)

const (
	epoch        = 1700000000000 // 2023-11-14T22:13:20Z in ms (arbitrary start)
	workerBits   = 10
	sequenceBits = 12

	workerMax   = -1 ^ (-1 << workerBits)
	sequenceMax = -1 ^ (-1 << sequenceBits)

	timeShift   = workerBits + sequenceBits
	workerShift = sequenceBits
)

// Generator produces snowflake IDs.
type Generator struct {
	mu        sync.Mutex
	epoch     int64
	workerID  int64
	sequence  int64
	lastStamp int64
}

// New returns a Generator for the given worker ID.
func New(workerID int64) *Generator {
	return &Generator{
		epoch:    epoch,
		workerID: workerID & workerMax,
	}
}

// NextID returns the next unique ID.
func (g *Generator) NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()

	if now < g.lastStamp {
		// Clock moved backwards; wait until it catches up.
		for now < g.lastStamp {
			time.Sleep(time.Millisecond)
			now = time.Now().UnixMilli()
		}
	}

	if now == g.lastStamp {
		g.sequence = (g.sequence + 1) & sequenceMax
		if g.sequence == 0 {
			// Sequence exhausted this millisecond; wait for next ms.
			for now <= g.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastStamp = now

	id := ((now - g.epoch) << timeShift) |
		(g.workerID << workerShift) |
		g.sequence

	return id
}
