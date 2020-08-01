package opcua

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// OPCUA : Structure for all the plugin info
type OPCUA struct {
	ServerName string      `toml:"ServerName"`
	URL        string      `toml:"URL"`
	Nodes      []opcuaNode `toml:"Nodes"`
	ctx        context.Context
	client     *opcua.Client
	ID         []*ua.NodeID
	ReadValID  []*ua.ReadValueID
	req        *ua.ReadRequest
}

// Init : function to intialize the plugin
func (o *OPCUA) Init() error {

	o.ctx = context.Background()
	// This version doesn't support certificates yet, only anonymous and passwords
	o.client = opcua.NewClient(o.URL, opcua.SecurityMode(ua.MessageSecurityModeNone))

	log.Print("opcua: Starting opcua plugin to monitor: ", o.URL)

	if err := o.client.Connect(o.ctx); err != nil {
		log.Print("opcua: Fatal error in client connect")
		log.Fatal(err)
	}

	for i := range o.Nodes {
		tempID, err := ua.ParseNodeID(o.Nodes[i].NodeID)
		if err != nil {
			log.Fatalf("opcua: invalid node id: %v", err)
		}
		o.ID = append(o.ID, tempID)
		o.ReadValID = append(o.ReadValID, &ua.ReadValueID{NodeID: tempID})

		// Initialize info on the nodes for deviation and update interval checks
		o.Nodes[i].currentValue = math.MaxFloat64
		o.Nodes[i].lastUpdate = time.Now()

		o.Nodes[i].maxTimeInterval, _ = time.ParseDuration(o.Nodes[i].AtLeastEvery)
		log.Print("Adding ", o.Nodes[i].NodeID, " with absolute deviation of ", o.Nodes[i].AbsDeviation, " at least every ", o.Nodes[i].maxTimeInterval)
	}

	o.req = &ua.ReadRequest{
		MaxAge:             2000,
		NodesToRead:        o.ReadValID,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	return nil
}

func init() {
	inputs.Add("opcua", func() telegraf.Input { return &OPCUA{} })

}

// Gather implements the telegraf plugin interface method for data accumulation
func (o *OPCUA) Gather(acc telegraf.Accumulator) error {

	resp, err := o.client.Read(o.req)

	if err != nil {
		log.Fatalf("opcua: Read failed: %s", err)
	}

	for i := range resp.Results {

		if resp.Results[i].Status != ua.StatusOK {
			log.Fatalf("opcua: Status not OK: %v", resp.Results[i].Status)
		}
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		// Update the value to the latest read
		o.Nodes[i].UpdateValue(resp.Results[i].Value.Float())

		if o.Nodes[i].NeedsUpdate() {

			tags["server"] = o.ServerName
			tags["tag"] = o.Nodes[i].Tag
			tags["NodeID"] = o.Nodes[i].NodeID
			fields["value"] = o.Nodes[i].currentValue

			acc.AddFields("opcua", fields, tags, resp.Results[i].SourceTimestamp)

			o.Nodes[i].UpdateLastUpdate()

		}
	}

	return nil
}

const description = `Monitor nodes on an OPC-UA Server`
const sampleConfig = `
  ## OPC-UA Connection Configuration
  ##
  ## The plugin designed to connect to OPC UA devices
  ## Currently supports anonymous mode only
  ##
  ## Name given to OPC UA server for logging and tags
  ServerName = "Device"
  ## URL including endpoint. Only anonymous logins at this point
  URL = "opc.tcp://localhost.com:4840/endpoint"

  ## List of Nodes to monitor including Tag (name), NodeID, and the absolute deviation (set to 0.0 to record all points)
  ## AtLeastEvery forces an update on the point in Golang time, "10s", "30m", "24h", etc.
  Nodes = [
  {Tag = "HeatExchanger1 Temp", NodeID = "ns=2;s=TE-800-07/AI1/PV.CV", AbsDeviation = 0.10, AtLeastEvery = "30s"},
  {Tag = "Heat Exchanger1 Pressure", NodeID = "ns=2;i=1234", AbsDeviation = 0.0, AtLeastEvery = "1h"},
  ]
`

// SampleConfig returns a basic configuration for the plugin
func (o *OPCUA) SampleConfig() string {
	return sampleConfig
}

// Description returns a short description of what the plugin does
func (o *OPCUA) Description() string {
	return description
}

type opcuaNode struct {
	Tag             string  `toml:"Tag"`
	NodeID          string  `toml:"NodeID"`
	AbsDeviation    float64 `toml:"AbsDeviation"`
	AtLeastEvery    string  `toml:"AtLeastEvery"`
	maxTimeInterval time.Duration
	lastUpdate      time.Time
	currentValue    float64
	previousValue   float64
}

func (node *opcuaNode) UpdateValue(val float64) {
	node.previousValue = node.currentValue
	node.currentValue = val
}

func (node opcuaNode) NeedsUpdate() bool {

	timeNow := time.Now()

	if (math.Abs(node.currentValue-node.previousValue) >= node.AbsDeviation) || (timeNow.Sub(node.lastUpdate) >= node.maxTimeInterval) {
		return true
	}
	return false
}

func (node *opcuaNode) UpdateLastUpdate() {
	node.lastUpdate = time.Now()
}
