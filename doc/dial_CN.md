# Anchorage
## Dial
### Dial 模块是`client`模块的一个子模块，是将自身服务连接到已经注册在`server`上的`listen`模块的业务模块。
- `enable: false # loaded then to work <bool>`
  - 当配置enable时，将在`anchorage`服务启动时`client`模块会一并启动该`dial`模块，当出现无法启动的`dial`模块时，将导致整个`anchorage`服务启动失败。
- `node: [] # node links <[]string>`
  - 指定连接路线，如果不指定，将由`client`和`server`内部自行决定。
- `PP2PNetwork: "" # specify the p2p network type <string>`
  - 指定优先使用的p2p网络类型。
- `switchUP2P: false # whether support udp p2p to link <bool>`
  - 是否启用udp p2p连接。
- `switchTP2P: false # whether support tcp p2p to link <bool>`
  - 是否启用tcp p2p连接。
- `forceP2P: false # whether force p2p to link <bool>`
  - 是否强制使用p2p网络。
- ```
  auth: # node auth ( username  password ) <*config.AuthInfo>
    username: ""
    password: ""
  ```
  - 基础的用户密码验证。
- ```
  inNetwork: # in network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
  ```
  - 入口网络，指定后，将监听服务该地址，并将服务的连接与对端进行对接。
- ```
  outNetwork: # out network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
  ```
  - 出口网络，如果指定了出口网络，那么对端的出口流量将指向该地址；如果不指定出口网络，那么就会由对插件决定，如果都不指定将无法启动模块。
- `multi: 0 # whether support multi io count to link (set -1 to close) <int>`
  - 多路复用的最大复用连接数，设置为-1时将取消多路复用。注意，当对端与本端的多路复用特性不一致时，将无法对接。
- `multiIdle: 0 # whether support multi io idle count to link <int>`
  - 指定多路复用空闲保持的连接数。（保持空闲连接，能给更快启动）
- `plugin: "" # plugin name <string>`
  - 插件。
### 模板的配置（注释部位为非必填项）
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

