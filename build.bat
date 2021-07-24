SET GOOS=linux
go build -o bin/chinadns cmd/main.go

SET GOOS=windows
go build -o bin/chinadns.exe cmd/main.go