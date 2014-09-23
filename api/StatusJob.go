package ninja

import (
	"fmt"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/wolfeidau/usage"
)

type metricUsage struct {
	Memory uint64  `json:"memory"`
	CPU    float64 `json:"cpu"`
}

// StatusJob internal state for the status job
type StatusJob struct {
	serialNo       string
	driverName     string
	conn           *Connection
	statusTicker   *time.Ticker
	processMonitor *usage.ProcessMonitor
}

// CreateStatusJob create and configure a new status job which will log cpu usage and memory
func CreateStatusJob(conn *Connection, driverName string) (*StatusJob, error) {
	serial := config.Serial()

	return &StatusJob{processMonitor: usage.CreateProcessMonitor(), conn: conn, serialNo: serial, driverName: driverName}, nil
}

// Start spawn the status job
func (m *StatusJob) Start() {
	m.statusTicker = time.NewTicker(2 * time.Second)

	go func() {
		for {
			<-m.statusTicker.C

			m.conn.SendNotification(fmt.Sprintf("$node/%s/module/status", m.serialNo), m.driverName, m.buildUsage())

		}
	}()

}

func (m *StatusJob) buildName() *simplejson.Json {

	js, _ := simplejson.NewJson([]byte(`{}`))

	js.SetPath([]string{}, m.driverName)

	return js
}

func (m *StatusJob) buildUsage() *simplejson.Json {

	js, _ := simplejson.NewJson([]byte(`{}`))

	memUsage := m.processMonitor.GetMemoryUsage()
	cpuUsage := m.processMonitor.GetCpuUsage()

	js.SetPath([]string{}, &metricUsage{Memory: memUsage.Resident, CPU: cpuUsage.Total})

	return js
}
