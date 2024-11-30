# Anchorage
## Client
### The Client module is mainly connected to the `server` module, manages multiple sub-modules, and is the module directly involved in the business.
- `enable: false # loaded then to work <bool>`
  - When enable is configured, the `client` module will be started when the `anchorage` service is started. If the `client` module cannot be started, the entire `anchorage` service will fail to start.
- `config: # client config <*config.ClientConfig>`
  - This is the configuration for the `client` module.
- `nodes: # dial to server node list <[]config.NodeConfig>`
  - This is the `server` cluster that the `client` module needs to connect to. When multiple nodes are configured, the access point will be selected according to the corresponding request when processing business.
- `nodeName: "" # must be <string>`
  - The name of the connected `server` node should be consistent with the configuration of the other end.
- ```
  baseNetwork: # base network list <[]config.BaseNetworkConfig>
    - network: "" # must be tcp,udp(quic) <string>
      address: "" # work address <string>
  ```
  - Configure the listening network type of the connected `server` module, supporting `tcp` and `udp(quic)`. Multiple listening addresses means that the `server` module will listen to multiple addresses for service at the same time, which should be consistent with the configuration of the other end.
- ```
  exNetworks: # expand network list <[]config.ExNetworkConfig>
    - network: "" # must be tcp,udp,quic,tls,ws,wss,http,https; if tcp or udp will without; tls, wss and https must have cert; if ws and http has cert will up grader to wss or https; quic optional cert. <string>
      certFile: "" # cert file path <string>
      keyFile: "" # key file path <string>
      certRaw: "" # raw cert <string>
      keyRaw: "" # raw key <string>
      insecureSkipVerify: false # cert insecure skip verify <bool>
  ```
  - Configure additional listening network types connected to the `server` module, supporting `tcp,udp,quic,tls,ws,wss,http,https`. Multiple listening configurations mean that the `server` module will be nested hierarchically.
  - For example, if `tcp` is configured in `baseNetwork`, `wss` is configured in `exNetworks`, and `tls` is configured, it means that the protocol `tcp->wss->tls` will be used in the network flow.
  - Setting too many additional network types will result in inefficient network transmission, but the encryption camouflage effect will be better.
  - It should be consistent with the configuration of the peer.
- ```
  crypto: # crypto config <[]config.CryptoConfig>
    - name: "" # crypto expand name <string>
      crypto: "" # crypto type <string>
      keyFiles: [] # key flies; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      keys: [] # raw key; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      priority: 0 # crypto priority <int8>
  ```
  - Configure the encryption type of the current node. Multiple encryption configurations mean that the encryption method can be provided. When communicating with the peer, one of the encryption methods will be selected to encrypt the signal.
  - It should be consistent with the configuration of the peer.
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
### Template configuration (comments are optional)
```
enable: false # loaded then to work <bool>
config: # client config <*config.ClientConfig>
    nodes: # dial to server node list <[]config.NodeConfig>
        - nodeName: "" # must be <string>
          baseNetwork: # base network list <[]config.BaseNetworkConfig>
            - network: "" # must be tcp,udp(quic) <string>
              address: "" # work address <string>
#          exNetworks: # expand network list <[]config.ExNetworkConfig>
#            - network: "" # must be tcp,udp,quic,tls,ws,wss,http,https; if tcp or udp will without; tls, wss and https must have cert; if ws and http has cert will up grader to wss or https; quic optional cert. <string>
#              certFile: "" # cert file path <string>
#              keyFile: "" # key file path <string>
#              certRaw: "" # raw cert <string>
#              keyRaw: "" # raw key <string>
#              insecureSkipVerify: false # cert insecure skip verify <bool>
#          crypto: # crypto config <[]config.CryptoConfig>
#            - name: "" # crypto expand name <string>
#              crypto: "" # crypto type <string>
#              keyFiles: [] # key flies; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
#              keys: [] # raw key; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
#              priority: 0 # crypto priority <int8>
#          handshakeTimeout: 0 # handshake timeout (unit ms) <uint>
#          handleTimeout: 0 # handle timeout (unit ms) <uint>
#          auth: # node auth ( username  password ) <*config.AuthInfo>
#            username: ""
#            password: ""
```

