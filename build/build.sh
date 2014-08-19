#! /bin/bash

mkdir tmp

# -a -v
go build -o=tmp/server ../server/server/server.go
go build -o=tmp/updater ../server/updater/updater/updater.go

cp -R ../releases/atom/Melange.app tmp/

rm -rf tmp/Melange.app/Contents/Resources/default_app

mkdir tmp/Melange.app/Contents/Resources/app

cp -R ../client tmp/Melange.app/Contents/Resources/app
cp -R ../lib tmp/Melange.app/Contents/Resources/app
cp -R ../plugins tmp/Melange.app/Contents/Resources/app
cp -R ../*.html tmp/Melange.app/Contents/Resources/app
cp -R ../*.js tmp/Melange.app/Contents/Resources/app
cp -R ../*.json tmp/Melange.app/Contents/Resources/app

mkdir tmp/Melange.app/Contents/Resources/app/bin

mv tmp/server tmp/Melange.app/Contents/Resources/app/bin/
mv tmp/updater tmp/Melange.app/

# rm -rf tmp
