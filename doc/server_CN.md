# Anchorage
## Server
### Server 模块主要负责流量的转发，不直接参与业务，合理的配置`server`是整个工作流的基石。
- `enable: false # loaded then to work <bool>`
  - 当配置enable时，将在`anchorage`服务启动时一并启动`server`模块，当出现无法启动的`server`模块时，将导致整个`anchorage`服务启动失败。
- `config:  # server config <*config.ServerConfig>`
  - 这是`server`模块的配置。
- `nodeInfo: # current server node config <config.NodeConfig>`
  - 这是当前节点的`server`模块的一些配置，在启动模块时，将根据这些配置进行监听服务。
- ` nodeName: "" # must be <string>`
  - 当前节点的`server`模块名称，必填。
- ```
  baseNetwork: # base network list <[]config.BaseNetworkConfig>
    - network: "" # must be tcp,udp(quic) <string>
      address: "" # work address <string>
  ```
  - 配置当前节点的`server`模块的监听网络类型，支持`tcp`和`udp(quic)`，多个监听地址，意味着该`server`模块会同时监听多个地址进行服务。
- ```
  exNetworks: # expand network list <[]config.ExNetworkConfig>
    - network: "" # must be tcp,udp,quic,tls,ws,wss,http,https; if tcp or udp will without; tls, wss and https must have cert; if ws and http has cert will up grader to wss or https; quic optional cert. <string>
      certFile: "" # cert file path <string>
      keyFile: "" # key file path <string>
      certRaw: "" # raw cert <string>
      keyRaw: "" # raw key <string>
      insecureSkipVerify: false # cert insecure skip verify <bool>
  ```
  - 配置当前节点的附加监听网络类型，支持`tcp,udp,quic,tls,ws,wss,http,https`，多个监听配置，意味着该`server`模块会进行层级进行嵌套。
  - 例如，在`baseNetwork`中配置了`tcp`，`exNetworks`中配置了`wss`后又配置了`tls`，这意味着会在网络流中使用`tcp->wss->tls`的协议。
  - 设置过多的附加网络类型会导致网络传输效率不高，但加密伪装效果会更好。
- ```
  crypto: # crypto config <[]config.CryptoConfig>
    - name: "" # crypto expand name <string>
      crypto: "" # crypto type <string>
      keyFiles: [] # key flies; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      keys: [] # raw key; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      priority: 0 # crypto priority <int8>
  ```
  - 配置当前节点的加密类型，多个加密配置，意味着能够提供那种加密方式，与对端通信时，将选取其中一个加密方式加密信号。
- `handshakeTimeout: 0 # handshake timeout (unit ms) <uint>`
  - 在建立连接时的超时时间。
- `handleTimeout: 0 # handle timeout (unit ms) <uint>`
  - 处理业务的超时时间。
- ```
  auth: # node auth ( username  password ) <*config.AuthInfo>
    username: ""
    password: ""
  ```
  - 基础的用户密码验证。
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
  - 需要同步的节点信息，与`nodeInfo`的配置字段含义相同。
  - 注意，当前节点配置`syncNodes`意味着将自身的信息同步给对端`server`，而不是拉去对端的`server`的信息，这是一种单向的同步机制。
- `syncTimeInterval: 0 # sync server time interval (unit ms) <uint>`
  - 同步信息的间隔。
- `linkTimeout: 0 # link time out (unit ms) <uint>`
  - 路由连接业务超时时间。
- `proxyMulti: 0 # whether support proxy multi io count to link <int>`
  - 代理业务使用的多路复用数。
### 模板的配置（注释部位为非必填项）
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