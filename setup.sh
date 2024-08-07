#!/bin/bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

mkdir -p ~/opt && mkdir -p ~/opt/semantifly && cd ~/opt/semantifly

# Download and install protoc
wget https://github.com/protocolbuffers/protobuf/releases/download/v27.3/protoc-27.3-linux-x86_64.zip
unzip protoc-27.3-linux-x86_64.zip
rm protoc-27.3-linux-x86_64.zip

cd -
cd src
go get -d ./...

# Build binary, move to PATH
go build -o semantifly .
sudo mv semantifly /usr/local/bin/

# Update PATH
. ~/.bashrc

echo "You can now use 'semantifly' command from anywhere"