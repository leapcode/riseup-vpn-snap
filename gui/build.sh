#!/bin/bash
set -e
set -x

XBUILD=${XBUILD-no}
OSX_TARGET=10.11
WIN64="win64"
GO=`which go`

PROJECT=bitmask.pro
TARGET_GOLIB=lib/libgoshim.a
SOURCE_GOLIB=gui/backend.go

QTBUILD=build/qt
RELEASE=$QTBUILD/release

PLATFORM=$(uname -s)
LDFLAGS=""

if [ "$TARGET" == "" ]
then
    TARGET=riseup-vpn
fi

if [ "$XBUILD" == "$WIN64" ]
then
    # TODO allow to override vars
    QMAKE="`pwd`/../../mxe/usr/x86_64-w64-mingw32.static/qt5/bin/qmake"
    PATH="`pwd`/../../mxe/usr/bin"/:$PATH
    CC=x86_64-w64-mingw32.static-gcc
else
    if [ "$QMAKE" == "" ]
    then
        QMAKE=`which qmake`
    fi
fi

PLATFORM=`uname -s`

function init {
    mkdir -p lib
}

# TODO this should be moved to the makefile
function buildGoLib {
    echo "[+] Using go in" $GO "[`go version`]"
    $GO generate ./pkg/config/version/genver/gen.go
    if [ "$PLATFORM" == "Darwin" ]
    then
        GOOS=darwin
	CC=clang
	CGO_CFLAGS="-g -O2 -mmacosx-version-min=$OSX_TARGET"
	CGO_LDFLAGS="-g -O2 -mmacosx-version-min=$OSX_TARGET"
    fi
    if [ "$PLATFORM" == "MINGW64_NT-10.0" ]
    then
	LDFLAGS="-H=windowsgui"
    fi
    if [ "$XBUILD" == "no" ]
    then
        echo "[+] Building Go library with standard Go compiler"
        CGO_ENABLED=1 GOOS=$GOOS CC=$CC CGO_CFLAGS=$CGO_CFLAGS CGO_LDFLAGS=$CGO_LDFLAGS go build -ldflags $LDFLAGS -buildmode=c-archive -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
    if [ "$XBUILD" == "$WIN64" ]
    then
        echo "[+] Building Go library with mxe"
        echo "[+] Using cc:" $CC
        CC=$CC CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -buildmode=c-archive -o $TARGET_GOLIB $SOURCE_GOLIB
    fi
}

function buildQmake {
    echo "[+] Now building Qml app with Qt qmake"
    echo "[+] Using qmake in:" $QMAKE
    mkdir -p $QTBUILD
    $QMAKE -o $QTBUILD/Makefile "CONFIG-=debug CONFIG+=release" $PROJECT
}

function renameOutput {
    # i would expect that passing QMAKE_TARGET would produce the right output, but nope.
    if [ "$PLATFORM" == "Linux" ]
    then
    	mv $RELEASE/bitmask $RELEASE/$TARGET
    	strip $RELEASE/$TARGET
    	echo "[+] Binary is in" $RELEASE/$TARGET
    fi
    if [ "$PLATFORM" == "Darwin" ]
    then
    	rm -rf $RELEASE/$TARGET.app
    	mv $RELEASE/bitmask.app/ $RELEASE/$TARGET.app/
    	echo "[+] App is in" $RELEASE/$TARGET
    fi
    if [ "$PLATFORM" == "MINGW64_NT-10.0" ]
    then
    	mv $RELEASE/bitmask.exe $RELEASE/$TARGET.exe
    fi
}

echo "[+] Building BitmaskVPN"

lrelease bitmask.pro

# FIXME move buildGoLib to the main makefile, to avoid redundant builds if possible
buildGoLib
buildQmake

make -C $QTBUILD clean
make -C $QTBUILD -j4 all

renameOutput
echo "[+] Done."
