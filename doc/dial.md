# Anchorage
## Dial
### The Dial module is a submodule of the `client` module. It is a business module that connects its own services to the `listen` module that has been registered on the `server`.
- `enable: false # loaded then to work <bool>`
  - When enable is configured, the `client` module will start the `dial` module when the `anchorage` service starts. When a `dial` module cannot be started, it will cause the entire `anchorage` service to fail to start.
- `node: [] # node links <[]string>`
  - Specify the connection route. If not specified, it will be determined internally by `client` and `server`.
- `PP2PNetwork: "" # specify the p2p network type <string>`
  - Specify the preferred p2p network type.
- `switchUP2P: false # whether support udp p2p to link <bool>`
  - Whether to enable udp p2p connection.
- `switchTP2P: false # whether support tcp p2p to link <bool>`
  - Whether to enable tcp p2p connection.
- `forceP2P: false # whether force p2p to link <bool>`
  - Whether to force the use of p2p network.
- ```
  auth: # node auth ( username  password ) <*config.AuthInfo>
    username: ""
    password: ""
  ```
  - Basic user password verification.
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
- `multi: 0 # whether support multi io count to link (set -1 to close) <int>`
  - The maximum number of multiplexed connections for multiplexing. When set to -1, multiplexing will be cancelled. Note that when the multiplexing characteristics of the peer end and the local end are inconsistent, the connection will not be possible.
- `multiIdle: 0 # whether support multi io idle count to link <int>`
  - Specifies the number of connections to be kept idle by the multiplexer. (Keeping idle connections allows for faster startup)
- `plugin: "" # plugin name <string>`
  - Plugin.
### Template configuration (comments are optional)
```
enable: false # loaded then to work <bool>
#node: [] # node links <[]string>
link: "" # link service name <string>
#PP2PNetwork: "" # specify the p2p network type <string>
#switchUP2P: false # whether support udp p2p to link <bool>
#switchTP2P: false # whether support tcp p2p to link <bool>
#forceP2P: false # whether force p2p to link <bool>
#auth: # service auth ( username  password ) <*config.AuthInfo>
#    username: ""
#    password: ""
inNetwork: # in network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
#outNetwork: # out network config <*sdk.NetworkConfig>
#    network: "" # must be tcp or udp <string>
#    address: "" # must be tcp or udp <string>
#multi: 0 # whether support multi io count to link (set -1 to close) <int>
#multiIdle: 0 # whether support multi io idle count to link <int>
#plugin: "" # plugin name <string>
```

