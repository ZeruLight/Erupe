#!/usr/bin/env bash

set -e

readonly PROTOC_VER="3.12.3"
readonly PROTOC_ARCHIVE="protoc-${PROTOC_VER}-linux-$( uname -m ).zip"
readonly PROTOC_URL="https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VER}/${PROTOC_ARCHIVE}"

readonly PROTOC_GEN_GO_VER="1.25.0"
readonly PROTOC_GEN_GO_ARCHIVE="protoc-gen-go.v${PROTOC_GEN_GO_VER}.linux.amd64.tar.gz"
readonly PROTOC_GEN_GO_URL="https://github.com/protocolbuffers/protobuf-go/releases/download/v${PROTOC_GEN_GO_VER}/${PROTOC_GEN_GO_ARCHIVE}"

readonly PROTOC_GEN_GO_GRPC_VER="v1.30.0"
readonly PROTOC_GEN_GO_GRPC_REPO_URL="https://github.com/grpc/grpc-go"

readonly GRPC_ECOSYSTEM_VER="v1.14.6"
readonly GRPC_ECOSYSTEM_URL="https://github.com/grpc-ecosystem/grpc-gateway/releases/download/${GRPC_ECOSYSTEM_VER}"

readonly PROTOC_GEN_SWAGGER_BIN="protoc-gen-swagger-${GRPC_ECOSYSTEM_VER}-linux-$( uname -m )"
readonly PROTOC_GEN_SWAGGER_URL="${GRPC_ECOSYSTEM_URL}/${PROTOC_GEN_SWAGGER_BIN}"

readonly PROTOC_GEN_GRPC_GATEWAY_BIN="protoc-gen-grpc-gateway-${GRPC_ECOSYSTEM_VER}-linux-$( uname -m )"
readonly PROTOC_GEN_GRPC_GATEWAY_URL="${GRPC_ECOSYSTEM_URL}/${PROTOC_GEN_GRPC_GATEWAY_BIN}"

readonly DEFAULT_DEST="/usr/local"

get() {
    local url="$1"
    local dest="$2"

    if [[ -w $dest ]]
    then
        wget --quiet "$url" -O "$dest"
    else
        sudo wget --quiet "$url" -O "$dest"
    fi
}

install_bin() {
    local dest_dir="${1}/bin"
    local source_bin="$2"
    local source_url="$3"
    local target_bin="$4"
    local auth

    if [[ ! -w $dest_dir ]]
    then
        auth="sudo"
    fi

    [ -d "${dest_dir}" ] || $auth mkdir -p "${dest_dir}"

    get "$source_url" "${dest_dir}/$source_bin"
    $auth chmod +x "${dest_dir}/$source_bin"
    $auth ln -s -f "${dest_dir}/${source_bin}" "${dest_dir}/$target_bin"

    echo "installed $source_bin to ${dest_dir}/${target_bin}"
}

install_protoc() {
    local dest="$1"
    local output="${dest}/protoc-${PROTOC_VER}"
    local tmpfile auth

    tmpfile="$( mktemp )"
    get "$PROTOC_URL" "$tmpfile"

    if [[ ! -w $dest ]]
    then
        auth="sudo"
    fi

    $auth unzip -qq -o -d "$output" "$tmpfile"
    rm -rf "$tmpfile"
    [ -d "${dest}/bin" ] || $auth mkdir "${dest}/bin"
    $auth ln -s -f "${output}/bin/protoc"  "${dest}/bin/protoc"

    echo "installed protoc v$PROTOC_VER to ${dest}/bin/protoc"
}

install_protoc_gen_go() {
    local dest="$1"
    local tmpfile auth

    tmpfile="$( mktemp )"
    get "$PROTOC_GEN_GO_URL" "$tmpfile"

    if [[ ! -w $dest ]]
    then
        auth="sudo"
    fi

    $auth tar -xzf "$tmpfile" -C "${dest}"
    rm -rf "$tmpfile"
    [ -d "${dest}/bin" ] || $auth mkdir "${dest}/bin"
    $auth ln -s -f "${dest}/protoc-gen-go"  "${dest}/bin/protoc-gen-go"

    echo "installed protoc-gen-go v${PROTOC_GEN_GO_VER} to ${dest}/bin/protoc-gen-go"
}

# TODO Switch to binary once https://github.com/grpc/grpc.io/issues/298 is resolved
install_protoc_gen_go_grpc() {
    local parent_dir="$( go env GOPATH )/src/github.com/grpc"
    local checkout_dir="${parent_dir}/grpc-go"

    rm -rf ${checkout_dir}
    mkdir -p ${parent_dir}
    cd ${parent_dir}
    git clone --quiet -c advice.detachedHead=false -b ${PROTOC_GEN_GO_GRPC_VER} ${PROTOC_GEN_GO_GRPC_REPO_URL}
    cd grpc-go/cmd/protoc-gen-go-grpc
    go install .

    echo "installed protoc-gen-go-grpc v$PROTOC_VER"
}

install_protoc_gen_swagger() {
    local dest="$1"

    install_bin \
        "$dest"\
        "$PROTOC_GEN_SWAGGER_BIN"\
        "$PROTOC_GEN_SWAGGER_URL"\
        "protoc-gen-swagger"
}

install_protoc_gen_grpc_gateway() {
    local dest="$1"

    install_bin \
        "$dest"\
        "$PROTOC_GEN_GRPC_GATEWAY_BIN"\
        "$PROTOC_GEN_GRPC_GATEWAY_URL"\
        "protoc-gen-grpc-gateway"
}

main() {
    local dest="${1:-$DEFAULT_DEST}"

    install_protoc "$dest"
    install_protoc_gen_go "$dest"
    install_protoc_gen_go_grpc
    install_protoc_gen_swagger "$dest"
    install_protoc_gen_grpc_gateway "$dest"
}

main "$@"