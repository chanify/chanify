# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Workflow](https://img.shields.io/github/workflow/status/chanify/chanify/ci?label=build&logo=github&style=flat-square)](https://github.com/chanify/chanify/actions?workflow=ci)
[![Codecov](https://img.shields.io/codecov/c/github/chanify/chanify?logo=codecov&style=flat-square)](https://codecov.io/gh/chanify/chanify)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/app/id1531546573)
[![GitHub](https://img.shields.io/github/license/chanify/chanify?style=flat-square)](LICENSE)

<a href="https://www.producthunt.com/posts/chanify?utm_source=badge-featured&utm_medium=badge&utm_souce=badge-chanify" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=287376&theme=light" alt="Chanify - Safe and simple notification tools | Product Hunt" style="width: 185px; height: 40px;" width="185" height="40" /></a>

Chanify is a safe and simple notification tools. For developers, system administrators, and everyone can push notifications with API.

## Getting Started

1. Install [iOS App](https://itunes.apple.com/us/app/id1531546573)(0.9.10 version and above).
2. Get send token.
3. Send message.

## Installation

#### Command line

```bash
$ go install github.com/chanify/chanify
```

#### Docker

```bash
$ docker pull wizjin/chanify:latest
```

## Usage

### As Sneder Client

Use chanify to send message.

```bash
$ chanify send --token=<token> --text=<message>
```

### As Serverless node

Chanify run in stateless mode, no device token will be stored in node.

All device token will be stored in api.chanify.net.

Message will send to apple apns server by api.chanify.net.

Send message workflow:

```
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

```
Start => node server => Apple server => iOS client
```

Start server

```bash
# Start chanify
$ chanify serve --host=<ip address> --port=<port> --name=<node name> --datapath=~/.chanify

# Docker 
$ docker run -it -v /my/data:/root/.chanify wizjin/chanify:latest serve --name=<node name> --endpoint=http://<address>:<port>
```

### Add New Node

- Start node server
- iOS client can scan QRCode(```http://<address>:<port>/```) to add node.

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

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

Distributed under the MIT License. See [`LICENSE`](LICENSE) for more information.
