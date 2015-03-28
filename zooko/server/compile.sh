#!bin/sh
rm *.pb.go
protoc source/*.proto --go_out=.
mv source/*.pb.go ./
go install getmelange.com/zooko/server
