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
	ServerName string      `toml:"ServerName"`
	URL        string      `toml:"URL"`
	Nodes      []opcuaNode `toml:"Nodes"`
	ctx        context.Context
	client     *gopcua.Client
	ID         []*ua.NodeID
	ReadValID  []*ua.ReadValueID
	req        *ua.ReadRequest
}

// Add this plugin to telegraf
func (o *OPCUA) Init() error {

	//var err error

	o.ctx = context.Background()
	// This version doesn't support security yet
	o.client = gopcua.NewClient(o.URL, gopcua.SecurityMode(ua.MessageSecurityModeNone))

	log.Print("Starting opcua plugin to monitor: ", o.URL)

	if err := o.client.Connect(o.ctx); err != nil {
		log.Print("Fatal error in client connect")
		log.Fatal(err)
	}

	for i := range o.Nodes {
		tempID, err := ua.ParseNodeID(o.Nodes[i].NodeID)
		if err != nil {
			log.Fatalf("invalid node id: %v", err)
		}
		o.ID = append(o.ID, tempID)
		o.ReadValID = append(o.ReadValID, &ua.ReadValueID{NodeID: tempID})

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
		log.Fatalf("Read failed: %s", err)
	}

	for i := range resp.Results {
		if resp.Results[i].Status != ua.StatusOK {
			log.Fatalf("Status not OK: %v", resp.Results[i].Status)
		}
		fields := make(map[string]interface{})
		tags := make(map[string]string)

		tags["server"] = o.ServerName
		tags["tag"] = o.Nodes[i].Tag
		fields["value"] = resp.Results[i].Value.Float()

		acc.AddFields("opcua", fields, tags, resp.Results[i].SourceTimestamp)
	}

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
  ServerName = "Device"
  ## URL including endpoint
  URL = "http://localhost.com:4840/endpoint"
  ##
  ## List of Nodes to monitor

  Nodes = [{Tag = "Tag1", NodeID = "ns=1;s=the.answer"},
  {Tag = "Tag2", NodeID = "ns=1;i=51028"}
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
	Tag    string `toml:"Tag"`
	NodeID string `toml:"NodeID"`
}
