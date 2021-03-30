# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Release](https://img.shields.io/github/v/release/chanify/chanify?logo=github&style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/cn/app/id1531546573)
[![Workflow](https://img.shields.io/github/workflow/status/chanify/chanify/ci?label=build&logo=github&style=flat-square)](https://github.com/chanify/chanify/actions?workflow=ci)
[![Codecov](https://img.shields.io/codecov/c/github/chanify/chanify?logo=codecov&style=flat-square)](https://codecov.io/gh/chanify/chanify)
[![GitHub](https://img.shields.io/github/license/chanify/chanify?style=flat-square)](LICENSE)
[![Docker pull](https://img.shields.io/docker/pulls/wizjin/chanify?style=flat-square)](https://hub.docker.com/r/wizjin/chanify)

[English](README.md) | 简体中文

Chanify是一个简单的消息推送工具。每一个人都可以利用提供的API来发送消息推送到自己的iOS设备上。

<details open="open">
  <summary><h2 style="display: inline-block">目录</h2></summary>
  <ol>
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
                </ul>
            </li>
        </ul>
    </li>
    <li>
        <a href="#http-api">HTTP API</a>
        <ul>
            <li><a href="#发送文本">发送文本</a></li>
            <li><a href="#发送图片">发送图片</a></li>
        </ul>
    </li>
    <li><a href="#贡献">贡献</a></li>
    <li><a href="#许可证">许可证</a></li>
  </ol>
</details>

## 入门

1. 从AppStore安装[iOS应用](https://itunes.apple.com/cn/app/id1531546573)（1.0.0或以上版本）。
2. 获取发送使用的令牌```token```，[更多细节](https://github.com/chanify/chanify-ios)。
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
$ chanify send --token=<token> --text=<message>

# 图片消息
$ chanify send --token=<token> --image=<image file path>
```

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
- 获取服务器二维码（```http://<address>:<port>/```）
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

- ```text/plain```: Body is text message
- ```multipart/form-data```: The block of data("text") is text message
- ```application/x-www-form-urlencoded```: ```text=<url encoded text message>```
- ```application/json; charset=utf-8```: 字段都是可选的
```json
{
    "token": "<令牌Token>",
    "title": "<消息标题>",
    "text": "<文本消息内容>",
    "sound": 1,
    "priority": 10,
}
```

支持以下参数：

| 参数名    | 描述                               |
| -------- | --------------------------------- |
| title    | 通知消息的标题                      |
| sound    | `1` 启用声音提示, 其他情况会静音推送   |
| priority | `10` 默认优先级, 或者 `5` 较低优先级  |


例如：

```
http://<address>:<port>/v1/sender/<token>?sound=1&priority=10&title=hello
```

### 发送图片

目前仅支持使用 **POST** 方法通过自建的有状态服务器才能发送图片。

- Content-Type: ```image/png``` 或者 ```image/jpeg```

```bash
cat <jpeg文件路径> | curl -H "Content-Type: image/jpeg" --data-binary @- "http://<address>:<port>/v1/sender/<token>"
```

- Content-Type: ```multipart/form-data```

```bash
$ curl --form "image=@<jpeg文件路径>" "http://<address>:<port>/v1/sender/<token>"
```

## 贡献

贡献使开源社区成为了一个令人赞叹的学习，启发和创造场所。 **十分感谢**您做出的任何贡献。

1. Fork本项目
2. 创建您的Feature分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启一个Pull Request

## 许可证

根据MIT许可证分发，详情查看[`LICENSE`](LICENSE)。
