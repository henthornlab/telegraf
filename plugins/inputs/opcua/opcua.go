package opcua

import (
	"context"
	"log"

	gopcua "github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Structure for all the plugin info
type OPCUA struct {
	ServerName string   `toml:"ServerName"`
	URL        string   `toml:"URL"`
	Nodes      []string `toml:"Nodes"`
	ctx        context.Context
	client     *gopcua.Client
	ID         *ua.NodeID
	req        *ua.ReadRequest
}

// Add this plugin to telegraf
func (o *OPCUA) Init() error {

	var err error
	o.ctx = context.Background()

	o.client = gopcua.NewClient(o.URL, gopcua.SecurityMode(ua.MessageSecurityModeNone))

	log.Print("Starting opcua plugin to monitor:", o.URL)
	for i := range o.Nodes {
		log.Print(o.Nodes[i])
	}

	if err := o.client.Connect(o.ctx); err != nil {
		log.Print("Fatal error in client connect")
		log.Fatal(err)
	}

	log.Print("Parsing nodeID")
	o.ID, err = ua.ParseNodeID("ns=1;s=the.answer")
	if err != nil {
		log.Fatalf("invalid node id: %v", err)
	}

	log.Print("Building ReadRequest")
	o.req = &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			&ua.ReadValueID{NodeID: o.ID}},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	return nil
}

func init() {
	inputs.Add("opcua", func() telegraf.Input { return &OPCUA{} })

}

// Gather implements the telegraf plugin interface method for data accumulation
func (o *OPCUA) Gather(acc telegraf.Accumulator) error {

	log.Print("In Gather() and attempting read")

	resp, err := o.client.Read(o.req)

	if err != nil {
		log.Fatalf("Read failed: %s", err)
	}
	if resp.Results[0].Status != ua.StatusOK {
		log.Fatalf("Status not OK: %v", resp.Results[0].Status)
	}
	log.Printf("%#v", resp.Results[0].Value.Value())

	fields := make(map[string]interface{})
	tags := make(map[string]string)

	fields["ns1"] = resp.Results[0].Value.Value()
	tags["server"] = o.ServerName

	acc.AddFields("Answer", fields, tags)

	return nil
}

const description = `Connect to OPC-UA Server`
const sampleConfig = `
  ## OPC-UA Connection Configuration
  ##
  ## The plugin designed to connect to OPC UA devices
  ## Currently supports anonymous mode only
  ##
  ## Name given to OPC UA server for logging and tags
  name = "Device"
  ## URL including endpoint
  URL = "http://localhost.com:4840/endpoint"
  ##
  ## List of Nodes to monitor
  ## List of Nodes to monitor

  Nodes = ["ns=1;s=the.answer","ns=1;i=2345"]
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
	name   string `toml:"Name"`
	nodeID string `toml:"NodeID"`
}
