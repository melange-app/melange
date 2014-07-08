#!bin/sh
rm *.pb.go
protoc source/*.proto --go_out=.
mv source/*.pb.go ./
# go install airdispat.ch/tracker
