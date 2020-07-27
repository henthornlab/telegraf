package opcua

import (
	"context"
	"log"

	"github.com/gopcua/opcua"
	gopcua "github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// OPCUA : Structure for all the plugin info
type OPCUA struct {
	ServerName    string      `toml:"ServerName"`
	URL           string      `toml:"URL"`
	Nodes         []opcuaNode `toml:"Nodes"`
	Authorization string      `toml:"Authorization"`
	Username      string      `toml:"Username"`
	Password      string      `toml:"Password"`
	ctx           context.Context
	client        *gopcua.Client
	ID            []*ua.NodeID
	ReadValID     []*ua.ReadValueID
	req           *ua.ReadRequest
}

// Init : function to intialize the plugin
func (o *OPCUA) Init() error {

	var authOption gopcua.Option
	opts := []opcua.Option{}

	// Need to determine how the user wishes to connect
	switch o.Authorization {
	case "anonymous":
		authOption = opcua.AuthAnonymous()

	case "user-password":
		authOption = opcua.AuthUsername(o.Username, o.Password)

	default:
		log.Print("opcua: No valid authorization chosen... defaulting to anonymous")
		authOption = opcua.AuthAnonymous()
	}

	opts = append(opts, authOption)
	opts = append(opts, gopcua.SecurityMode(ua.MessageSecurityModeNone))

	o.ctx = context.Background()
	// This version doesn't support security yet
	//o.client = gopcua.NewClient(o.URL, gopcua.SecurityMode(ua.MessageSecurityModeNone))
	o.client = gopcua.NewClient(o.URL, opts...)

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

		tags["server"] = o.ServerName
		tags["tag"] = o.Nodes[i].Tag
		fields["value"] = resp.Results[i].Value.Float()

		acc.AddFields("opcua", fields, tags, resp.Results[i].SourceTimestamp)
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
  ## URL including endpoint
  URL = "http://localhost.com:4840/endpoint"
  ## Select authorization mode. Either "anonymous" or "user-password"
  ## Be sure to provide a username/password if selecting "user-password"
  Authorization = "anonymous"
  # Username = "foo"
  # Password = "bar"

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
