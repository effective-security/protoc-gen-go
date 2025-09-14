#!/usr/bin/env bash
#
# Generate all project's protobuf bindings.
#
set -e

#
# genproto.sh
#   --dirs {dird}               - specifies dirs to build
#   --i {path}                  - specifies path for imports
#   --oapi {path}               - specifies path for Open API output
#   --json {path}               - specifies to generate --go-json_out
#   --mock {path}               - specifies to generate --go-mock_out
#   --proxy {path}              - specifies to generate --go-proxy_out
#   --http {pkg} {path}         - specifies to generate --go-http_out
#   --methods {path}            - specifies to generate --go-allocator_out
#   --enum {path}               - specifies to generate --go-enum_out
#   --python {path}             - specifies to generate --python_out
#   --ts-enum {import} {path}   - specifies to generate --ts-enum_out
#   --csharp {path}             - specifies to generate --csharp_out
#   --out                       - specifies to output folder, default is '.'

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
    JSON="--go-json_out=enums_as_ints=true,orig_name=true,partial=true,multiline=false,paths=source_relative:$2"
    shift # past argument
    shift # past value
    ;;
    --enum)
    ENUM="--go-enum_out=logs=true:$2"
    shift # past argument
    shift # past value
    ;;
    --ts-enum)
    TS="--ts-enum_out=logs=true,import=$2:$3"
    shift # past argument
    shift # past import value
    shift # past path value
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
    HTTP="--go-http_out=logs=true,pbpkg=$2:$3"
    shift # past argument
    shift # past pkg value
    shift # past path value
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
    --csharp)
    CS="--csharp_opt=serializable,file_extension=.gen.cs --grpc_opt=no_server --grpc_out=$2 --plugin=protoc-gen-grpc=/usr/local/bin/grpc_csharp_plugin --csharp_out=$2"
    shift # past argument
    shift # past value
    ;;
    *)
    echo "genproto.sh: invalid flag $key: use --help to see the option"
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
echo "ENUM    = $ENUM"
echo "MOCK    = $MOCK"
echo "PROXY   = $PROXY"
echo "HTTP    = $HTTP"
echo "METHODS = $METHODS"
echo "GOLANG  = $GOLANG"
echo "TS      = $TS"
echo "CS      = $CS"

for dir in $DIRS; do
	pushd "$dir"
        echo "running protoc in $dir"
		protoc $IMPORTS $OPENAPI \
			-I=. \
			-I=/usr/local/include \
            $JSON $ENUM $MOCK $PROXY $METHODS $HTTP $GOLANG $PYTHON $TS $CS $FILES
	popd
done
