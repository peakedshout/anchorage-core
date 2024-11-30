# Anchorage
## Listen
### Listen 模块是`client`模块的一个子模块，是将自身服务注册到`server`模块的业务模块。
- `enable: false # loaded then to work <bool>`
  - 当配置enable时，将在`anchorage`服务启动时`client`模块会一并启动该`listen`模块，当出现无法启动的`listen`模块时，将导致整个`anchorage`服务启动失败。
- `node: "" # register a specified node <string>`
  - 指定该`listen`模块将注册到哪个`server`模块上。
- `name: "" # service name <string>`
  - 注册到`server`到名称，用于路由的定位。
- `notes: "" # service notes <string>`
  - `listen`模块的备注说明。
- ```
  auth: # node auth ( username  password ) <*config.AuthInfo>
    username: ""
    password: ""
  ```
  - 基础的用户密码验证。
- `switchHide: false # not to be discovered by others <bool>`
  - 该注册信息是否隐藏。
- `switchLink: true # whether or not to allow the link <bool>`
  - 是否可被连接。
- `switchUP2P: false # whether support udp p2p to link <bool>`
  - 是否启用udp p2p连接。  
- `switchTP2P: false # whether support tcp p2p to link <bool>`
  - 是否启用tcp p2p连接。
- ```
  outNetwork: # out network config <*sdk.NetworkConfig>
    network: "" # must be tcp or udp <string>
    address: "" # must be tcp or udp <string>
  ```
  - 出口网络，如果指定了出口网络，那么对端的请求就会被约束；如果不指定出口网络，那么就会由对端决定出口网络。
- `multi: true # whether support multi io to link <bool>`
  - 是否启用多路复用（会产生一定的性能开销，但对于复用连接会有很好的快速启动）
- `plugin: "" # plugin name <string>`
  - 插件。
### 模板的配置（注释部位为非必填项）
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


