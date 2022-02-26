# Copyright 2016-17, OpenCensus Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

workspace(name = "opencensus_proto")

load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

# Import rules_proto for proto_library().
http_archive(
    name = "rules_proto",
    sha256 = "602e7161d9195e50246177e7c55b2f39950a9cf7366f74ed5f22fd45750cd208",
    strip_prefix = "rules_proto-97d8af4dc474595af3900dd85cb3a29ad28cc313",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_proto/archive/97d8af4dc474595af3900dd85cb3a29ad28cc313.tar.gz",
        "https://github.com/bazelbuild/rules_proto/archive/97d8af4dc474595af3900dd85cb3a29ad28cc313.tar.gz",
    ],
)
load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")
rules_proto_dependencies()
rules_proto_toolchains()

# Import grpc for cc_grpc_library().
http_archive(
    name = "com_github_grpc_grpc",
    sha256 = "bbda0588df1b31aca41b83d09cf69816df653b5a89f820d16ab95af8eb1f9072",
    strip_prefix = "grpc-b1547c9d19034740b1e5b4d73cb91d068dc0acb4",
    urls = ["https://github.com/grpc/grpc/archive/b1547c9d19034740b1e5b4d73cb91d068dc0acb4.zip"],
)

load("@com_github_grpc_grpc//bazel:grpc_deps.bzl", "grpc_deps")
grpc_deps()

# grpc_deps() cannot load() its deps, this WORKSPACE has to do it.
# See also: https://github.com/bazelbuild/bazel/issues/1943
load(
    "@build_bazel_rules_apple//apple:repositories.bzl",
    "apple_rules_dependencies",
)

apple_rules_dependencies()

load(
    "@build_bazel_apple_support//lib:repositories.bzl",
    "apple_support_dependencies",
)

apple_support_dependencies()

load("@upb//bazel:repository_defs.bzl", "bazel_version_repository")

bazel_version_repository(name = "bazel_version")

# go rules related
git_repository(
    name = "io_bazel_rules_go",
    remote = "https://github.com/bazelbuild/rules_go.git",
    # we need rules_go of tag 0.12.0 or greater
    # which includes golang protobuf that supports "paths=source_relative"
    # the "paths=source_relative" option for protoc-gen-go makes bazel works as expect
    # when go_package was declared in proto files
    # see https://github.com/bazelbuild/rules_go/blob/release-0.12/go/private/repositories.bzl#L75
    # for the included golang protobuf version and
    # see https://github.com/golang/protobuf/pull/544 for "paths=source_relative" usage
    tag = "v0.19.9",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")
go_rules_dependencies()
go_register_toolchains()

# py rules related
http_archive(
    name = "six_archive",
    build_file = "@com_google_protobuf//:six.BUILD",
    sha256 = "105f8d68616f8248e24bf0e9372ef04d3cc10104f1980f54d57b2ce73a5ad56a",
    urls = ["https://pypi.python.org/packages/source/s/six/six-1.10.0.tar.gz#md5=34eed507548117b2ab523ab14b2f8b55"],
)

bind(
    name = "six",
    actual = "@six_archive//:six",
)
