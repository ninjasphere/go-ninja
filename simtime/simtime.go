package simtime

import (
	"container/heap"
	"runtime/debug"
	"time"

	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/logger"
)

var log = logger.GetLogger("simtime")

var enabled = config.Bool(false, "simtime.enable")
var startOffset = config.Duration(time.Hour*24*30, "simtime.startOffset")
var offset = config.Duration(0, "simtime.offset")
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

	log.Infof("Starting. Enabled %t", enabled)

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

			if !allowFuture && currentTime.After(time.Now().Add(-offset)) {
				//spew.Dump("Sleeping")
				time.Sleep(currentTime.Sub(time.Now()))
			}

			if event.c != nil {
				select {
				case event.c <- currentTime:
				case <-time.After(time.Minute * 5):
					panic("An simtime channel listener didn't respond. (Check that you aren't calling simtime.Continue() more than once per tick.)")
				}

				select {
				case <-tick:
				case <-time.After(time.Minute * 5):
					panic("When using simtime, you MUST call simtime.Continue() when you are done with your time event with 5 minutes.")
				}
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
	log.Infof("Starting. Time: %s", Now())
	select {
	case start <- true:
	default:
	}
}

func Continue() {
	if enabled {
		select {
		case tick <- true:
		case <-time.After(time.Minute * 5):
			panic("Continue took more than 5 seconds. Check that you are only calling it once per channel tick.")
		}
	}
}

func SetCurrentTime(time time.Time) {
	log.Infof("Setting time to %s", time)
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
		stack:  string(debug.Stack()),
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
		stack:  string(debug.Stack()),
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
	stack    string
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
