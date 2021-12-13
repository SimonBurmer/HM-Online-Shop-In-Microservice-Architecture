# Zeile im terminal ausführen. Achtung du musst in einem dic sein in dem es ein messagges.proto gibt!
messages.pb.go: api.proto
	go install github.com/gogo/protobuf/protoc-gen-gogoslick
	protoc -I=.  --gogoslick_out=plugins=grpc:. api.proto 

