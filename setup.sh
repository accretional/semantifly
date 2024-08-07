#!/bin/bash

# Install protoc-gen-go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Create directories
mkdir -p ~/opt/semantifly
cd ~/opt/semantifly

# Download and install protoc
wget https://github.com/protocolbuffers/protobuf/releases/download/v27.3/protoc-27.3-linux-x86_64.zip
unzip protoc-27.3-linux-x86_64.zip
rm protoc-27.3-linux-x86_64.zip  # Clean up the zip file

# Navigate back to the script's directory (which should be the Semantifly repo)
cd - 

# Ensure we're in the directory containing src/main.go
if [ ! -f "./src/main.go" ]; then
    echo "Error: src/main.go not found. Make sure you're running this script from the root of the Semantifly repository."
    exit 1
fi

# Install dependencies
go get -d ./...

# Build the binary
go build -o semantifly ./src

# Move the binary to a directory in PATH
sudo mv semantifly /usr/local/bin/

# Update PATH
echo 'export PATH=$PATH:~/opt/semantifly/bin:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc

echo "Protoc installed to ~/opt/semantifly/bin"
echo "Semantifly binary installed to /usr/local/bin"
echo "Please restart your terminal or run 'source ~/.bashrc' to update your PATH"
echo "You can now use 'semantifly' command from anywhere"
echo "For protoc, you can invoke as: 'protoc -I=src/proto --go_out=src/proto src/proto/index.proto'"