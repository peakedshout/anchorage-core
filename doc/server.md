# Anchorage
## Server
### The Server module is mainly responsible for forwarding traffic and does not directly participate in business. Proper configuration of `server` is the cornerstone of the entire workflow.
- `enable: false # loaded then to work <bool>`
  - When enable is configured, the `server` module will be started when the `anchorage` service starts. When a `server` module that cannot be started appears, the entire `anchorage` service will fail to start.
- `config:  # server config <*config.ServerConfig>`
  - This is the configuration of the `server` module.
- `nodeInfo: # current server node config <config.NodeConfig>`
  - These are some configurations of the `server` module of the current node. When starting the module, the listening service will be based on these configurations.
- ` nodeName: "" # must be <string>`
  - The `server` module name of the current node, required.
- ```
  baseNetwork: # base network list <[]config.BaseNetworkConfig>
    - network: "" # must be tcp,udp(quic) <string>
      address: "" # work address <string>
  ```
  - Configure the listening network type of the `server` module of the current node. It supports `tcp` and `udp(quic)`, and multiple listening addresses, which means that the `server` module will listen to multiple addresses at the same time to provide services.
- ```
  exNetworks: # expand network list <[]config.ExNetworkConfig>
    - network: "" # must be tcp,udp,quic,tls,ws,wss,http,https; if tcp or udp will without; tls, wss and https must have cert; if ws and http has cert will up grader to wss or https; quic optional cert. <string>
      certFile: "" # cert file path <string>
      keyFile: "" # key file path <string>
      certRaw: "" # raw cert <string>
      keyRaw: "" # raw key <string>
      insecureSkipVerify: false # cert insecure skip verify <bool>
  ```
  - Configure the additional listening network type of the current node, supporting `tcp, udp, quic, tls, ws, wss, http, https`. Multiple listening configurations mean that the `server` module will be nested hierarchically.
  - For example, if `tcp` is configured in `baseNetwork`, `wss` is configured in `exNetworks`, and `tls` is configured, it means that the protocol `tcp->wss->tls` will be used in the network flow.
  - Setting too many additional network types will result in inefficient network transmission, but the encryption camouflage effect will be better.
- ```
  crypto: # crypto config <[]config.CryptoConfig>
    - name: "" # crypto expand name <string>
      crypto: "" # crypto type <string>
      keyFiles: [] # key flies; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      keys: [] # raw key; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      priority: 0 # crypto priority <int8>
  ```
  - Configure the encryption type of the current node. Multiple encryption configurations mean that the encryption method can be provided. When communicating with the peer, one of the encryption methods will be selected to encrypt the signal.
- `handshakeTimeout: 0 # handshake timeout (unit ms) <uint>`
  - Timeout when establishing a connection.
- `handleTimeout: 0 # handle timeout (unit ms) <uint>`
  - Timeout period for processing business.
- ```
  auth: # node auth ( username  password ) <*config.AuthInfo>
    username: ""
    password: ""
  ```
  - Basic user password verification.
- ```
  syncNodes: # sync server node config <[]config.NodeConfig>
    - nodeName: "" # must be <string>
      baseNetwork: [] # base network list <[]config.BaseNetworkConfig>
      exNetworks: [] # expand network list <[]config.ExNetworkConfig>
      crypto: [] # crypto config <[]config.CryptoConfig>
      handshakeTimeout: 0 # handshake timeout (unit ms) <uint>
      handleTimeout: 0 # handle timeout (unit ms) <uint>
      auth: null # node auth ( username  password ) <*config.AuthInfo>
  ```
  - The node information that needs to be synchronized has the same meaning as the configuration field of `nodeInfo`.
  - Note that the current node configuration of `syncNodes` means synchronizing its own information to the opposite `server`, rather than pulling the information of the opposite `server`. This is a one-way synchronization mechanism.
- `syncTimeInterval: 0 # sync server time interval (unit ms) <uint>`
  - The interval for synchronizing information.
- `linkTimeout: 0 # link time out (unit ms) <uint>`
  - Routing connection service timeout.
- `proxyMulti: 0 # whether support proxy multi io count to link <int>`
  - Number of multiplexes used by proxy services.
### Template configuration (comments are optional)
```
enable: false # loaded then to work <bool>
config: # server config <*config.ServerConfig>
    nodeInfo: # current server node config <config.NodeConfig>
        nodeName: "" # must be <string>
        baseNetwork: # base network list <[]config.BaseNetworkConfig>
            - network: "" # must be tcp,udp(quic) <string>
              address: "" # work address <string>
#        exNetworks: # expand network list <[]config.ExNetworkConfig>
#            - network: "" # must be tcp,udp,quic,tls,ws,wss,http,https; if tcp or udp will without; tls, wss and https must have cert; if ws and http has cert will up grader to wss or https; quic optional cert. <string>
#              certFile: "" # cert file path <string>
#              keyFile: "" # key file path <string>
#              certRaw: "" # raw cert <string>
#              keyRaw: "" # raw key <string>
#              insecureSkipVerify: false # cert insecure skip verify <bool>
#        crypto: # crypto config <[]config.CryptoConfig>
#            - name: "" # crypto expand name <string>
#              crypto: "" # crypto type <string>
#              keyFiles: [] # key flies; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
#              keys: [] # raw key; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
#              priority: 0 # crypto priority <int8>
#        handshakeTimeout: 0 # handshake timeout (unit ms) <uint>
#        handleTimeout: 0 # handle timeout (unit ms) <uint>
#        auth: # node auth ( username  password ) <*config.AuthInfo>
#            username: ""
#            password: ""
#    syncNodes: # sync server node config <[]config.NodeConfig>
#        - nodeName: "" # must be <string>
#          baseNetwork: [] # base network list <[]config.BaseNetworkConfig>
#          exNetworks: [] # expand network list <[]config.ExNetworkConfig>
#          crypto: [] # crypto config <[]config.CryptoConfig>
#          handshakeTimeout: 0 # handshake timeout (unit ms) <uint>
#          handleTimeout: 0 # handle timeout (unit ms) <uint>
#          auth: null # node auth ( username  password ) <*config.AuthInfo>
#    syncTimeInterval: 0 # sync server time interval (unit ms) <uint>
#    linkTimeout: 0 # link time out (unit ms) <uint>
#    proxyMulti: 0 # whether support proxy multi io count to link <int>
```