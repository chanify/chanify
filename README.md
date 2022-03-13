# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Release](https://img.shields.io/github/v/release/chanify/chanify?logo=github&style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/us/app/id1531546573)
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

English | [简体中文](README-zh_CN.md)

Chanify is a safe and simple notification tools. For developers, system administrators, and everyone can push notifications with API.

<details open="open">
  <summary><h2 style="display: inline-block">Table of Contents</h2></summary>
  <ol>
    <li><a href="#features">Features</a></li>
    <li><a href="#getting-started">Getting Started</a></li>
    <li>
        <a href="#installation">Installation</a>
        <ul>
            <li><a href="#precompiled-binary">Precompiled binary</a></li>
            <li><a href="#docker">Docker</a></li>
            <li><a href="#from-source">From source</a></li>
        </ul>
    </li>
    <li>
        <a href="#usage">Usage</a>
        <ul>
            <li><a href="#as-sender-client">As Sender Client</a></li>
            <li><a href="#as-serverless-node">As Serverless node</a></li>
            <li><a href="#as-serverful-node">As Serverful node</a></li>
            <li><a href="#add-new-node">Add New Node</a></li>
            <li>
                <a href="#send-message">Send message</a>
                <ul>
                    <li><a href="#command-line">Command Line</a></li>
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
            <li><a href="#send-text">Send Text</a></li>
            <li><a href="#send-link">Send Link</a></li>
            <li><a href="#send-image">Send Image</a></li>
            <li><a href="#send-audio">Send Audio</a></li>
            <li><a href="#send-file">Send File</a></li>
            <li><a href="#send-actions">Send Actions</a></li>
        </ul>
    </li>
    <li><a href="#configuration">Configuration</a></li>
    <li>
        <a href="#security">Security</a>
        <ul>
            <li><a href="#setting-registrable">Setting Registrable</a></li>
            <li><a href="#token-lifetime">Token Lifetime</a></li>
        </ul>
    </li>
    <li><a href="#chrome-extension">Chrome Extension</a></li>
    <li><a href="#windows-client">Windows Client</a></li>
    <li><a href="#docker-compose">Docker Compose</a></li>
    <li><a href="#lua-api">Lua API</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ol>
</details>

## Features

Chanify is include these features:

- Customize channel for notifications.
- Deploy your own node server.
- Distributed architecture design.
- Design for privacy protection.
- Support text/image/audio/file message format.

## Getting Started

1. Install [iOS App](https://itunes.apple.com/us/app/id1531546573)(1.0.0 version and above) or [macOS App](https://apps.apple.com/us/app/id1531546573)(1.3.0 version and above).
2. Get send token, [more detail](https://github.com/chanify/chanify-ios).
3. Send message.

## Installation

### Precompiled binary

Download precompiled binary from [here](https://github.com/chanify/chanify/releases/latest).

### Docker

```bash
$ docker pull wizjin/chanify:latest
```

### From source

```bash
$ git clone https://github.com/chanify/chanify.git
$ cd chanify
$ make install
```

## Usage

### As Sender Client

Use chanify to send message.

```bash
# Text message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --text=<message>

# URL message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --link=<web url>

# Image message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --image=<image file path>

# Audio message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --audio=<audio file path>

# File message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --file=<file path> --text=<file description>

# Action  message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --text=<message> --title=<title> --action="<action name>|<action url>"

# Timeline message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --timeline.code=<code> <item1>=<value1> <item2>=<value2> ...
```

`endpoint` default value is `https://api.chanify.net`, and notification will send by default server.
If you have own node server, please set `endpoint` to your node server url.

### As Serverless node

Chanify run in stateless mode, no device token will be stored in node.

All device token will be stored in api.chanify.net.

Message will send to apple apns server by api.chanify.net.

Send message workflow:

```text
Start => node server => api.chanify.net => Apple server => iOS client
```

```bash
# Start chanify
$ chanify serve --host=<ip address> --port=<port> --secret=<secret key> --name=<node name> --endpoint=http://<address>:<port>

# Docker
$ docker run -it wizjin/chanify:latest serve --secret=<secret key> --name=<node name> --endpoint=http://<address>:<port>
```

### As Serverful node

Chanify run in stateful mode, device token will be stored in node.

Message will send to apple apns server direct.

Send message workflow:

```text
Start => node server => Apple server => iOS client
```

Start server

```bash
# Start chanify
$ chanify serve --host=<ip address> --port=<port> --name=<node name> --datapath=~/.chanify --endpoint=http://<address>:<port>

# Docker
$ docker run -it -v /my/data:/root/.chanify wizjin/chanify:latest serve --name=<node name> --endpoint=http://<address>:<port>
```

Use MySQL as a backend

```bash
--dburl=mysql://<user>:<password>@tcp(<ip address>:<port>)/<database name>?charset=utf8mb4&parseTime=true&loc=Local
```

Chanify will not create database.

### Add New Node

- Start node server
- iOS client can scan QRCode(`http://<address>:<port>/`) to add node.

### Send message

#### Command Line

```bash
# Text message
$ curl --form-string "text=hello" "http://<address>:<port>/v1/sender/<token>"

# Text file
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

### Send Text

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
- `application/json; charset=utf-8`: The fields are optional
```json
{
    "token": "<token>",
    "title": "<message title>",
    "text": "<text message content>",
    "copy": "<copy text for text message>",
    "autocopy": 1,
    "sound": 1,
    "priority": 10,
    "interruptionlevel": 0,
    "actions": [
        "ActionName1|http://<action host>/<action1>",
        "ActionName2|http://<action host>/<action2>",
        ...
    ],
    "timeline": {
        "code": "<timeline code>",
        "timestamp": 1620000000000,
        "items": {
            "key1": "value1",
            "key2": "value2",
            ...
        }
    }
}
```

Additional params

| Key                | Default  | Description                                      |
| ------------------ | -------- | ------------------------------------------------ |
| title              | None     | The title for notification message.              |
| copy               | None     | The copy text for text notification.             |
| autocopy           | `0`      | Enable autocopy text for text notification.      |
| sound              | `0`      | `1` enable sound, otherwise disable sound.       |
| priority           | `10`     | `10` normal, `5` lower level.                    |
| interruption-level | `active` | Interruption level for timing of a notification. |
| actions            | None     | Actions list.                                    |
| timeline           | None     | Timeline object.                                 |

`sound`:
  - 1 enable default sound
  - sound code, e.g. "bell"

`interruption-level`:
  - `active`: Lights up screen and may play a sound.
  - `passive`: Does not light up screen or play sound.
  - `time-sensitive`: Lights up screen and may play a sound; May be presented during Do Not Disturb.

`timestamp` in milliseconds (timezone - UTC)

E.g.

```url
http://<address>:<port>/v1/sender/<token>?sound=1&priority=10&title=hello&copy=123&autocopy=1
```

Overwrite `Content-Type`

```url
http://<address>:<port>/v1/sender/<token>?content-type=<text|json>
```

### Send Link

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

### Send Image

Send image only support **POST** method used serverful node.

- Content-Type: `image/png` OR `image/jpeg`

```bash
cat <jpeg image path> | curl -H "Content-Type: image/jpeg" --data-binary @- "http://<address>:<port>/v1/sender/<token>"
```

- Content-Type: `multipart/form-data`

```bash
$ curl --form "image=@<jpeg image path>" "http://<address>:<port>/v1/sender/<token>"
```

### Send Audio

Send mp3 audio only support **POST** method used serverful node.

- Content-Type: `audio/mpeg`

```bash
cat <mp3 audio path> | curl -H "Content-Type: audio/mpeg" --data-binary @- "http://<address>:<port>/v1/sender/<token>"
```

- Content-Type: `multipart/form-data`

```bash
$ curl --form "audio=@<mp3 audio path>" "http://<address>:<port>/v1/sender/<token>"
```

### Send File

Send file only support **POST** method used serverful node.

- Content-Type: `multipart/form-data`

```bash
$ curl --form "file=@<file path>" "http://<address>:<port>/v1/sender/<token>"
```

### Send Actions

Send Actions (Up to 4 actions).

- Content-Type: `multipart/form-data`

```bash
$ curl --form "action=ActionName1|http://<action host>/<action1>" "http://<address>:<port>/v1/sender/<token>"
```

## Configuration

Chanify can be configured with a yml format file, and the default path is `~/.chanify.yml`.

```yml
server:
    host: 0.0.0.0   # Listen ip address
    port: 8080      # Listen port
    endpoint: http://my.server/path # Endpoint URL
    name: Node name # Name for node server
    secret: <secret code> # key for serverless node server
    datapath: <data path> # data storage path for serverful node server
#   pluginpath: <plugin path> # plugin path for lua
    dburl: mysql://root:test@tcp(127.0.0.1:3306)/chanify?charset=utf8mb4&parseTime=true&loc=Local # database dsn for serverful node server
    http:
        - readtimeout: 10s  # 10 seconds for http read timeout
        - writetimeout: 10s # 10 seconds for http write timeout
    register:
        enable: false # Disable user register
        whitelist: # whitelist for user register
            - <user id 1>
            - <user id 2>
#   plugin:
#       webhook:
#           - name: github  # POST http://my.server/path/v1/webhook/github/<token>
#             file: webhook/github.lua # <pluginpath>/webhook/github.lua
#             env:
#               secret_token: "secret token"

client: # configuration for sender client
    sound: 1    # enable sound
    endpoint: <default node server endpoint>
    token: <default token>
    interruption-level: <interruption level>
```

## Security

### Setting Registrable

Node server can be disabled user registration and become a private server.

```bash
chanify serve --registerable=false --whitelist=<user1 id>,<user2 id>
```

- `--registerable=false`: used to disable user registration
- `whitelist`: list users can be add into node server

### Token Lifetime

- Token lifetime is about 90 days (default).
- Can configurable token lifetime (1 day ~ 5 years) in channel detail page.

If your token is leaked, add leaked token into the blocklist (iOS client settings).

*Note: Please protect your token from leakage. The blockist need trusted node server (1.1.9 version and above).*

## Chrome Extension

Download the extension from [Chrome web store](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg).

Extension features:

- Send select `text/image/audio/url` message to Chanify
- Send page url to Chanify

## Windows Client

Get the [Windows Client](https://github.com/chanify/chanify-win) from [here](https://github.com/chanify/chanify-win/releases/latest).

Windows Client features:

- Support Chanify to the Windows `Send To` Menu.
- Support send message with `CLI`.

## Docker Compose

1. Install [docker compose](https://docs.docker.com/compose/install).
2. Edit configuration file (`docker-compose.yml`).
3. Start docker compose: `docker-compose up -d`

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

| Key      | Description                     |
| -------- | ------------------------------- |
| workdir  | Work directory for node server. |

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

| Key             | Description                              |
| --------------- | ---------------------------------------- |
| hostname or ip  | Internet hostname or ip for node server. |
| ssl key         | SSL key file for node server.            |

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

| Key             | Description                              |
| --------------- | ---------------------------------------- |
| hostname or ip  | Internet hostname or ip for node server. |
| node name       | Name for node server.                    |
| user id         | User ids for whitelist.                  |

## Lua API

Usage

```lua
local hex = require "hex"
local bytes = hex.decode(hex_string)
local text = hex.encode(bytes_data)

local json = require "json"
local obj = json.decode(json_string)
local str = json.encode(json_object)

local crypto = require "crypto"
local is_equal = crypto.equal(mac1, mac2)
local mac = crypto.hmac("sha1", key, message) -- Support md5 sha1 sha256

-- Http request
local req = ctx:request()
local token_string = req:token()
local http_url = req:url()
local body_string = req:body()
local query_value = req:query("key")
local header_value = req:header("key")

-- Send message
ctx:send({
    title = "message title", -- Optional
    text = "message body",
    sound = "sound or not",  -- Optional
    copy = "copy",           -- Optional
    autocopy = "autocopy",   -- Optional 
})
```

Example: [Github webhook event](plugin/webhook/github.lua)

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Change to dev Branch (`git checkout dev`)
3. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
4. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
5. Push to the Branch (`git push origin feature/AmazingFeature`)
6. Open a Pull Request (merge to `chanify:dev` branch)

## License

Distributed under the MIT License. See [`LICENSE`](LICENSE) for more information.
