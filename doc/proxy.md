# Anchorage
## Proxy
### The Proxy module is a submodule of the `client` module and is a business module that proxies traffic to the exit of the `server` module.
- `enable: false # loaded then to work <bool>`
  - When enable is configured, the `client` module will start the `proxy` module when the `anchorage` service starts. When a `proxy` module cannot be started, it will cause the entire `anchorage` service to fail to start.
- `node: [] # node links <[]string>`
  - Specify the connection route. If not specified, it will be determined internally by `client` and `server`.
- ```
  inNetwork: # in network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
  ```
  - The entrance network, after being specified, will listen to the service address and connect the service connection with the peer.
- ```
  outNetwork: # out network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
  ```
  - Egress network, if the egress network is specified, the egress traffic of the opposite end will be directed to this address; if the egress network is not specified, it will be determined by the plug-in. If neither is specified, the module will not be started.
- `multi: 0 # whether support multi io count to link <int>`
  - The maximum number of multiplexed connections for multiplexing.
- `plugin: "" # plugin name <string>`
  - Plugin.
### Template configuration (comments are optional)
```
enable: false # loaded then to work <bool>
#node: [] # node links <[]string>
inNetwork: # in network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
#outNetwork: # out network config <*sdk.NetworkConfig>
#    network: "" # must be tcp or udp <string>
#    address: "" # must be tcp or udp <string>
#multi: 0 # whether support multi io count to link <int>
#plugin: "" # plugin name <string>
```

