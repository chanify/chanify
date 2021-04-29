# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Release](https://img.shields.io/github/v/release/chanify/chanify?logo=github&style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/cn/app/id1531546573)
[![WebStore](https://img.shields.io/chrome-web-store/v/llpdpmhkemkjeeigibdamadahmhoebdg?logo=Google%20Chrome&logoColor=white&style=flat-square)](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg)
[![Workflow](https://img.shields.io/github/workflow/status/chanify/chanify/ci?label=build&logo=github&style=flat-square)](https://github.com/chanify/chanify/actions?workflow=ci)
[![CodeQL](https://img.shields.io/github/workflow/status/chanify/chanify/codeql?label=codeql&logo=github&style=flat-square)](https://github.com/chanify/chanify/actions?workflow=codeql)
[![Codecov](https://img.shields.io/codecov/c/github/chanify/chanify?logo=codecov&style=flat-square)](https://codecov.io/gh/chanify/chanify)
[![Total alerts](https://img.shields.io/lgtm/alerts/g/chanify/chanify.svg?logo=lgtm&logoWidth=18&style=flat-square)](https://lgtm.com/projects/g/chanify/chanify/alerts/)
[![Go Report Card](https://goreportcard.com/badge/github.com/chanify/chanify?style=flat-square)](https://goreportcard.com/report/github.com/chanify/chanify)
[![GitHub](https://img.shields.io/github/license/chanify/chanify?style=flat-square)](LICENSE)
[![Docker pull](https://img.shields.io/docker/pulls/wizjin/chanify?style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Downloads](https://img.shields.io/github/downloads/chanify/chanify/total?style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![Users](https://img.shields.io/chrome-web-store/users/llpdpmhkemkjeeigibdamadahmhoebdg?style=flat-square)](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg)

[English](README.md) | 简体中文

Chanify是一个简单的消息推送工具。每一个人都可以利用提供的API来发送消息推送到自己的iOS设备上。

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
            <li><a href="#发送文件">发送文件</a></li>
        </ul>
    </li>
    <li><a href="#配置文件">配置文件</a></li>
    <li><a href="#安全">安全</a></li>
    <li><a href="#chrome插件">Chrome插件</a></li>
    <li><a href="#贡献">贡献</a></li>
    <li><a href="#许可证">许可证</a></li>
  </ol>
</details>

## 功能

Chanify包括这些功能：

- 支持自定义频道分类消息
- 支持部署自己的节点服务器
- 依照分布式架构设计系统
- 随机账号生成保护隐私
- 支持文本/图片/文件等多种消息格式

## 入门

1. 从AppStore安装[iOS应用](https://itunes.apple.com/cn/app/id1531546573)（1.0.0或以上版本）。
2. 获取发送使用的令牌`token`，[更多细节](https://github.com/chanify/chanify-ios)。
3. 使用API来发送消息。

## 安装

### 预编译包

可以[这里](https://github.com/chanify/chanify/releases/latest)下载最新的预编译二进制包。

### Docker

```bash
$ docker pull wizjin/chanify:latest
```

### 从源代码

```bash
$ go install github.com/chanify/chanify
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

# 文件消息
$ chanify send --endpoint=http://<address>:<port> --token=<token> --file=<文件路径> --text=<文件描述>
```

`endpoint` 默认值是 `https://api.chanify.net`，并且会使用默认服务器发送消息。
如果使用的是自建的节点服务器，请在讲`endpoint`设置成自建服务器的URL。

### 作为无状态服务器

Chanify可以作为无状态服务器运行，在这种模式下节点服务器不会保存设备信息（APNS令牌）。

所有的设备信息会被存储在 api.chanify.net。

消息会在节点服务器加密之后由 api.chanify.net 代理发送给苹果的APNS服务器。

消息的流动如下:

```
开始 => 自建节点服务器 => api.chanify.net => 苹果APNS服务器 => iOS客户端
```

```bash
# 命令行启动
$ chanify serve --host=<ip address> --port=<port> --secret=<secret key> --name=<node name> --endpoint=http://<address>:<port>

# 使用Docker启动
$ docker run -it wizjin/chanify:latest serve --secret=<secret key> --name=<node name> --endpoint=http://<address>:<port>
```

### 作为有状态服务器

Chanify可以作为有状态服务器运行，在这种模式下节点服务器会保存用户的设备信息（APNS令牌）。

消息会直接由节点服务器加密之后发送给苹果的APNS服务器。

消息的流动如下:

```
开始 => 自建节点服务器 => Apple server => iOS客户端
```

启动服务器

```bash
# 命令行启动
$ chanify serve --host=<ip address> --port=<port> --name=<node name> --datapath=~/.chanify --endpoint=http://<address>:<port>

# 使用Docker启动
$ docker run -it -v /my/data:/root/.chanify wizjin/chanify:latest serve --name=<node name> --endpoint=http://<address>:<port>
```

默认会使用sqlite保存数据，如果要使用MySQL作为数据库存储信息可以添加如下参数：

```bash
--dburl=mysql://<user>:<password>@tcp(<ip address>:<port>)/<database name>?charset=utf8mb4&parseTime=true&loc=Local
```

注意：Chanify不会创建数据库，只会创建表格，所以使用前请先自行建立数据库。

### 添加节点服务器

- 启动节点服务器
- 获取服务器二维码（`http://<address>:<port>/`）
- 打开iOS的客户端扫描二维码添加节点

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
```
http://<address>:<port>/v1/sender/<token>/<message>
```

- __POST__
```
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
}
```

支持以下参数：

| 参数名    | 默认值 | 描述                              |
| -------- | ----- | -------------------------------- |
| title    | 无    | 通知消息的标题                      |
| copy     | 无    | 可选的复制文本（仅文本消息有效）       |
| autocopy | `0`   | 是否自动复制文本（仅文本消息有效）     |
| sound    | `0`   | `1` 启用声音提示, 其他情况会静音推送  |
| priority | `10`  | `10` 正常优先级, `5` 较低优先级     ｜

例如：

```
http://<address>:<port>/v1/sender/<token>?sound=1&priority=10&title=hello&copy=123&autocopy=1
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
cat <jpeg文件路径> | curl -H "Content-Type: image/jpeg" --data-binary @- "http://<address>:<port>/v1/sender/<token>"
```

- Content-Type: `multipart/form-data`

```bash
$ curl --form "image=@<jpeg文件路径>" "http://<address>:<port>/v1/sender/<token>"
```

### 发送文件

目前仅支持使用 **POST** 方法通过自建的有状态服务器才能发文件。

- Content-Type: `multipart/form-data`

```bash
$ curl --form "file=@<文件路径>" "http://<address>:<port>/v1/sender/<token>"
```

## 配置文件

可以通过yml文件来配置Chanify，默认路径`~/.chanify.yml`。

```yml
server:
    host: 0.0.0.0   # 监听IP地址
    port: 8080      # 监听端口
    endpoint: http://my.server/path # 入口URL
    name: Node name # 节点名称
    secret: <secret code> # 无状态服务器使用的密钥
    datapath: <data path> # 有状态服务器使用的数据存储路径
    dburl: mysql://root:test@tcp(127.0.0.1:3306)/chanify?charset=utf8mb4&parseTime=true&loc=Local # 有状态服务器使用的数据库链接
    register:
        enable: false # 关闭注册
        whitelist: # 白名单
            - <user id 1>
            - <user id 2>

client: # 作为客户端发送消息时使用
    sound: 1    # 是否有提示音
    endpoint: <default node server endpoint>
    token: <default token>
```

## 安全

可以通过禁用节点服务器的用户注册功能，来使Node服务器成为私有服务器，防止非授权用户使用。

```bash
chanify serve --registerable=false --whitelist=<user1 id>,<user2 id>
```

- `--registerable=false`: 这个参数用来禁用用户注册
- `whitelist`: 服务器禁用用户注册后，仍然可以添加使用的用户

## Chrome插件

可以从[Chrome web store](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg)下载插件.

插件有以下功能:

- 发送选中的`文本/图片/链接`消息到Chanify
- 发送网页链接到Chanify

## 贡献

贡献使开源社区成为了一个令人赞叹的学习，启发和创造场所。 **十分感谢**您做出的任何贡献。

1. Fork 本项目
2. 切换到 dev 分支 (`git checkout dev`)
3. 创建您的 Feature 分支 (`git checkout -b feature/AmazingFeature`)
4. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
5. 推送到分支 (`git push origin feature/AmazingFeature`)
6. 开启一个 Pull Request (合并到 `chanify:dev` 分支)

## 许可证

根据MIT许可证分发，详情查看[`LICENSE`](LICENSE)。
