# OPC UA Input Plugin

Allows for the collection of metrics from an OPC UA server. Currently supports anonymous (no) authentication. Certificates are not supported at this time.

To use this plugin you will need the following:

1. Knowledge of your OPC UA server's address, port, and endpoint. For example: opc.tcp://server.com:4840/endpoint
2. Knowledge of the Node IDs of the tags you wish to monitor. For example: ns=1;i=51028

It is recommended to download a high quality OPC UA client (Prosys, UA Expert, etc.) which will allow you to browse your server. Record the Node IDs of the desired tags and enter them in the config portion of telegraf.

## Configuration

```toml
  ## OPC-UA Connection Configuration
  ##
  ## The plugin designed to connect to OPC UA devices
  ## Currently supports anonymous mode only
  ##
  ## Name given to OPC UA server for logging and tags
  ServerName = "Device"
  ## URL including endpoint
  URL = "opc.tcp://localhost.com:4840/endpoint"
  ## Select authorization mode. Either "anonymous" or "user-password"
  ## Be sure to provide a username/password if selecting "user-password"
  Authorization = "anonymous"
  # Username = "foo"
  # Password = "bar"

  ## List of Nodes to monitor including Tag (name), NodeID, and the absolute Deadband
  Nodes = [
  {Tag = "HeatExchanger1 Temp", NodeID = "ns=2;s=TE-800-07/AI1/PV.CV", AbsDeadband = 0.10},
  {Tag = "Heat Exchanger1 Pressure", NodeID = "ns=2;i=1234", AbsDeadband = 0.10},
  ]
```

### Tested Configurations

Open62541 Test server (https://github.com/open62541/open62541) in anonymous mode
DeltaV 14.3.1 in anonymous mode
