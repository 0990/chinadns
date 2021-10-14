SET GOOS=linux
go build -o bin/chinadns cmd/server/main.go

SET GOOS=windows
go build -o bin/chinadns.exe cmd/server/main.go