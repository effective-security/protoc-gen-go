#!/usr/bin/env bash
#
# Generate all project's protobuf bindings.
#
set -e

#
# genproto.sh
#   --dirs {dird}       - specifies dirs to build
#   --i {path}          - specifies path for imports
#   --oapi {path}       - specifies path for Open API output
#   --json {path}       - specifies to generate --go-json_out
#   --mock {path}       - specifies to generate --go-mock_out
#   --proxy {path}      - specifies to generate --go-proxy_out
#   --methods {path}    - specifies to generate --go-allocator_out
#   --python {path}     - specifies to generate --python_out
#   --ts {path}         - specifies to generate --grpc-web_out
#   --out               - specifies to output folder, default is '.'

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --dirs)
    DIRS="$2"
    shift # past argument
    shift # past value
    ;;
    --files)
    FILES="$2"
    shift # past argument
    shift # past value
    ;;
    --i)
    IMPORTS="$2"
    shift # past argument
    shift # past value
    ;;
    --oapi)
    OPENAPI="--openapi_out=$2"
    shift # past argument
    shift # past value
    ;;
    --json)
    JSON="--go-json_out=orig_name=true,partial=true,multiline=true,paths=source_relative:$2"
    shift # past argument
    shift # past value
    ;;
    --mock)
    MOCK="--go-mock_out=logs=true:$2"
    shift # past argument
    shift # past value
    ;;
    --proxy)
    PROXY="--go-proxy_out=logs=true:$2"
    shift # past argument
    shift # past value
    ;;
    --http)
    HTTP="--go-http_out=logs=true:$2"
    shift # past argument
    shift # past value
    ;;
    --methods)
    METHODS="--go-allocator_out=logs=true:$2"
    shift # past argument
    shift # past value
    ;;
    --golang)
    GOLANG="--go_out=paths=source_relative:$2 --go-grpc_out=require_unimplemented_servers=false,paths=source_relative:$2"
    shift # past argument
    shift # past value
    ;;
    --python)
    PYTHON="--python_out=$2 --pyi_out=$2"
    shift # past argument
    shift # past value
    ;;
    --ts)
    TS="--js_out=import_style=commonjs,binary:$2 --grpc-web_out=import_style=typescript,mode=grpcweb:$2"
    shift # past argument
    shift # past value
    ;;
    *)
    echo "invalid flag $key: use --help to see the option"
    exit 1
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

[ -z "$DIRS" ] &&  echo "Specify --dirs" && exit 1
[ -z "$FILES" ] && FILES="./*.proto"

echo "DIRS    = $DIRS"
echo "FILES   = $FILES"
echo "IMPORTS = $IMPORTS"
echo "OPENAPI = $OPENAPI"
echo "JSON    = $JSON"
echo "MOCK    = $MOCK"
echo "PROXY   = $PROXY"
echo "HTTP    = $HTTP"
echo "METHODS = $METHODS"
echo "GOLANG  = $GOLANG"
echo "TS      = $TS"

for dir in $DIRS; do
	pushd "$dir"
        echo "running protoc in $dir"
		protoc $IMPORTS $OPENAPI \
			-I=. \
			-I=/usr/local/include \
            $JSON $MOCK $PROXY $METHODS $HTTP $GOLANG $PYTHON $TS $FILES
	popd
done
