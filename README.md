# GNU Tox Party

## Dependencies
* [Tox](https://github.com/TokTok/c-toxcore)
* [Go](https://golang.org/)
* [go-toxcore](https://github.com/TokTok/go-toxcore-c)
* [go-ircevent](https://github.com/thoj/go-ircevent)

## Installation
```bash
$ sudo pacman -Sy toxcore go pkg-config
$ go get github.com/TokTok/go-toxcore-c
$ go get github.com/thoj/go-ircevent
```

## Configuration
You need to rename *config.sample.json* as *config.json*

## Run
```bash
$ go run main.go
```