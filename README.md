# Anchorage
#### EN/[CN](./README_CN.md)

## INTRODUCTION
This project is a program used to implement a powerful traffic center, aiming to be [efficient, powerful, multi-functional, and safe].

### Why is there this thing?
- Because I am always tired of adjusting the network, I thought about whether I could make something that would allow me to conveniently switch and control traffic.
- So, `Anchorage` was born.
- Maybe you know [go-CFC](https://github.com/peakedshout/go-CFC), which is the predecessor of `Anchorage`. It summed up experience and expanded it to produce `Anchorage`.
- `Anchorage` has been working for me for a long time and it plays a big role in my daily work.
- So I thought I'd share it and hope someone can use it and make some suggestions.

### What can it do for you?
- Traffic forwarding, nothing more.
- `Anchorage` does not do anything too advanced, but it can accomplish some things that we usually find cumbersome.
- For example, expose an intranet service, divert some local traffic for proxy, etc.
- If you go one step further, you can access the SDK for specialized development.

## CATALOG
- [Installation-instructions](#Installation-instructions)
- [Quick-start](#Quick-start)
- [Build-and-compile](#Build-and-compile)
- [TODO](#TODO)

## Installation-instructions
### System requirements
- Operating system: Windows, macOS, or Linux (Unix)

### Installation-steps
1. Get executable binaries in recent releases:
   ```
   root@ubuntu-s:~/anchorage# ./anchorage_linux_amd64 -v
   anchorage version 
       (_)    
     <--|-->  
    _   |   _ 
   `\__/ \__/`
     `-.o.-'  
    anch'rage 
   anchorage-code:v0.1.0
   ```
   The executable file outputs version-related information to indicate that it is executable, and you can start the anchorage journey!（●´3｀）~♪

## Quick-start
### Start-service
1. We need to initialize a configuration file for service startup
   ```
   Usage:
   anchorage init [config] [flags]
   ```
    - `config` is used to specify the service configuration file to be started. If it is empty, ``~/.anchorage/cfg.yaml`` will be used by default.
2. Anchorage’s command line tool is very simple to start.
   ```
   Usage:
   anchorage start [config] [flags]
   Flags:
       --cmd.addr string   cmd address
       --cmd.cert string   cmd tls cert file
       --cmd.i             cmd tls insecure
       --cmd.key string    cmd tls key file
       --cmd.nk string     cmd network
       --cmd.p string      cmd password
       --cmd.tls           cmd tls enable
       --cmd.u string      cmd username
   -d, --debug string      Enable debug pprof and metrics
   -h, --help              help for start
   -i, --init              If config not exist, will init config
   ```
    - `config` is used to specify the service configuration file to be started. If it is empty, `~/.anchorage/cfg.yaml` will be used by default to start the service.
    - `--cmd.addr` This is the listening address of the background management. By default, `127.0.0.1:2024` is used for monitoring.
    - `--cmd.nk` This is the listening network type of background management. By default, `tcp` is used.
    - `--cmd.u` `--cmd.p` This is the user password for background management, and is empty by default.
    - `--cmd.tls` `--cmd.cert` `--cmd.key` These are some configurations of tls, which are not enabled by default.
    - `-d --debug` This is the function of enabling golang's pprof and metrics. You can use golang's pprof tool for performance analysis, or you can use prometheus for monitoring data collection.
    - `-i --init` This will initialize an empty configuration file when the configuration file does not exist.
3. The anchorage service process is already running! But if you want anchorage to serve stably in the background, you need some necessary auxiliary tools.
   ```
   [Unit]
   Description=anchorage
   After=network.target
   
   [Service]
   User=root
   ExecStart=/root/anchorage/anchorage_linux_amd64 start --cmd.addr 0.0.0.0:2024
   ExecStop=/usr/bin/killall -9 anchorage
   Restart=on-failure
   PrivateTmp=true
   
   [Install]
   WantedBy=multi-user.target
   ```
   Here we take Linux systemd as an example to write a service file to let anchorage continue to run in the background!
### Configuration-service
1. Although the anchor service is already running, because the configuration file is an empty shell, we need to load some functions to make anchorage officially work for us! (The following will be performed through the command line tool that comes with anchorage)
   First, we need a traffic forwarding center-`server`!
   ```
   anchorage core server command
   
   Usage:
   anchorage server [command]
   
   Available Commands:
   add         add one anchorage core server. (requires vim, vi, nano, or emacs tool)
   config      print one anchorage core server config.
   del         del one anchorage core server. (if it is in a working state, the related work will be stopped)
   reload      reload one anchorage core server.
   start       start one anchorage core server.
   stop        stop one anchorage core server.
   update      update one anchorage core server config and reload (if in working). (requires vim, vi, nano, or emacs tool)
   
   Flags:
       --cmd.addr string   cmd address
       --cmd.i             cmd tls insecure
       --cmd.nk string     cmd network
       --cmd.p string      cmd password
       --cmd.tls           cmd tls enable
       --cmd.u string      cmd username
   -h, --help              help for server
   ```
   Through the `add` command, you can add a `server` module. Generally speaking, there will be text editors such as vim in the terminal, so we only need to fill in the template.
   But we don’t need to think too much, because we are starting quickly, so we only need to fill in the necessary items.
   ```
   enable: false # loaded then to work <bool>
   config: # server config <*config.ServerConfig>
      nodeInfo: # current server node config <config.NodeConfig>
         nodeName: "test1" # must be <string>
         baseNetwork: # base network list <[]config.BaseNetworkConfig>
         - network: "tcp" # must be tcp,udp(quic) <string>
           address: "127.0.0.1:20240" # work address <string>
   ```
   It will be automatically submitted after saving. At this time, we can view the existing `server` module through `./anchorage_linux_amd64 view server`, and then copy the id (a new id will be generated every time the anchorage service is started, so there is no need to remember it specially it)
   You can start this module through `./anchorage_linux_amd64 server start [id]`! If you want anchorage to also start the `server` module at startup, you only need to change `enable: false # loaded then to work <bool>` in the configuration to true.
2. Along the way, you may be bored without experiencing the functions of anchorage. Sorry, anchorage is not a thing that can be used out of the box. It requires a little configuration before it can serve.
   Let's add a `client` to connect to the previously configured `server`.
   ```
   anchorage core client command
   
   Usage:
   anchorage client [command]
   
   Available Commands:
   add         add one anchorage core client. (requires vim, vi, nano, or emacs tool)
   config      print one anchorage core client config.
   del         del one anchorage core client. (if it is in a working state, the related work will be stopped)
   reload      reload one anchorage core client.
   start       start one anchorage core client.
   stop        stop one anchorage core client.
   update      update one anchorage core client config and reload (if in working). (requires vim, vi, nano, or emacs tool)

   Flags:
       --cmd.addr string   cmd address
       --cmd.i             cmd tls insecure
       --cmd.nk string     cmd network
       --cmd.p string      cmd password
       --cmd.tls           cmd tls enable
       --cmd.u string      cmd username
   -h, --help              help for server
   ```
   Like `server`, add a `client` module through the `add` command and fill in the necessary items. The information filled here is consistent with the `server` configuration.
   ```
   enable: false # loaded then to work <bool>
   config: # client config <*config.ClientConfig>
      nodes: # dial to server node list <[]config.NodeConfig>
         -  nodeName: "test1" # must be <string>
            baseNetwork: # base network list <[]config.BaseNetworkConfig>
            -  network: "tcp" # must be tcp,udp(quic) <string>
               address: "127.0.0.1:20240" # work address <string>
   ```
   It will be automatically submitted after saving. At this time, we can view the existing `client` module through `./anchorage_linux_amd64 view client`, and then copy the id (a new id will be generated every time the anchorage service is started, so there is no need to remember it specially it)
   You can start this module through `./anchorage_linux_amd64 view start [id]`! If you want anchorage to also start the `client` module at startup, you only need to change `enable: false # loaded then to work <bool>` in the configuration to true.
3. After creating the `client` module, we simply create a `listen` and `dial` modules.
   `./anchorage_linux_amd64 listen add [client id]` Use the id of the `client` just used and fill in the necessary items.
   ```
   enable: false # loaded then to work <bool>
   name: "testssh" # service name <string>
   switchLink: true # whether or not to allow the link <bool>
   outNetwork: # out network config <*sdk.NetworkConfig>
      network: "tcp" # must be tcp or udp <string>
      address: "127.0.0.1:22" # must be tcp or udp <string>
   multi: true # whether support multi io to link <bool>
   ```
   Here we simply register a `listen` module pointing to port 22 ssh, and query the `listen` id through `./anchorage_linux_amd64 view client default [client id]`.
   Finally, start it through `./anchorage_linux_amd64 listen start [client id] [listen id]`. The enable in the configuration has the same effect as before.
   `./anchorage_linux_amd64 dial add [client id]` Use the id of the `client` just used and fill in the necessary items.
   ```
   enable: false # loaded then to work <bool>
   link: "testssh" # link service name <string>
   inNetwork: # in network config <*sdk.NetworkConfig>
      network: "tcp" # must be tcp or udp <string>
      address: "127.0.0.1:20241" # must be tcp or udp <string>
   outNetwork: # out network config <*sdk.NetworkConfig>
      network: "tcp" # must be tcp or udp <string>
      address: "127.0.0.1:22" # must be tcp or udp <string>
   ```
   Finally, start it through `./anchorage_linux_amd64 dial start [client id] [listen id]`. The enable in the configuration has the same effect as before.
4. Through the above steps, we have established a connection tunnel of `dial` -> `client` -> `server` -> `client` -> `listen`.
   Use `ssh user@127.0.0.1 -p 20241` to test whether the tunnel is working properly.
5. More usage methods can be found [here](./doc/README.md)

## Build-and-compile
### Install-golang
- For details, see [go.dev](https://go.dev)
### Clone-repository
```
git clone https://github.com/peakedshout/anchorage-core.git
cd anchorage-core
```
### Compile-using-golang
`go build -o anchorage ./cmd/anchorage`
or
`go install github.com/peakedshout/anchorage-core/cmd/anchorage`

## TODO
- [ ] More convenient cli calling
- [ ] The corresponding GUI client (uncomfortable, GUI design is not my strong point)
- [ ] More plugin support
- [ ] Fix bug...
- [ ] Optimize more designs
- [ ] Hope to get more feedback on using experience