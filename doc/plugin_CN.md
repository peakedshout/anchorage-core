# Anchorage
## Plugin
### Plugin 模块是`client`模块的一个子模块，用于拓展子模块的一些行为。（目前插件模块比较不稳定，设计可能会经常变动以求更好的服务）
- `name: "" # plugin name <string>`
  - 该插件的名称。
- `type: [] # plugin type (dial,listen,proxy) <[]string>`
  - 插件的服务类型。
- `list: [] # list plugin setting <[]map[string]interface {}>`
  - 插件列表。
  - 插件的列表会从后向前（自下而上）的生效，例如设置了`dns`-`socks`插件，那么入口连接将从`socks`->`dns`的顺序被处理。
  - 如果插件的顺序混乱可能会导致引用模块无法启动，例如设置了`socks`-`dns`的插件顺序，在不指定出口网络时，将无法正确启动该模块，因为`dns`模块并没有截取监听的功能。
### 模板的配置（注释部位为非必填项）
```
name: "" # plugin name <string>
type: [] # plugin type (dial,listen,proxy) <[]string>
list: [] # list plugin setting <[]map[string]interface {}>
```
### 目前支持一些插件配置
- `dns`
  - type: `dial` `proxy`
  - ```
    name: dns # plugin name <string> // 插件名称
    network: "" # dns server network <string> // dns网络类型 可选 udp,tcp,tls,url,url-q
    dnsServer: "" # dns server <string> // dns 服务器地址
    insecureTls: false # dns tls insecure skip verify <bool> // 是否忽略tls校验
    preV6: false # enable and precedence ipv6 <bool> // 是否启用并优先ipv6
    ```
  - `network`
    - `udp`,`tcp`,`tls` 则是使用对应的网络类型请求dns。
    - `url` 是通过`url`发起`http`请求dns，也就是`doh`；`url-q`是通过`http3`请求dns。
- `geo_router`
  - type: `dial` `proxy`
  - ```
    name: geo_router # plugin name <string> // 插件名称
    default: "" # default method <string> // 默认模式
    proxy: [] # proxy list <[][2]string> // [0]类型 [1]区域:正则
    direct: [] # direct list <[][2]string> // [0]类型 [1]区域:正则
    block: [] # block list <[][2]string> // [0]类型 [1]区域:正则
    geoipPath: "" # geoip.dat path <string> // geoip 文件路径
    geositePath: "" # geosite.dat path <string> // geosite 文件路径
    ```
  - `default`取值为`proxy`,`direct`,`block`，即默认的行为。留空则为`proxy`。
  - `proxy`,`direct`,`block`，是要具体的匹配的列表。
    - `[2]string`中，第一个元素指定类型，是`geoip`还是`geosite`。
    - 第二个元素是指定区域和正则或单指定区域，其中正则式匹配是匹配`geosite`文件的`domain`的`attribute`。
- `router`
  - type: `dial` `proxy`
  - ```
    name: router # plugin name <string> // 插件名称
    default: "" # default method <string> // 默认模式
    proxy: [] # proxy list <[][2]string> // [0]网络类型正则 [1]网络地址正则
    direct: [] # direct list <[][2]string> // [0]网络类型正则 [1]网络地址正则
    block: [] # block list <[][2]string> // [0]网络类型正则 [1]网络地址正则
    ```
  - `default`取值为`proxy`,`direct`,`block`，即默认的行为。留空则为`proxy`。
  - `proxy`,`direct`,`block`，是要具体的匹配的列表。
    - `[2]string`中，第一个元素指定网络类型正则，例如`tcp\d?$`用来匹配`tcp`,`tcp4`,`tcp6`。
    - 第二个元素是指定网络地址正则，例如`^10.`用来匹配`10.`开头的网络地址。
- `httpproxy`
  - type: `dial` `proxy`
  - ```
    name: "" # plugin name <string> // 插件名称
    Auth: [] # Auth list <[][2]string> // 认证列表
    OSProxySettings: false # OS proxy settings <bool> // 系统代理设置
    ```
  - `Auth`的列表是用户名+密码的列表，第一个元素为用户名，第二个元素为密码。
  - `OSProxySettings` 是否启用将操作系统的代理进行设置为该代理（目前仅支持windows）。
- `sokcs`
  - type: `dial` `proxy`
  - ```
    name: socks # plugin name <string> // 插件名称
    v4: true # socks4 ? <bool> // 是否启用sock4/4a
    v5: true # socks5 ? <bool> // 是否启用sock5
    CMDCONNECT: true # CMDCONNECT ? <bool> // 是否启用CONNECT
    CMDBIND: false # CMDBIND ? <bool> // 是否启用BIND
    CMDUDPASSOCIATE: false # UDPASSOCIATE ? <bool> // 是否启用UDPASSOCIATE
    S4Auth: [] # S4Auth list <[]string> // sock4/4a 认证列表
    S5Auth: [] # S5Auth list <[][2]string> // sock5 认证列表
    args: {} # plugin args <map[string]string>
    ```
  - 目前只支持`CONNECT`模式。
  - `S5Auth`只支持用户密码的认证模式。


