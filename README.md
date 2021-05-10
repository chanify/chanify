# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Release](https://img.shields.io/github/v/release/chanify/chanify?logo=github&style=flat-square)](https://github.com/chanify/chanify/releases/latest)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/us/app/id1531546573)
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
            <li><a href="#send-file">Send File</a></li>
        </ul>
    </li>
    <li><a href="#configuration">Configuration</a></li>
    <li><a href="#security">Security</a></li>
    <li><a href="#chrome-extension">Chrome Extension</a></li>
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
- Support text/image/file message format.

## Getting Started

1. Install [iOS App](https://itunes.apple.com/us/app/id1531546573)(1.0.0 version and above).
2. Get send token, [more detail](https://github.com/chanify/chanify-ios).
3. Send message.

## Installation

### Precompiled binary

Download precompiled binary from [this](https://github.com/chanify/chanify/releases/latest).

### Docker

```bash
$ docker pull wizjin/chanify:latest
```

### From source

```bash
$ go install github.com/chanify/chanify
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

# File message
$ chanify send --endpoint=http://<address>:<port> --token=<token> --file=<file path> --text=<file description>
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
}
```

Additional params

| Key      | Default | Description                                 |
| -------- | ------- | ------------------------------------------- |
| title    | None    | The title for notification message.         |
| copy     | None    | The copy text for text notification.        |
| autocopy | `0`     | Enable autocopy text for text notification. |
| sound    | `0`     | `1` enable sound, otherwise disable sound.  |
| priority | `10`    | `10` normal, `5` lower level.               |

E.g.

```url
http://<address>:<port>/v1/sender/<token>?sound=1&priority=10&title=hello&copy=123&autocopy=1
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

### Send File

Send file only support **POST** method used serverful node.

- Content-Type: `multipart/form-data`

```bash
$ curl --form "file=@<file path>" "http://<address>:<port>/v1/sender/<token>"
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
    dburl: mysql://root:test@tcp(127.0.0.1:3306)/chanify?charset=utf8mb4&parseTime=true&loc=Local # database dsn for serverful node server
    register:
        enable: false # Disable user register
        whitelist: # whitelist for user register
            - <user id 1>
            - <user id 2>

client: # configuration for sender client
    sound: 1    # enable sound
    endpoint: <default node server endpoint>
    token: <default token>
```

## Security

Node server can be disabled user registration and become a private server.

```bash
chanify serve --registerable=false --whitelist=<user1 id>,<user2 id>
```

- `--registerable=false`: used to disable user registration
- `whitelist`: list users can be add into node server

## Chrome Extension

Download the extension for [Chrome web store](https://chrome.google.com/webstore/detail/chanify/llpdpmhkemkjeeigibdamadahmhoebdg).

Extension features:

- Send select `text/image/url` message to Chanify
- Send page url to Chanify

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
