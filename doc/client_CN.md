# Anchorage
## Client
### Client 模块主要连接`server`模块，管理多个子模块，是直接参与业务的模块。
- `enable: false # loaded then to work <bool>`
  - 当配置enable时，将在`anchorage`服务启动时一并启动`client`模块，当出现无法启动的`client`模块时，将导致整个`anchorage`服务启动失败。
- `config: # client config <*config.ClientConfig>`
    - 这是`client`模块的配置。
- `nodes: # dial to server node list <[]config.NodeConfig>`
  - 这是配置`client`模块需要连接的`server`集群，当配置多个节点时，在处理业务时会根据对应的请求选择接入点。
- `nodeName: "" # must be <string>`
  - 连接`server`节点的名称，应该与对端的配置保持一致。
- ```
  baseNetwork: # base network list <[]config.BaseNetworkConfig>
    - network: "" # must be tcp,udp(quic) <string>
      address: "" # work address <string>
  ```
  - 配置连接的`server`模块的监听网络类型，支持`tcp`和`udp(quic)`，多个监听地址，意味着该`server`模块会同时监听多个地址进行服务，应该与对端的配置保持一致。
- ```
  exNetworks: # expand network list <[]config.ExNetworkConfig>
    - network: "" # must be tcp,udp,quic,tls,ws,wss,http,https; if tcp or udp will without; tls, wss and https must have cert; if ws and http has cert will up grader to wss or https; quic optional cert. <string>
      certFile: "" # cert file path <string>
      keyFile: "" # key file path <string>
      certRaw: "" # raw cert <string>
      keyRaw: "" # raw key <string>
      insecureSkipVerify: false # cert insecure skip verify <bool>
  ```
  - 配置连接`server`模块的附加监听网络类型，支持`tcp,udp,quic,tls,ws,wss,http,https`，多个监听配置，意味着该`server`模块会进行层级进行嵌套。
  - 例如，在`baseNetwork`中配置了`tcp`，`exNetworks`中配置了`wss`后又配置了`tls`，这意味着会在网络流中使用`tcp->wss->tls`的协议。
  - 设置过多的附加网络类型会导致网络传输效率不高，但加密伪装效果会更好。
  - 应该与对端的配置保持一致。
- ```
  crypto: # crypto config <[]config.CryptoConfig>
    - name: "" # crypto expand name <string>
      crypto: "" # crypto type <string>
      keyFiles: [] # key flies; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      keys: [] # raw key; if crypto is asymmetry, first is cert then key; if crypto is symmetry, only key. <[]string>
      priority: 0 # crypto priority <int8>
  ```
  - 配置当前节点的加密类型，多个加密配置，意味着能够提供那种加密方式，与对端通信时，将选取其中一个加密方式加密信号。
  - 应该与对端的配置保持一致。
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
### 模板的配置（注释部位为非必填项）
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

