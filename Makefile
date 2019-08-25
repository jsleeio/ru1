binaries: ru1 ru1.openbsd

ru1: *.go
	go build

ru1.openbsd: *.go
	GOOS=openbsd go build -o ru1.openbsd
