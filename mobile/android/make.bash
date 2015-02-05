#!/usr/bin/env bash
# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -e

if [ ! -f make.bash ]; then
	echo 'make.bash must be run from $GOPATH/src/getmelange.com/mobile'
	exit 1
fi

mkdir -p libs/armeabi-v7a src/go/melange
ANDROID_APP=$PWD

# Copy Go Libraries
# (cd ../.. && ln -sf $PWD/app/Go.java $ANDROID_APP/src/go)
# (cd ../.. && cp $PWD/golang.org/x/mobile/app/Go.java $ANDROID_APP/src/go)
# (cd ../.. && ln -sf $PWD/bind/java/Seq.java $ANDROID_APP/src/go)
# (cd ../.. && cp $PWD/golang.org/x/mobile/bind/java/Seq.java $ANDROID_APP/src/go)

CGO_ENABLED=1 GOOS=android GOARCH=arm GOARM=7 \
	go build -o="libmelange" -ldflags="-shared" .
mv -f libmelange libs/armeabi-v7a/libgojni.so
ant release
