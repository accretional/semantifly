#!/bin/bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

mkdir -p ~/opt/semantifly
cd ~/opt/semantifly

# Download and install protoc
wget https://github.com/protocolbuffers/protobuf/releases/download/v27.3/protoc-27.3-linux-x86_64.zip
unzip protoc-27.3-linux-x86_64.zip
rm protoc-27.3-linux-x86_64.zip

cd - 
if [ ! -f "./src/main.go" ]; then
    echo "Error: src/main.go not found. Make sure you're running this script from the root of the Semantifly repository."
    exit 1
fi

cd src
go get -d ./...

# Build binary, move to PATH
go build -o semantifly .
sudo mv semantifly /usr/local/bin/

# Update PATH
cd ..
echo 'export PATH=$PATH:~/opt/semantifly/bin:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc

echo "You can now use 'semantifly' command from anywhere"
echo "For protoc, you can invoke as: 'protoc -I=src/proto --go_out=src/proto src/proto/index.proto'"