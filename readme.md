gode
===
a node server replacement with go.


usage
===

```
$ go build -race cmd/web_server/web_server.go
$ cp .env.example .env
$ ./web_server
```

執行後會在 port:80 listen /casino/{game_type} 並轉接到 flash2db

testing
===
執行所有的測試
```
$ go test -v ./... -race
```