# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Release](https://img.shields.io/github/v/release/chanify/chanify?logo=github&style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/cn/app/id1531546573)
[![Windows](https://img.shields.io/github/v/release/chanify/chanify-win?label=windows&logo=windows&style=flat-square)](https://github.com/chanify/chanify-win/releases/latest)
[![Codecov](https://img.shields.io/codecov/c/github/chanify/chanify?logo=codecov&style=flat-square)](https://codecov.io/gh/chanify/chanify)
[![Go Report Card](https://goreportcard.com/badge/github.com/chanify/chanify?style=flat-square)](https://goreportcard.com/report/github.com/chanify/chanify)
[![Go Reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/chanify/chanify)
[![GitHub](https://img.shields.io/github/license/chanify/chanify?style=flat-square)](LICENSE)
[![Docker pull](https://img.shields.io/docker/pulls/wizjin/chanify?style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Downloads](https://img.shields.io/github/downloads/chanify/chanify/total?style=flat-square)](https://github.com/chanify/chanify/releases/latest)

[English](README.md) | 简体中文

Chanify 是一个简单的消息推送工具。每一个人都可以利用提供的 API 来发送消息推送到自己的 iOS 设备上。

<details open="open">
  <summary><h2 style="display: inline-block">目录</h2></summary>
  <ol>
    <li><a href="#功能">功能</a></li>
    <li><a href="#入门">入门</a></li>
    <li>
        <a href="#安装">安装</a>
        <ul>
            <li><a href="#预编译包">预编译包</a></li>
            <li><a href="#docker">Docker</a></li>
            <li><a href="#从源代码">从源代码</a></li>
        </ul>
    </li>
    <li>
        <a href="#用法">用法</a>
        <ul>
            <li><a href="#作为客户端">作为客户端</a></li>
            <li><a href="#作为无状态服务器">作为无状态服务器</a></li>
            <li><a href="#作为有状态服务器">作为有状态服务器</a></li>
            <li><a href="#添加节点服务器">添加节点服务器</a></li>
            <li>
                <a href="#发送消息">发送消息</a>
                <ul>
                    <li><a href="#命令行">命令行</a></li>
                    <li><a href="#python-3">Python 3</a></li>
                    <li><a href="#ruby">Ruby</a></li>
                    <li><a href="#nodejs">NodeJS</a></li>
                    <li><a href="#php">PHP</a></li>
                </ul>
            </li>
        </ul>
    </li>
    <li>
        <a href="#http-api">HTTP API</a>
        <ul>
            <li><a href="#发送文本">发送文本</a></li>
            <li><a href="#发送链接">发送链接</a></li>
            <li><a href="#发送图片">发送图片</a></li>
            <li><a href="#发送音频">发送音频</a></li>
            <li><a href="#发送文件">发送文件</a></li>
            <li><a href="#发送动作">发送动作</a></li>
        </ul>
    </li>
    <li><a href="#配置文件">配置文件</a></li>
    <li>
        <a href="#安全">安全</a>
        <ul>
            <li><a href="#设置可注册">设置可注册</a></li>
            <li><a href="#令牌生命周期">令牌生命周期</a></li>
        </ul>
    </li>
    <li><a href="#chrome-插件">Chrome 插件</a></li>
    <li><a href="#windows-客户端">Windows 客户端</a></li>
    <li><a href="#docker-compose">Docker Compose</a></li>
    <li><a href="#lua-api">Lua API</a></li>
    <li><a href="#贡献">贡献</a></li>
    <li><a href="#许可证">许可证</a></li>
  </ol>
</details>

## 功能

Chanify 包括这些功能：

- 支持自定义频道分类消息
- 支持部署自己的节点服务器
- 依照去中心化设计的系统架构
- 随机账号生成保护隐私
- 支持文本/图片/音频/文件等多种消息格式

## 入门

1. 从 AppStore 安装 [iOS 应用](https://itunes.apple.com/cn/app/id1531546573)（1.0.0 或以上版本）或 [macOS 应用](https://apps.apple.com/cn/app/id1531546573) (1.3.0 或以上版本)。
2. 获取发送使用的令牌`token`，[更多细节](https://github.com/chanify/chanify-ios)。
3. 使用 API 来发送消息。

## Coding agent 使用指南

把 [AGENT_GUIDE.md](AGENT_GUIDE.md) 直接交给 coding agent，即可按文档配置 CLI、本地 stdio MCP 或远程 MCP。文档包含可复制的任务指令、凭据区别、发送示例、结果与重试规则及投递排查；不要把真实凭据粘贴到对话。文档描述 1.5.0 源码接口，使用前先检查已安装版本的命令帮助。

## 安装

### 预编译包

可以[这里](https://github.com/chanify/chanify/releases/latest)下载最新的预编译二进制包。

### Docker

```bash
$ docker pull wizjin/chanify:latest
```

### 从源代码

```bash
$ git clone https://github.com/chanify/chanify.git
$ cd chanify
$ make install
```

## 用法

### 作为客户端

可以使用下列命令来发送消息

```bash
# 文本消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --text=<文本消息>

# 链接消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --link=<网页链接>

# 图片消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --image=<图片文件路径>

# 音频消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --audio=<音频文件路径>

# 文件消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --file=<文件路径> --text=<文件描述>

# 动作消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --text=<文本消息> --title=<文本标题> --action="<动作名字>|<动作链接 url>"

# 时间线消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --timeline.code=<代号> <键值 1>=<数值 1> <键值 2>=<数值 2> ...
```

`endpoint` 默认值是 `https://api.chanify.net`，并且会使用默认服务器发送消息。
如果使用的是自建的节点服务器，请在讲`endpoint`设置成自建服务器的 URL。

### 作为无状态服务器

Chanify 可以作为无状态服务器运行，在这种模式下节点服务器不会保存设备信息（APNS 令牌）。

所有的设备信息会被存储在 api.chanify.net。

消息会在节点服务器加密之后由 api.chanify.net 代理发送给苹果的 APNS 服务器。

消息的流动如下:

```text
开始 => 自建节点服务器 => api.chanify.net => 苹果APNS服务器 => iOS客户端
```

```bash
# 命令行启动
$ chanify serve --host=<ip address> --port=<port> --secret=<secret key> --name=<node name> --endpoint=http://<address>:<port>

# 使用Docker启动
$ docker run -it wizjin/chanify:latest serve --secret=<secret key> --name=<node name> --endpoint=http://<address>:<port>
```

### 作为有状态服务器

Chanify 可以作为有状态服务器运行，在这种模式下节点服务器会保存用户的设备信息（APNS 令牌）。

消息会直接由节点服务器加密之后发送给苹果的 APNS 服务器。

消息的流动如下:

```text
开始 => 自建节点服务器 => Apple server => iOS客户端
```

启动服务器

```bash
# 命令行启动
$ chanify serve --host=<ip address> --port=<port> --name=<node name> --datapath=~/.chanify --endpoint=http://<address>:<port>

# 使用Docker启动
$ docker run -it -v /my/data:/root/.chanify wizjin/chanify:latest serve --name=<node name> --endpoint=http://<address>:<port>
```

默认会使用 sqlite 保存数据，如果要使用 MySQL 作为数据库存储信息可以添加如下参数：

```bash
--dburl=mysql://<user>:<password>@tcp(<ip address>:<port>)/<database name>?charset=utf8mb4&parseTime=true&loc=Local
```

注意：Chanify 不会创建数据库，只会创建表格，所以使用前请先自行建立数据库。

### 添加节点服务器

- 启动节点服务器
- 获取服务器二维码（`http://<address>:<port>/`）
- 打开 iOS 的客户端扫描二维码添加节点

### 发送消息

#### 命令行

```bash
# 发送文本消息
$ curl --form-string "text=hello" "http://<address>:<port>/v1/sender/<token>"

# 发送文本文件
$ cat message.txt | curl -H "Content-Type: text/plain" --data-binary @- "http://<address>:<port>/v1/sender/<token>"
```

#### Python 3

```python
from urllib import request, parse

data = parse.urlencode({ 'text': 'hello' }).encode()
req = request.Request("http://<address>:<port>/v1/sender/<token>", data=data)
request.urlopen(req)
```

#### Ruby

```ruby
require 'net/http'

uri = URI('http://<address>:<port>/v1/sender/<token>')
res = Net::HTTP.post_form(uri, 'text' => 'hello')
puts res.body
```

#### NodeJS

```javascript
const https = require('https')
const querystring = require('querystring');

const data = querystring.stringify({ text: 'hello' })
const options = {
    hostname: '<address>:<port>',
    port: 80,
    path: '/v1/sender/<token>',
    method: 'POST',
    headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
        'Content-Length': data.length
        }
    }
    var req = https.request(options, (res) => {
    res.on('data', (d) => {
        process.stdout.write(d);
    });
});  
req.write(data);
req.end();
```

#### PHP

```php
$curl = curl_init();

curl_setopt_array($curl, [
    CURLOPT_URL           => 'http://<address>:<port>/v1/sender/<token>',
    CURLOPT_CUSTOMREQUEST => 'POST',
    CURLOPT_POSTFIELDS    => [ 'text' => 'hello' ],
]);

$response = curl_exec($curl);

curl_close($curl);
echo $response;
```

## HTTP API

### 发送文本

- __GET__
```url
http://<address>:<port>/v1/sender/<token>/<message>
```

- __POST__
```url
http://<address>:<port>/v1/sender/<token>
```

Content-Type:

- `text/plain`: Body is text message
- `multipart/form-data`: The block of data("text") is text message
- `application/x-www-form-urlencoded`: `text=<url encoded text message>`
- `application/json; charset=utf-8`: 字段都是可选的
```json
{
    "token": "<令牌Token>",
    "title": "<消息标题>",
    "text": "<文本消息内容>",
    "copy": "<可选的复制文本>",
    "autocopy": 1,
    "sound": 1,
    "priority": 10,
    "actions": [
        "动作名称1|http://<action host>/<action1>",
        "动作名称2|http://<action host>/<action2>",
        ...
    ],
    "timeline": {
        "code": "<标识 code>",
        "timestamp": 1620000000000,
        "items": {
            "key1": "value1",
            "key2": "value2",
            ...
        }
    }
}
```

支持以下参数：

| 参数名              | 默认值    | 描述                              |
| ------------------ | -------- | -------------------------------- |
| title              | 无       | 通知消息的标题                      |
| copy               | 无       | 可选的复制文本（仅文本消息有效）       |
| autocopy           | `0`      | 是否自动复制文本（仅文本消息有效）     |
| sound              | `0`      | `1` 启用声音提示, 其他情况会静音推送  |
| priority           | `10`     | `10` 正常优先级, `5` 较低优先级     |
| interruption-level | `active` | 通知时间的中断级别                  |
| actions            | 无       | 动作列表                           |
| timeline           | 无       | Timeline 对象                     |

`sound`:
  - 1 使用默认铃声
  - 使用铃声代码，例如："bell"

`interruption-level`:
  - `active`: 点亮屏幕并可能播放声音。
  - `passive`: 不点亮屏幕或播放声音。
  - `time-sensitive`: 点亮屏幕并可能播放声音； 可能会在“请勿打扰”期间展示。

`timestamp` 单位为毫秒 (时区 - UTC)

例如：

```url
http://<address>:<port>/v1/sender/<token>?sound=1&priority=10&title=hello&copy=123&autocopy=1
```

覆盖 `Content-Type`

```url
http://<address>:<port>/v1/sender/<token>?content-type=<text|json>
```

### 发送链接

```bash
$ curl --form "link=@<web url>" "http://<address>:<port>/v1/sender/<token>"
```

```json
{
    "link": "<web url>",
    "sound": 1,
    "priority": 10,
}
```

### 发送图片

目前仅支持使用 **POST** 方法通过自建的有状态服务器才能发送图片。

- Content-Type: `image/png` 或者 `image/jpeg`

```bash
cat <jpeg 文件路径> | curl -H "Content-Type: image/jpeg" --data-binary @- "http://<address>:<port>/v1/sender/<token>"
```

- Content-Type: `multipart/form-data`

```bash
$ curl --form "image=@<jpeg 文件路径>" "http://<address>:<port>/v1/sender/<token>"
```

### 发送音频

目前仅支持使用 **POST** 方法通过自建的有状态服务器才能发送 mp3 音频。

- Content-Type: `audio/mpeg`

```bash
cat <mp3 音频文件路径> | curl -H "Content-Type: audio/mpeg" --data-binary @- "http://<address>:<port>/v1/sender/<token>"
```

- Content-Type: `multipart/form-data`

```bash
$ curl --form "audio=@<mp3 音频文件路径>" "http://<address>:<port>/v1/sender/<token>"
```

### 发送文件

目前仅支持使用 **POST** 方法通过自建的有状态服务器才能发文件。

- Content-Type: `multipart/form-data`

```bash
$ curl --form "file=@<文件路径>" "http://<address>:<port>/v1/sender/<token>"
```

### 发送动作

发送动作消息（最多 4 个动作）。

- Content-Type: `multipart/form-data`

```bash
$ curl --form "action=动作名称1|http://<action host>/<action1>" "http://<address>:<port>/v1/sender/<token>"
```

脚本动作的 URL:

```bash
chanify://action/run-script/<脚本名称>?<参数名称1>=<参数值1>&<参数名称2>=<参数值2>
```

## 配置文件

可以通过 yml 文件来配置 Chanify，默认路径`~/.chanify.yml`。

```yml
server:
    host: 0.0.0.0   # 监听IP地址
    port: 8080      # 监听端口
    endpoint: http://my.server/path # 入口URL
    name: Node name # 节点名称
    secret: <secret code> # 无状态服务器使用的密钥
    datapath: <data path> # 有状态服务器使用的数据存储路径
#   pluginpath: <plugin path> # Lua 插件路径
    dburl: mysql://root:test@tcp(127.0.0.1:3306)/chanify?charset=utf8mb4&parseTime=true&loc=Local # 有状态服务器使用的数据库链接
    http:
        - readtimeout: 10s  # Http 读取超时时间设置为 10 秒
        - writetimeout: 10s # Http 写回超时时间设置为 10 秒
    register:
        enable: false # 关闭注册
        whitelist: # 白名单
            - <user id 1>
            - <user id 2>
#   plugin:
#       webhook:
#           - name: github  # POST http://my.server/path/v1/webhook/github/<token>
#             file: webhook/github.lua # <pluginpath>/webhook/github.lua
#             env:
#               secret_token: "secret token"

client: # 作为客户端发送消息时使用
    sound: 1    # 是否有提示音
    endpoint: <default node server endpoint>
    token: <default token>
    interruption-level: <interruption level>
```

## 安全

### 设置可注册

可以通过禁用节点服务器的用户注册功能，来使 Node 服务器成为私有服务器，防止非授权用户使用。

```bash
chanify serve --registerable=false --whitelist=<user1 id>,<user2 id>
```

- `--registerable=false`: 这个参数用来禁用用户注册
- `whitelist`: 服务器禁用用户注册后，仍然可以添加使用的用户

### 令牌生命周期

- 令牌默认寿命约 90 天左右
- 可以在频道详情页面配置令牌寿命 (1 天 ~ 5 年)

如果令牌被泄露，请将泄露的令牌添加到禁用列表（在 iOS 客户端设置页面）。

*注意: 请保护您的令牌不被泄露，禁用列表需要受信任的节点服务器（1.1.9 或以上版本）。*

## Chrome 插件

可以从[Chrome web store](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg)下载插件.

插件有以下功能:

- 发送选中的`文本/图片/音频/链接`消息到 Chanify
- 发送网页链接到 Chanify

## Windows 客户端

可以从[这里](https://github.com/chanify/chanify-win/releases/latest)获取到 [Windows 客户端](https://github.com/chanify/chanify-win)。

客户端有以下功能:

- 支持在 Windows 的 `发送到` 菜单中添加 Chanify
- 支持利用`命令行`发送消息

## Docker Compose

1. 安装 [docker compose](https://docs.docker.com/compose/install).
2. 编辑配置文件 (`docker-compose.yml`).
3. 启动 docker compose: `docker-compose up -d`

`docker-compose.yml`:
```yml
version: "3"
services:
    web:
        image: nginx:alpine
        restart: always
        volumes:
            - <workdir>/nginx.conf:/etc/nginx/nginx.conf
            - <workdir>/ssl:/ssl
        ports:
            - 80:80
            - 443:443
    chanify:
        image: wizjin/chanify:dev
        restart: always
        volumes:
            - <workdir>/data:/data
            - <workdir>/chanify.yml:/root/.chanify.yml
```

| 参数名    | 描述                        |
| -------- | --------------------------- |
| workdir  | 节点服务器的数据和配置存储目录。 |

`<workdir>/nginx.conf`:
```txt
user  nginx;
worker_processes  auto;

error_log  /var/log/nginx/error.log warn;
pid        /var/run/nginx.pid;

events {
	worker_connections  1024;
}

http {
	include       /etc/nginx/mime.types;
	default_type  application/octet-stream;
	log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
    '$status $body_bytes_sent "$http_referer" '
    '"$http_user_agent" "$http_x_forwarded_for"';

	access_log  /var/log/nginx/access.log  main;

	server_tokens   off;
	autoindex       off;
	sendfile        on;
	tcp_nopush      on;
	tcp_nodelay     on;

	keepalive_timeout  10;

    server {
		listen				80;
		server_name         <hostname or ip>;
		access_log          off;
		return 301 https://$host$request_uri;
	}

	server {
		listen              443 ssl http2;
		server_name         <hostname or ip>;
		ssl_certificate     /ssl/<ssl key>.crt;
		ssl_certificate_key /ssl/<ssl key>.key;
		ssl_protocols       TLSv1.2 TLSv1.3;
		ssl_ciphers         HIGH:!aNULL:!MD5;
		keepalive_timeout   30;
		charset             UTF-8;
		access_log          off;

		location / {
			proxy_set_header   Host               $host;
			proxy_set_header   X-Real-IP          $remote_addr;
			proxy_set_header   X-Forwarded-Proto  $scheme;
			proxy_set_header   X-Forwarded-For    $proxy_add_x_forwarded_for;
			proxy_pass http://chanify:8080/;
		}
	}
}
```

| 参数名           | 描述                           |
| --------------- | ----------------------------- |
| hostname or ip  | 节点服务器的外网域名或者 ip 地址。 |
| ssl key         | SSL 证书的文件名。               |

`<workdir>/chanify.yml`:
```yml
server:
    endpoint: https://<hostname or ip>
    host: 0.0.0.0
    port: 80
    name: <node name>
    datapath: /data
    register:
        enable: false
        whitelist: # whitelist    
            - <user id>
```

| 参数名           | 描述                           |
| --------------- | ----------------------------- |
| hostname or ip  | 节点服务器的外网域名或者 ip 地址。 |
| node name       | 节点服务器名称。                 |
| user id         | 用户白名单。                    |

## Lua API

使用方法

```lua
local hex = require "hex"
local bytes = hex.decode(hex_string)
local text = hex.encode(bytes_data)

local json = require "json"
local obj = json.decode(json_string)
local str = json.encode(json_object)

local crypto = require "crypto"
local is_equal = crypto.equal(mac1, mac2)
local mac = crypto.hmac("sha1", key, message) -- 支持 md5 sha1 sha256

-- Http 请求
local req = ctx:request()
local token_string = req:token()
local http_url = req:url()
local body_string = req:body()
local query_value = req:query("key")
local header_value = req:header("key")

-- 发送消息
ctx:send({
    title = "message title", -- 可选
    text = "message body",
    sound = "sound or not",  -- 可选
    copy = "copy",           -- 可选
    autocopy = "autocopy",   -- 可选 
})
```

例子: [Github webhook event](plugin/webhook/github.lua)

## 贡献

贡献使开源社区成为了一个令人赞叹的学习，启发和创造场所。 **十分感谢**您做出的任何贡献。

1. Fork 本项目
2. 切换到 dev 分支 (`git checkout dev`)
3. 创建您的 Feature 分支 (`git checkout -b feature/AmazingFeature`)
4. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
5. 推送到分支 (`git push origin feature/AmazingFeature`)
6. 开启一个 Pull Request (合并到 `chanify:dev` 分支)

## 许可证

根据 MIT 许可证分发，详情查看[`LICENSE`](LICENSE)。

### 服务器级发送令牌

支持设备存储的自建节点通过 `auth.server-token` feature 声明此能力。iOS 在节点详情中生成以 `cst1:` 开头的服务器发送令牌；它绑定用户、节点与有效期，不是 MCP 令牌。

省略 `channel` 时发送到未分类频道 `sys.none`；显式空值或无效值返回错误。当前值为不带填充的 base64url 编码 protobuf `Channel`，与 iOS 的频道 ID 格式一致。例如 `CAE` 明确表示未分类频道，`CAIaA2Rldg` 表示 code 为 `dev` 的用户频道。当前不维护服务端频道目录，不按显示名称查询频道；设备专属频道继续使用原频道令牌。

```sh
curl -X POST 'https://your-node.example/v1/sender' \
  -H 'Token: <server-wide-token>' \
  -H 'Content-Type: application/json' \
  --data '{"channel":"CAIaA2Rldg","title":"Build","text":"Tests passed"}'

chanify send --endpoint=https://your-node.example \
  --token='<server-wide-token>' --channel=CAIaA2Rldg --text='Tests passed'
```

`channel` 支持 JSON、URL 查询参数、表单和 multipart 字段；原始图片或音频请求通过查询参数指定。Webhook 的 Lua `send` 选项也支持 `channel`。旧频道令牌继续使用令牌内频道，外部参数不会覆盖它。

显式空值或无效频道返回 HTTP 400；无效签名、过期、节点不匹配或转发用户返回 HTTP 401。合法请求仍可能因没有绑定设备等原因失败，成功不代表已读。服务器级令牌不能用于附件下载鉴权。

服务器令牌默认不存在。在 iOS 的节点详情中打开“发送令牌”，点击右上角添加，填写名称和有效期后手工创建。令牌登记成功后才显示完整值；支持多枚并存，生成后只能改名。删除操作会立即吊销该令牌，其他令牌不受影响。iOS 列表支持下拉刷新。

`POST /rest/v1/server-tokens` 提供 `list`、`create`、`rename`、`revoke`，使用现有账户签名 `CHUserSign` 认证，不能用发送令牌管理凭据。请求包含 `user`、毫秒 `nonce` 和 `action`；创建附带 `name`/`token`，改名附带 `id`/`name`，吊销附带 `id`。nonce 允许正负 5 分钟时差。服务端持久化元数据、SHA-256 摘要和使用账户公钥加密的令牌副本；名称不参与令牌签名。SQLite/MySQL 启动时自动创建登记表。

服务器仅接受已登记且未吊销的全局令牌，未登记的早期试验令牌需通过新流程重新创建；旧频道令牌不受影响。登记和吊销随重启保留，删除账户时清理其记录。Web 管理与 MCP 接入见下文。

### MCP 接入（1.5.0）

仅持久化自建节点支持 MCP，默认不开启。配置在启动时读取，修改后重启：

```yaml
server:
  # 保留原有 host、port、endpoint、datapath 或 dburl 等配置
  mcp:
    enabled: true
    listen: false # 按需启用 SSE 监听，普通工具调用始终使用 JSON 响应
```

MCP 地址为节点的 `/mcp`。使用独立的 MCP Token，通过 `Authorization: Bearer <mcp-token>` 认证；频道发送 Token、服务器级发送 Token 和管理密码不能替代它。MCP 不依赖 `sender_token`。静态 Bearer 接入要求 agent 支持配置请求头，目前不提供 OAuth 登录授权。

使用本机数据库管理命令创建凭据，不需要开启 Web 管理：

```sh
chanify --config /path/to/chanify.yml mcp-token create \
  --user '<user-id>' --name coding-agent \
  --expires-in 2160h
chanify --config /path/to/chanify.yml mcp-token list --user '<user-id>'
chanify --config /path/to/chanify.yml mcp-token rename --user '<user-id>' --id '<credential-id>' --name new-name
chanify --config /path/to/chanify.yml mcp-token revoke --user '<user-id>' --id '<credential-id>'
```

命令从 `server.dburl` 或 `server.datapath` 读取节点数据库，也可通过 `--dburl` 显式指定。必须使用目标节点的数据库，并指定已经绑定且启用设备存储的用户。创建 MCP Token 不接受 --channel，不限制频道；发送时省略频道默认 sys.none。完整凭据只在创建时返回，数据库仅保存 SHA-256 摘要，列表不返回完整值；名称可修改，有效期不可修改。吊销后新请求立即拒绝，已有监听在下一次凭据检查时关闭（约一秒内）。账户删除会清理其 MCP 凭据。

agent 的连接参数示意（实际配置格式以客户端为准）：

```json
{
  "url": "https://your-node.example/mcp",
  "headers": {"Authorization": "Bearer <mcp-token>"}
}
```

工具按节点实际能力发现：

| 工具 | 参数与行为 |
| --- | --- |
| `send_text` | 必填 `text`；可选 `channel`、`title`、`summary`。有文件能力时最多 100,000 个 Unicode 字符，否则最多 400 个；标题最多 80 个，摘要最多 200 个。短文本直接发送，长文本完整保存为 UTF-8 附件，绝不静默截断正文 |
| `send_image` | 仅支持图片存储时提供。必填 `data_base64`，可选 `channel`；支持 PNG/JPEG，解码后最多 4 MiB、32 百万像素。可选标题和说明摘要 |
| `send_file` | 仅支持文件存储时提供。必填 `filename`、`data_base64`，可选 `channel`；文件名不能包含目录，最多 128 个字符，解码后最多 4 MiB；可选标题和摘要 |

频道代码对应 iOS 的频道 code：用户频道可使用字母/数字开头、包含字母数字和 `_ . -` 的 1–64 字符代码。`sys.none` 明确选择未分类频道，`sys.timesets` 选择时间线；设备专属频道不可用。省略频道默认 sys.none；显式空值、null 或无效代码报错。不限制频道，也不检查其是否在 iOS 创建。list_channels 已移除，未同步 iOS 完整目录。MCP 内部负责转换 protobuf 频道 ID，agent 不需要自行编码。

附件必须传实际 Base64 字节，不读取 agent 本地路径或抓取远程 URL。请求体最多 6 MiB。成功返回 `status: submitted`、实际 `format` 和 `request_id`，表示提交结果，不表示已读。网络超时后重试可能重复发送，目前不承诺幂等。

使用官方 Go MCP SDK 1.8.0 的无会话 Streamable HTTP，支持普通 JSON 调用；旧版 GET SSE 和 HTTP+SSE 不提供。开启 `listen` 后，可按 MCP 2026-07-28 使用 `subscriptions/listen` 监听；配置本身不热更新，因此不会凭空产生配置变更事件。SSE 每 15 秒保活，部署代理需允许流式响应并关闭缓冲。MCP 路径校验 Origin，远程部署使用 HTTPS。也可通过 Web 管理页面管理 MCP 令牌；另有账户签名认证的 `POST /rest/v1/mcp-tokens` 提供 `create/list/rename/revoke`，不会向发送工具开放凭据管理能力。


### Web 管理（1.5.0）

持久化自建节点可启用 `/admin`，与 MCP 独立开关。直接在原有配置中设置管理密码（12–72 个 UTF-8 字节），然后重启 Server：

```yaml
server:
  web_admin:
    password: '<管理密码>'
```

浏览器使用 Web Crypto 对密码的 UTF-8 字节计算 SHA-256，仅提交十六进制摘要；服务端启动时对配置密码做相同摘要，再生成 bcrypt 内存校验值。不记录原始密码或摘要日志。管理页需通过 HTTPS 或 localhost 访问；摘要本身也是登录凭据，不能替代 HTTPS。配置文件包含明文密码，请限制其读取权限。

访问节点的 `/admin` 并输入原始管理密码。未配置或为空时，页面与管理 API 均不注册；不提供默认密码。密码不符合要求会导致启动失败。配置仅在启动时读取，管理会话有效期为 8 小时，退出登录或重启后失效。远程访问使用 HTTPS；反向代理时 `server.endpoint` 应使用对外 HTTPS 地址，便于设置安全 Cookie。管理端使用独立会话、登录限流、同源检查和 CSRF 校验，发送令牌与 MCP 令牌不能用于登录。

页面支持：

- 查看发送令牌及 MCP 令牌的名称、ID、所属用户、创建/过期时间和吊销状态，按用户筛选。
- 修改令牌名称，二次确认后吊销；不修改既有令牌内容或有效期。发送令牌仍由 iOS 手动生成。
- 为已绑定用户创建 MCP 令牌，完整值仅当次展示；列表不返回完整凭据、哈希或加密副本。
- 查看已登记发送令牌（含 Webhook）及 MCP 令牌的已认证请求活动。普通请求显示最近使用，SSE 单独显示当前监听数量；IP 为直接连接地址，User-Agent 是客户端自报信息。

活动最多保留 2048 条，仅存在于当前 Server 进程内，重启即清空；令牌登记和吊销状态持久化。旧频道令牌没有服务器登记记录，不在管理列表和活动记录中。管理页面可手动刷新，吊销不影响其他令牌，也不会撤回已提交的通知。

启用文件存储的节点还会显示“媒体管理”，用于查看、下载和清理 Server 上的附件，不同步或修改 iOS 通知记录。支持按文件名或 ID 搜索、类型筛选、大小/时间/名称排序，以及每页 50 项的分页。新上传文件记录名称、检测到的类型和上传时间；相同目录下相同内容只占用一份文件，并保留最近最多 1024 个不同名称用于搜索。旧文件简单展示 ID、大小和文件修改时间，仍可下载、删除。

支持单个删除、当前页批量删除和清空全部文件，确认时显示数量与文件容量，完成后报告已删除数量、文件容量及失败 ID。清空需输入 `CLEAR`，且确认期间文件清单变化时必须刷新后重新确认。删除会影响所有引用该文件的通知，已下载到客户端的副本不受影响。容量按附件文件字节统计，不包含文件系统开销与管理元数据；正在下载的文件可能要等下载结束才释放磁盘空间。未启用文件存储时入口隐藏，媒体接口不可用。

图片、纯文本以及浏览器支持的音视频可预览；文本最多预览 256 KiB，完整内容可下载，HTML/SVG 等主动内容仅提供下载。元数据保存在文件存储目录下的 `.metadata/` 中，备份附件时应一并备份。列表按需扫描磁盘，大规模文件目录的刷新和清空耗时与文件数量有关。

### 本地 stdio MCP（`chanify mcp`）

coding agent 可以启动本机 `chanify mcp`，通过 stdio 调用工具，再由 CLI 使用现有 HTTP 发送接口提交给目标服务器。服务器无需开启 `/mcp`，也无需开放本机监听端口。发送结果表示提交成功，不代表 APNs 已送达或消息已读。

在配置文件中设置发送 Token：

```yaml
client:
  endpoint: https://api.chanify.net # 可省略；自建节点填写其地址
  token: '<发送 Token>'
```

使用常见的 stdio MCP 配置形式（具体字段以 agent 为准；推荐使用绝对路径）：

```json
{
  "mcpServers": {
    "chanify": {
      "command": "/path/to/chanify",
      "args": ["--config", "/path/to/chanify.yml", "mcp"]
    }
  }
}
```

也可用 `chanify mcp --endpoint https://your-node.example` 覆盖节点地址。默认服务器仅提供 `send_text`，最多 400 个 Unicode 字符，不查询节点信息，不提供长文本、图片或文件。使用从 iOS 获取的普通频道 Token，目的地由 Token 固定。

自建节点读取 `/rest/v1/info` 的能力信息；支持文件存储时，`send_text` 上限为 100,000 字符，长文本完整上传为 `message.txt`，可附标题和摘要；支持图片/文件时提供 `send_image` / `send_file`。工具参数、Base64 附件限制及提交结果与远程 MCP 共用定义。旧节点的信息接口返回 404 时，普通频道 Token 保守降级为短文本；其他发现错误停止启动。附件发送前再次检查存储能力；不可用时返回错误，避免仅发送预览。旧版 Server 的附件标题/说明呈现以该节点实现为准。

- 普通频道 Token：工具描述显示固定频道，不提供 `channel` 参数，运行时拒绝频道覆盖。
- 全局发送 Token：需要目标节点声明 `auth.server-token` 且节点 ID 匹配；`channel` 使用可读 code，省略时为 `sys.none`，CLI 内部完成编码。
- 本地子命令配置的是发送 Token，不接受远程 MCP Token 或管理密码；Token 不进入工具参数、URL 或日志。凭据是否登记、签名是否有效及是否被吊销，由目标服务器最终验证。

stdout 仅输出 MCP 协议，配置提示和错误写 stderr。stdin 专用于协议；不使用 `chanify send` 的 stdin 文件输入方式。agent 关闭输入或终止进程时退出，网络请求支持取消，发送请求超时为 30 秒，不自动重试。网络失败或响应缺少提交 ID 时不能保证未发送，直接重试可能重复通知。

### Agent 一次性发送（`chanify agent send`）

coding agent 可执行一次命令完成发送，无需配置 MCP。使用已有 `client.endpoint` / `client.token` 配置，目标服务器无需开启 `/mcp`。既有 `chanify send` 继续保留。

```sh
chanify --config /path/to/chanify.yml agent send --text '任务已完成'
chanify --config /path/to/chanify.yml agent send --text-file /path/to/report.md --title '验证报告'
chanify --config /path/to/chanify.yml agent send --stdin < /path/to/report.md
chanify --config /path/to/chanify.yml agent send --image /path/to/screenshot.png --summary '测试结果'
chanify --config /path/to/chanify.yml agent send --file /path/to/report.pdf --summary '完整报告'
```

`--text`、`--text-file`、`--stdin`、`--image`、`--file` 必须且只能选一个。文本按字面处理，`@` 和 `-` 不自动表示文件或 stdin；附件说明使用 `--summary`。输入文件须为可读普通文件，读取或格式校验失败会终止，不降级为只发送说明。所有文本为 UTF-8；文本文件/stdin 不裁剪换行。

默认服务器仅支持 400 字符短文本。自建节点按能力支持最多 100,000 字符文本，较长正文自动完整转为 UTF-8 附件；标题最多 80 字符，说明最多 200 字符。图片仅 PNG/JPEG，最多 32 百万像素；图片及文件最多 4 MiB。图片须可完整解码。短文本直接发送正文，summary 不替换正文。

全局发送 Token 可使用 `--channel development`，省略时发送到 `sys.none`；普通频道 Token 固定目的地，不允许传 `--channel`。此处 channel 是可读 code，与旧 send 的编码频道参数不同。使用 `--endpoint` 可覆盖配置地址；`--timeout 45s` 设置一次调用的总预算（默认 30s，发现查询仍有 10 秒阶段上限）。显式配置文件无效会停止，不使用部分配置继续发送。

正常执行的 stdout 只有一个 JSON 对象；日志走 stderr，`--help` 为普通帮助文本。成功示例：

```json
{"schema_version":1,"status":"submitted","format":"text","request_id":"<request-id>"}
```

`format` 为 text / text_attachment / image / file。submitted 只表示提交确认，不表示 APNs 到达或已读；failed 表示发送前失败或明确拒绝；unknown 表示发送可能已发生，不能据此自动重试。错误以 `error.code` 和 `error.message` 返回，不包含凭据或原始服务端响应正文。

退出码：0 提交确认；1 内部故障；2 输入/配置；3 鉴权；4 能力/发现/发送前超时；5 结果不确定；130/143 捕获中断信号。发送后的超时、5xx、缺少请求 ID 或响应损坏保守返回 unknown，不自动重发；未能可靠判断是否属于前置拒绝的其他发送错误也按 unknown 处理。
