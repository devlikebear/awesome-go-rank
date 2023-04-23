# go build for ubuntu@latest

# build for linux
env GOOS=linux GOARCH=amd64 go build -o ./bin/linux/amd64/awesome-go-rank ./cmd/main.go
