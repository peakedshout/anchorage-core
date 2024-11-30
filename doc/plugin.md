# Anchorage
## Plugin
### The Plugin module is a submodule of the `client` module and is used to extend some behaviors of the submodule. (The plug-in module is currently relatively unstable, and the design may change frequently for better service)
- `name: "" # plugin name <string>`
  - The name of the plugin.
- `type: [] # plugin type (dial,listen,proxy) <[]string>`
  - The service type of the plugin.
- `list: [] # list plugin setting <[]map[string]interface {}>`
  - Plugin list.
  - The list of plug-ins will take effect from back to front (bottom up). For example, if the `dns`-`socks` plug-in is set, then the entry connection will be processed in the order of `socks`->`dns`.
  - If the order of the plug-ins is messed up, the referenced module may fail to start. For example, if the plug-in order is set to `socks`-`dns`, the module will not be started correctly when the export network is not specified, because the `dns` module does not have the function of intercepting and listening.
### Template configuration (comments are optional)
```
name: "" # plugin name <string>
type: [] # plugin type (dial,listen,proxy) <[]string>
list: [] # list plugin setting <[]map[string]interface {}>
```
### Currently supports some plugin configurations
- `dns`
  - type: `dial` `proxy`
  - ```
    name: dns # plugin name <string> // Plugin name
    network: "" # dns server network <string> // Dns network type optional udp,tcp,tls,url,url-q
    dnsServer: "" # dns server <string> // Dns server address
    insecureTls: false # dns tls insecure skip verify <bool> // Whether to ignore tls check
    preV6: false # enable and precedence ipv6 <bool> // Whether to enable and prioritize ipv6
    ```
  - `network`
    - `udp`,`tcp`,`tls` use the corresponding network type to request dns.
    - `url` initiates `http` request to dns through `url`, which is `doh`; `url-q` requests dns through `http3`.
- `geo_router`
  - type: `dial` `proxy`
  - ```
    name: geo_router # plugin name <string> // Plugin name
    default: "" # default method <string> // Default mode
    proxy: [] # proxy list <[][2]string> // [0]Type [1]Area:Regular
    direct: [] # direct list <[][2]string> // [0]Type [1]Area:Regular
    block: [] # block list <[][2]string> // [0]Type [1]Area:Regular
    geoipPath: "" # geoip.dat path <string> // geoip file path
    geositePath: "" # geosite.dat path <string> // geosite file path
    ```
  - The value of `default` is `proxy`, `direct`, `block`, which is the default behavior. Leave blank to use `proxy`.
  - `proxy`, `direct`, `block` are lists to be matched specifically.
    - In `[2]string`, the first element specifies the type, whether it is `geoip` or `geosite`.
    - The second element is a specified area and a regular or single specified area, where the regular match is the `attribute` that matches the `domain` of the `geosite` file.
- `router`
  - type: `dial` `proxy`
  - ```
    name: router # plugin name <string> // Plugin name
    default: "" # default method <string> // Default mode
    proxy: [] # proxy list <[][2]string> // [0]Network type regularity [1]Network address regularity
    direct: [] # direct list <[][2]string> // [0]Network type regularity [1]Network address regularity
    block: [] # block list <[][2]string> // [0]Network type regularity [1]Network address regularity
    ```
  - The value of `default` is `proxy`, `direct`, `block`, which is the default behavior. Leave blank to use `proxy`.
  - `proxy`, `direct`, `block` are lists to be matched specifically.
    - In `[2]string`, the first element specifies the network type regularity, for example `tcp\d?$` is used to match `tcp`, `tcp4`, `tcp6`.
    - The second element specifies a regular network address, for example, `^10.` is used to match network addresses starting with `10.`.
- `httpproxy`
  - type: `dial` `proxy`
  - ```
    name: "" # plugin name <string> // Plugin name
    Auth: [] # Auth list <[][2]string> // Certification list
    OSProxySettings: false # OS proxy settings <bool> // System proxy settings
    ```
  - The list of `Auth` is a list of username + password. The first element is the username and the second element is the password.
  - `OSProxySettings` Whether to enable setting the operating system's proxy as this proxy (currently only supports windows).
- `sokcs`
  - type: `dial` `proxy`
  - ```
    name: socks # plugin name <string> // Plugin name
    v4: true # socks4 ? <bool> // Whether to enable sock4/4a
    v5: true # socks5 ? <bool> // Whether to enable socks5
    CMDCONNECT: true # CMDCONNECT ? <bool> // Whether to enable CONNECT
    CMDBIND: false # CMDBIND ? <bool> // Whether to enable BIND
    CMDUDPASSOCIATE: false # UDPASSOCIATE ? <bool> // Whether to enable UDPASSOCIATE
    S4Auth: [] # S4Auth list <[]string> // sock4/4a certification list
    S5Auth: [] # S5Auth list <[][2]string> // sock5 certification list
    args: {} # plugin args <map[string]string>
    ```
  - Currently only `CONNECT` mode is supported.
  - `S5Auth` only supports user password authentication mode.


