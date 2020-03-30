package redis

import (
	"fmt"
	"sync/atomic"
	"time"
)

var (
	StateKeyOps  = New()
)
type StateKey struct {
	queryCount uint64
	hitCount uint64
	hitQuery int32
}

func New() *StateKey  {
	return &StateKey{queryCount:0,hitCount:0,hitQuery:0}
}
func (t *StateKey) IncQuery() {
	atomic.AddUint64(&t.queryCount, 1)
}
func (t *StateKey) HitQuery() bool {
	return atomic.AddInt32(&t.hitQuery, 1) > 0
}

func (t *StateKey) ResetHitQuery() bool {
	return atomic.AddInt32(&t.hitQuery, -1) >= 0
}
func (t *StateKey) GetQueryCount() uint64{
	return atomic.LoadUint64(&t.queryCount)
}
func (t *StateKey) IncHitCount() {
	atomic.AddUint64(&t.hitCount, 1)
}
func (t *StateKey) GetHitCount() uint64{
	return atomic.LoadUint64(&t.hitCount)
}
func (t *StateKey) GetHitQueryCount() int32{
	return atomic.LoadInt32(&t.hitQuery)
}
func (t *StateKey) CalcRate() string {
	queryCount := t.GetQueryCount()
	hitCount := t.GetHitCount()
	var u float32 = 0.0;
	if queryCount > 0 {
		u = float32(hitCount) / float32(queryCount)
	}
	return fmt.Sprintf("redis queryCount %d hitCount %d rate %0.2f", queryCount, hitCount, u)
}
func (t *StateKey) RunRate()  {
	for{
		timer := time.NewTimer(time.Second*2)
		<-timer.C
		fmt.Printf(t.CalcRate() + "\r\n");
	}
}
