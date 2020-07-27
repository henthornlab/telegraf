# OPC UA Input Plugin

Allows for the collection of metrics from an OPC UA server. Currently supports anonymous and username/password authentication. Certificates are not supported at this time.

To use this plugin you will need the following:

a. Knowledge of your OPC UA server's address and endpoint. E.g. opc.tcp://server.com/endpoint
b. Knowledge of the Node IDs of the tags you wish to monitor.

To simplify this it is recommended to download a high quality OPC UA client (Prosys, UA Expert, etc.) which will allow you to browse your server. Record the Node IDs of the desired tags.

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
  URL = "http://localhost.com:4840/endpoint"
  ## Select authorization mode. Either "anonymous" or "user-password"
  ## Be sure to provide a username/password if selecting "user-password"
  Authorization = "anonymous"
  # Username = "foo"
  # Password = "bar"

  ## List of Nodes to monitor
  Nodes = [
  {Tag = "Tag1", NodeID = "ns=1;s=the.answer"},
  {Tag = "Tag2", NodeID = "ns=1;i=51028"}
  ]
```

### Tested Configurations

Open62541 Test server (https://github.com/open62541/open62541)
