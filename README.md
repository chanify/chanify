# Chanify

[![Docker](https://img.shields.io/docker/v/wizjin/chanify?sort=semver&logo=docker&style=flat-square)](https://hub.docker.com/r/wizjin/chanify)
[![Workflow](https://img.shields.io/github/workflow/status/chanify/chanify/ci?label=build&logo=github&style=flat-square)](https://github.com/chanify/chanify/actions?workflow=ci)
[![Codecov](https://img.shields.io/codecov/c/github/chanify/chanify?logo=codecov&style=flat-square)](https://codecov.io/gh/chanify/chanify)
[![iTunes App Store](https://img.shields.io/itunes/v/1531546573?logo=apple&style=flat-square)](https://itunes.apple.com/app/id1531546573)
[![GitHub](https://img.shields.io/github/license/chanify/chanify?style=flat-square)](LICENSE)

<a href="https://www.producthunt.com/posts/chanify?utm_source=badge-featured&utm_medium=badge&utm_souce=badge-chanify" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=287376&theme=light" alt="Chanify - Safe and simple notification tools | Product Hunt" style="width: 185px; height: 40px;" width="185" height="40" /></a>

***WARNING: Node server api is an incomplete work-in-progress.***

Chanify is a safe and simple notification tools. For developers, system administrators, and everyone can push notifications with API.

## Getting Started

1. Install [iOS App](https://itunes.apple.com/us/app/id1531546573).
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

#### Use chanify to send message.

```bash
$ chanify send --token=<token> --text=<message>
```

#### Start chanify node server

```bash
$ chanify serve --host=0.0.0.0 --port=8080 --secret=<secret key> --name=<node name>
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
