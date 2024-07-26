all: buildserver buildagent

buildserver:
	GOOS=windows go build -o bin/server.exe cmd/server/main.go

buildagent:
	GOOS=windows go build -o bin/agent.exe cmd/agent/main.go