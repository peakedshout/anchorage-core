# Anchorage
## Listen
### The Listen module is a submodule of the `client` module and is a business module that registers its own services to the `server` module.
- `enable: false # loaded then to work <bool>`
  - When enable is configured, the `client` module will start the `listen` module when the `anchorage` service is started. When a `listen` module cannot be started, it will cause the entire `anchorage` service to fail to start.
- `node: "" # register a specified node <string>`
  - Specify which `server` module this `listen` module will be registered with.
- `name: "" # service name <string>`
  - Register to `server` to name, used for routing location.
- `notes: "" # service notes <string>`
  - Notes on the `listen` module.
- ```
  auth: # node auth ( username  password ) <*config.AuthInfo>
    username: ""
    password: ""
  ```
  - Basic user password verification.
- `switchHide: false # not to be discovered by others <bool>`
  - Whether the registration information is hidden.
- `switchLink: true # whether or not to allow the link <bool>`
  - Can be connected.
- `switchUP2P: false # whether support udp p2p to link <bool>`
  - Whether to enable udp p2p connection.  
- `switchTP2P: false # whether support tcp p2p to link <bool>`
  - Whether to enable tcp p2p connection.
- ```
  outNetwork: # out network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
  ```
  - Egress network, if the egress network is specified, the peer's request will be restricted; if the egress network is not specified, the peer will determine the egress network.
- `multi: true # whether support multi io to link <bool>`
  - Whether to enable multiplexing (there will be a certain performance overhead, but it will have a good quick start for multiplexed connections)
- `plugin: "" # plugin name <string>`
  - Plugin.
### Template configuration (comments are optional)
```
enable: false # loaded then to work <bool>
#node: "" # register a specified node <string>
name: "" # service name <string>
#notes: "" # service notes <string>
#auth: # service auth ( username  password ) <*config.AuthInfo>
#    username: ""
#    password: ""
#switchHide: false # not to be discovered by others <bool>
switchLink: true # whether or not to allow the link <bool>
#switchUP2P: false # whether support udp p2p to link <bool>
#switchTP2P: false # whether support tcp p2p to link <bool>
#outNetwork: # out network config <*sdk.NetworkConfig>
#    network: "" # must be tcp or udp <string>
#    address: "" # must be tcp or udp <string>
multi: true # whether support multi io to link <bool>
#plugin: "" # plugin name <string>
```


