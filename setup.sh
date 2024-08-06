go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
mkdir -p ~/opt && mkdir -p ~/opt/semantifly && cd ~/opt/semantifly
wget https://github.com/protocolbuffers/protobuf/releases/download/v27.3/protoc-27.3-linux-x86_64.zip
unzip protoc-27.3-linux-x86_64.zip
export PATH=$PATH:~/opt/semantifly/bin
echo "Protoc installed to ~/opt/semantifly/bin: fix pending to add to PATH"
echo "For now, invoke as eg '~/opt/semantifly/bin/protoc -I=src/proto --go_out=src/proto src/proto/index.proto'"
