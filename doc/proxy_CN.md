# Anchorage
## Proxy
### Proxy 模块是`client`模块的一个子模块，是代理流量到`server`模块出口的业务模块。
- `enable: false # loaded then to work <bool>`
  - 当配置enable时，将在`anchorage`服务启动时`client`模块会一并启动该`proxy`模块，当出现无法启动的`proxy`模块时，将导致整个`anchorage`服务启动失败。
- `node: [] # node links <[]string>`
  - 指定连接路线，如果不指定，将由`client`和`server`内部自行决定。
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
- `multi: 0 # whether support multi io count to link <int>`
  - 多路复用的最大复用连接数。
- `plugin: "" # plugin name <string>`
  - 插件。
### 模板的配置（注释部位为非必填项）
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

