SET GOOS=linux
SET GOARCH=amd64
go build -o bin/chinadns cmd/server/main.go

SET GOOS=windows
SET GOARCH=amd64
go build -o bin/chinadns.exe cmd/server/main.go

SET GOOS=linux
SET GOARCH=arm64
go build -o bin/chinadns_arm cmd/server/main.go