# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Release](https://img.shields.io/github/v/release/chanify/chanify?logo=github&style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/cn/app/id1531546573)
[![WebStore](https://img.shields.io/chrome-web-store/v/llpdpmhkemkjeeigibdamadahmhoebdg?logo=Google%20Chrome&logoColor=white&style=flat-square)](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg)
[![Windows](https://img.shields.io/github/v/release/chanify/chanify-win?label=windows&logo=windows&style=flat-square)](https://github.com/chanify/chanify-win/releases/latest)
[![Workflow](https://img.shields.io/github/workflow/status/chanify/chanify/ci?label=build&logo=github&style=flat-square)](https://github.com/chanify/chanify/actions?workflow=ci)
[![CodeQL](https://img.shields.io/github/workflow/status/chanify/chanify/codeql?label=codeql&logo=github&style=flat-square)](https://github.com/chanify/chanify/actions?workflow=codeql)
[![Codecov](https://img.shields.io/codecov/c/github/chanify/chanify?logo=codecov&style=flat-square)](https://codecov.io/gh/chanify/chanify)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/chanify/chanify.svg?logo=lgtm&logoWidth=18&style=flat-square)](https://lgtm.com/projects/g/chanify/chanify/alerts/)
[![Go Report Card](https://goreportcard.com/badge/github.com/chanify/chanify?style=flat-square)](https://goreportcard.com/report/github.com/chanify/chanify)
[![Go Reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/chanify/chanify)
[![GitHub](https://img.shields.io/github/license/chanify/chanify?style=flat-square)](LICENSE)
[![Docker pull](https://img.shields.io/docker/pulls/wizjin/chanify?style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Downloads](https://img.shields.io/github/downloads/chanify/chanify/total?style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![Users](https://img.shields.io/chrome-web-store/users/llpdpmhkemkjeeigibdamadahmhoebdg?style=flat-square)](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg)

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
