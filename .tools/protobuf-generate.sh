#!/bin/bash

INPUT_FILE=""

if [ '$1' = '' ]; then
  echo "protobuf-generate.sh <input file>"
fi

INPUT_FILE=$1

PATHPREFIX="github.com/superp00t/gophercraft/bnet/bgs"

go install -v github.com/superp00t/gophercraft/cmd/protoc-gen-gcraft

cd $GOPATH/src/
protoc github.com/superp00t/gophercraft/bnet/public_protos/Login.proto --gcraft_out=. 
protoc github.com/superp00t/gophercraft/bnet/public_protos/RealmList.proto --gcraft_out=. 

cd $GOPATH/src/github.com/superp00t/gophercraft/

echo "$PATHPREFIX"

echo extracting

cd bnet

rm -rf proto_src

rm -rf protos
mkdir protos

OPATH=$GOPATH/src/github.com/superp00t/gophercraft/bnet/bgs
IPATH=$GOPATH/src/github.com/superp00t/gophercraft/bnet/proto_src/bgs/low/pb/client/
GPATH=$GOPATH/src/github.com/superp00t/gophercraft/bnet/protos/

./pbtk/extractors/from_binary.py "$INPUT_FILE" proto_src

mv proto_src/bgs/low/pb/client/* $GPATH

rm -rf proto_src

cd $GPATH

find .  -type f -exec sed -i "s#bgs/low/pb/client/##g" "{}" \;

go run github.com/superp00t/gophercraft/cmd/gcraft_protobuf_fix
