SET GOOS=linux
go build -o bin/chinadns cmd/server/main.go
go build -o bin/gfw cmd/gfw/main.go

SET GOOS=windows
go build -o bin/chinadns.exe cmd/server/main.go
go build -o bin/gfw.exe cmd/gfw/main.go