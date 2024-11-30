# Anchorage
#### [En](./README.md)/CN

## 简介
本项目是一个用于实现强大的流量中心的程序，旨在【高效、强势、多功能、安全】。

### 为什么会有这玩意？
- 因为我总是疲于调整网络，便思考是否能做出个能够让我便利的切换调控流量的玩意。
- 所以，`Anchorage` 诞生了。
- 也许你知道[go-CFC](https://github.com/peakedshout/go-CFC)，它是`Anchorage`的前身，总结了经验并拓展产生了`Anchorage`。
- `Anchorage`已经为我工作了很长一段时间了，它在我的日常工作中起到了不少的作用。
- 所以我想将它分享出来，并希望有人能使用它并提出一些建议。

### 它能为你做什么？
- 流量转发，仅仅而已。
- `Anchorage`并没有做太高深的事情，但它能够完成一些平日我们觉得繁琐的事情。
- 例如，将一个内网服务暴露出来，将本地的一些流量分流进行代理等等。
- 如果更近一步，可以接入sdk进行一个特殊化开发。 

## 目录
- [安装说明](#安装说明)
- [快速开始](#快速开始)
- [构建编译](#构建编译)
- [TODO](#TODO)

## 安装说明
### 系统要求
- 操作系统：Windows, macOS, or Linux (Unix)

### 安装步骤
1. 在最近版本中获取可执行二进制：
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
   执行文件输出版本相关信息即可说明可执行，可以开始anchorage之旅了！（●´3｀）~♪

## 快速开始
### 启动服务
1. 我们需要初始化一个配置文件用于服务启动
   ```
   Usage:
   anchorage init [config] [flags]
   ```
   - `config` 用于指定要启动的服务配置文件，如果为空则默认使用``~/.anchorage/cfg.yaml``。
2. anchorage的命令行工具想要启动十分简单
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
   - `config` 用于指定要启动的服务配置文件，如果为空则默认使用``~/.anchorage/cfg.yaml``进行启动服务。
   - `--cmd.addr` 这是后台管理的监听地址，默认使用`127.0.0.1:2024`进行监听。
   - `--cmd.nk` 这是后台管理的监听网络类型，默认使用`tcp`。
   - `--cmd.u` `--cmd.p` 这是后台管理的用户密码，默认为空。
   - `--cmd.tls` `--cmd.cert` `--cmd.key` 这是tls的一些配置，默认不启用。
   - `-d --debug` 这是启用golang的pprof和metrics的功能，可以用golang的pprof工具进行性能分析，也可以使用prometheus进行监控数据采集。
   - `-i --init` 这是当配置文件不存在时，将初始化一个空的配置文件。
3. anchorage的服务进程已经运行起来啦！但如果想让anchorage在后台稳定服务就需要一些必要的辅助工具。
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
   这里以Linux的systemd为例子，编写一个service文件，让anchorage在后台持续运行起来！
### 配置服务
1. 虽然anchorage服务已经运行起来了，但因为配置文件是一个空壳，所以我们需要装载一些功能让anchorage正式为我们工作！（以下将通过anchorage自带的命令行工具进行操作）
   首先，我们需要一个流量转发中心--`server`！
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
   通过`add`命令，可以增加一个`server`模块，一般来说，在终端基本会有vim等文本编辑器，所以我们只需要填写模板就行了。
   但我们不需要考虑太多，因为是快速开始，所以就只需要填写必要项。
   ```
   enable: false # loaded then to work <bool>
   config: # server config <*config.ServerConfig>
      nodeInfo: # current server node config <config.NodeConfig>
         nodeName: "test1" # must be <string>
         baseNetwork: # base network list <[]config.BaseNetworkConfig>
         - network: "tcp" # must be tcp,udp(quic) <string>
           address: "127.0.0.1:20240" # work address <string>
   ```
   保存后会自动提交，这时我们可以通过`./anchorage_linux_amd64 view server`进行查看现有的`server`模块，然后复制id（每次启动anchorage服务时会新生成一个id，所以不需要特意记住它）
   通过`./anchorage_linux_amd64 server start [id]`就能启动该模块啦！如果想anchorage在启动时也将该`server`模块启动，那么只需要将配置中的`enable: false # loaded then to work <bool>`改成true即可。
2. 一路做下来，你可能会厌烦还未体验anchorage的功能，抱歉，anchorage不是一个开箱即用的玩意，它需要一点配置才可以服务。
   让我们添加一个`client`对接之前配置的`server`。
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
   和`server`一样，通过`add`命令添加一个`client`模块，并填入必要项。这里填的和`server`配置一致。
   ```
   enable: false # loaded then to work <bool>
   config: # client config <*config.ClientConfig>
      nodes: # dial to server node list <[]config.NodeConfig>
         -  nodeName: "test1" # must be <string>
            baseNetwork: # base network list <[]config.BaseNetworkConfig>
            -  network: "tcp" # must be tcp,udp(quic) <string>
               address: "127.0.0.1:20240" # work address <string>
   ```
   保存后会自动提交，这时我们可以通过`./anchorage_linux_amd64 view client`进行查看现有的`client`模块，然后复制id（每次启动anchorage服务时会新生成一个id，所以不需要特意记住它）
   通过`./anchorage_linux_amd64 view start [id]`就能启动该模块啦！如果想anchorage在启动时也将该`client`模块启动，那么只需要将配置中的`enable: false # loaded then to work <bool>`改成true即可。
3. 在创建了`client`模块后，我们简单的创建一个`listen`和`dial`模块。
   `./anchorage_linux_amd64 listen add [client id]` 使用刚才使用的`client`的id，并填入必要项。
   ```
   enable: false # loaded then to work <bool>
   name: "testssh" # service name <string>
   switchLink: true # whether or not to allow the link <bool>
   outNetwork: # out network config <*sdk.NetworkConfig>
      network: "tcp" # must be tcp or udp <string>
      address: "127.0.0.1:22" # must be tcp or udp <string>
   multi: true # whether support multi io to link <bool>
   ```
   这里我们简单的注册一个指向22端口ssh的`listen`模块，并通过`./anchorage_linux_amd64 view client default [client id]`查询`listen`的id。
   最后通过`./anchorage_linux_amd64 listen start [client id] [listen id]`启动，配置中的enable和之前的效果一样。
   `./anchorage_linux_amd64 dial add [client id]` 使用刚才使用的`client`的id，并填入必要项。
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
   最后通过`./anchorage_linux_amd64 dial start [client id] [listen id]`启动，配置中的enable和之前的效果一样。
4. 通过以上步骤，我们已经建立起了一个`dial` -> `client` -> `server` -> `client` -> `listen` 的连接隧道。
   使用`ssh user@127.0.0.1 -p 20241`测试隧道是否正常工作。
5. 更多的使用方法可以见[这里](./doc/README_CN.md)

## 构建编译
### 安装golang
- 详细见[go.dev](https://go.dev)
### 克隆仓库
```
git clone https://github.com/peakedshout/anchorage-core.git
cd anchorage-core
```
### 使用golang编译
   `go build -o anchorage ./cmd/anchorage`
   or
   `go install github.com/peakedshout/anchorage-core/cmd/anchorage`

## TODO
- [ ] 更便利的cli调用
- [ ] 对应的gui客户端（难受，gui设计并不是我的强项）
- [ ] 更多的插件支持
- [ ] 修复bug ...
- [ ] 优化更多设计
- [ ] 希望得到更多的使用体验反馈