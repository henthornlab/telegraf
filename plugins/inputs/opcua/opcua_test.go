package opcua

import (
	"log"
	"testing"
	"time"
)

type testUpdate struct {
	currVal  float64
	prevVal  float64
	absDev   float64
	maxInter string
	expected bool
}

func TestOpcua_NeedsUpdate(t *testing.T) {

	// Assumes the item was last updated 60 seconds ago
	var testConds = []testUpdate{
		{1.0, 0.8, 0.01, "90s", true},
		{1.0, 1.0, 0.01, "90s", false},
		{-16.0, -15.0, 0.01, "90s", true},
		{-16.0, -15.0, 3, "15s", true},
		{-16.0, -15.0, 3, "15s", true},
		{-16.0, -15.0, 3, "90s", false},
	}

	for i := range testConds {

		tNow := time.Now()

		o := opcuaNode{
			currentValue:  testConds[i].currVal,
			previousValue: testConds[i].prevVal,
			AbsDeviation:  testConds[i].absDev,
			lastUpdate:    tNow.Add(-time.Second * 60),
		}

		o.maxTimeInterval, _ = time.ParseDuration(testConds[i].maxInter)

		log.Print("NeedsUpdate is ", o.NeedsUpdate())

		if o.NeedsUpdate() != testConds[i].expected {
			t.Error("Update incorrect")
		}
	}
}

func TestOpc_sendMetrics(t *testing.T) {

}
