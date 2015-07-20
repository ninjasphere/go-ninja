package simtime

import (
	"container/heap"
	"time"

	"github.com/ninjasphere/go-ninja/config"
)

var enabled = config.Bool(false, "simtime.enable")
var startOffset = config.Duration(time.Hour*24*30, "simtime.offset")
var allowFuture = config.Bool(false, "simtime.allowFuture")

var currentTime = time.Now()

var addQueue = make(chan *event)
var queue = eventQueue{}

var added = make(chan bool)
var tick = make(chan bool)
var start = make(chan bool)

func addQueuedEvents() {
	for {
		select {
		case e := <-addQueue:
			heap.Push(&queue, e)
		default:
			if len(queue) > 0 {
				return
			}
		}

	}
}

func init() {

	if !enabled {
		return
	}

	SetCurrentTime(time.Now().Add(-startOffset))

	go func() {
		<-start
		for {

			//spew.Dump("Adding queued events")
			// Add any to-be-queued events (the heap isn't threadsafe)
			addQueuedEvents()

			//spew.Dump("pop")

			event := heap.Pop(&queue).(*event)

			//spew.Dump("event", event)

			currentTime = time.Unix(0, event.fireAt)

			if !allowFuture && currentTime.After(time.Now()) {
				//spew.Dump("Sleeping")
				time.Sleep(currentTime.Sub(time.Now()))
			}

			if event.c != nil {
				//spew.Dump("Sending event")
				event.c <- currentTime
				//spew.Dump("Waiting for tick")
				select {
				case <-tick:
				case <-time.After(time.Second * 5):
					panic("When using simtime, you MUST call simtime.Continue() when you are done with your time event")
				}
				//spew.Dump("Got tick")
			} else {
				event.f(currentTime)
			}

			if event.interval > 0 {
				event.fireAt = currentTime.Add(event.interval).UnixNano()
				heap.Push(&queue, event)
			}

		}
	}()

}

func Start() {
	select {
	case start <- true:
	default:
	}
}

func Continue() {
	if enabled {
		tick <- true
	}
}

func SetCurrentTime(time time.Time) {
	currentTime = time
	queue = eventQueue{}
}

func Now() time.Time {
	if !enabled {
		return time.Now()
	}
	return currentTime
}

func Enabled() bool {
	return enabled
}

func Tick(d time.Duration) <-chan time.Time {
	if !enabled {
		return time.Tick(d)
	}
	return addChannelEvent(d, true)
}

func After(d time.Duration) <-chan time.Time {
	if !enabled {
		return time.After(d)
	}
	return addChannelEvent(d, false)
}

func Sleep(d time.Duration) {
	<-After(d)
}

func addFuncEvent(f func(time.Time), d time.Duration, interval bool) {
	e := &event{
		fireAt: currentTime.Add(d).UnixNano(),
		f:      f,
	}

	if interval {
		e.interval = d
	}

	addEvent(e)

}

func addChannelEvent(d time.Duration, interval bool) chan time.Time {

	e := &event{
		fireAt: currentTime.Add(d).UnixNano(),
		c:      make(chan time.Time),
	}

	if interval {
		e.interval = d
	}

	addEvent(e)

	return e.c
}

func addEvent(e *event) {
	addQueue <- e

	select {
	case added <- true:
	default:
	}
}

type event struct {
	f        func(time.Time)
	c        chan time.Time
	fireAt   int64
	interval time.Duration
}

// A eventQueue implements heap.Interface and holds Events.
type eventQueue []*event

func (pq eventQueue) Len() int { return len(pq) }

func (pq eventQueue) Less(i, j int) bool {
	return pq[i].fireAt < pq[j].fireAt
}

func (pq eventQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *eventQueue) Push(x interface{}) {
	item := x.(*event)
	*pq = append(*pq, item)
}

func (pq *eventQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}
