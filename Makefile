binaries: ru1 ru1.openbsd

ru1: *.go go.sum go.mod
	go build

ru1.openbsd: *.go go.sum go.mod
	GOOS=openbsd go build -o ru1.openbsd

clean:
	$(RM) ru1 ru1.openbsd
