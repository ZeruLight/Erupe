/*
 * Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License").
 * You may not use this file except in compliance with the License.
 * A copy of the License is located at
 *
 *  http://aws.amazon.com/apache2.0
 *
 * or in the "license" file accompanying this file. This file is distributed
 * on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package software.amazon.smithy.go.codegen;

/**
 * A class of constants for dependencies used by this package.
 */
public final class SmithyGoDependency {
    // The version in the stdlib dependencies should reflect the minimum Go version.
    // The values aren't currently used, but they could potentially used to dynamically
    // set the minimum go version.
    public static final GoDependency BIG = stdlib("math/big");
    public static final GoDependency TIME = stdlib("time");
    public static final GoDependency FMT = stdlib("fmt");
    public static final GoDependency CONTEXT = stdlib("context");
    public static final GoDependency STRCONV = stdlib("strconv");
    public static final GoDependency BASE64 = stdlib("encoding/base64");
    public static final GoDependency NET_URL = stdlib("net/url");
    public static final GoDependency NET_HTTP = stdlib("net/http");
    public static final GoDependency NET_HTTP_TEST = stdlib("net/http/httptest");
    public static final GoDependency BYTES = stdlib("bytes");
    public static final GoDependency STRINGS = stdlib("strings");
    public static final GoDependency JSON = stdlib("encoding/json");
    public static final GoDependency IO = stdlib("io");
    public static final GoDependency IOUTIL = stdlib("io/ioutil");
    public static final GoDependency CRYPTORAND = stdlib("crypto/rand", "cryptorand");
    public static final GoDependency TESTING = stdlib("testing");
    public static final GoDependency ERRORS = stdlib("errors");
    public static final GoDependency XML = stdlib("encoding/xml");
    public static final GoDependency SYNC = stdlib("sync");

    public static final GoDependency SMITHY = smithy(null, "smithy");
    public static final GoDependency SMITHY_HTTP_TRANSPORT = smithy("transport/http", "smithyhttp");
    public static final GoDependency SMITHY_MIDDLEWARE = smithy("middleware");
    public static final GoDependency SMITHY_TIME = smithy("time", "smithytime");
    public static final GoDependency SMITHY_HTTP_BINDING = smithy("encoding/httpbinding");
    public static final GoDependency SMITHY_JSON = smithy("encoding/json", "smithyjson");
    public static final GoDependency SMITHY_XML = smithy("encoding/xml", "smithyxml");
    public static final GoDependency SMITHY_IO = smithy("io", "smithyio");
    public static final GoDependency SMITHY_LOGGING = smithy("logging");
    public static final GoDependency SMITHY_PTR = smithy("ptr");
    public static final GoDependency SMITHY_RAND = smithy("rand", "smithyrand");
    public static final GoDependency SMITHY_TESTING = smithy("testing", "smithytesting");
    public static final GoDependency SMITHY_WAITERS = smithy("waiter", "smithywaiter");
    public static final GoDependency SMITHY_DOCUMENT = smithy("document", "smithydocument");
    public static final GoDependency SMITHY_DOCUMENT_JSON = smithy("document/json", "smithydocumentjson");
    public static final GoDependency SMITHY_SYNC = smithy("sync", "smithysync");

    public static final GoDependency GO_CMP = goCmp("cmp");
    public static final GoDependency GO_CMP_OPTIONS = goCmp("cmp/cmpopts");

    public static final GoDependency GO_JMESPATH = goJmespath(null);
    public static final GoDependency MATH = stdlib("math");

    private static final String SMITHY_SOURCE_PATH = "github.com/aws/smithy-go";
    private static final String GO_CMP_SOURCE_PATH = "github.com/google/go-cmp";
    private static final String GO_JMESPATH_SOURCE_PATH = "github.com/jmespath/go-jmespath";

    private SmithyGoDependency() {
    }

    /**
     * Get a {@link GoDependency} representing the standard library package import path.
     *
     * @param importPath standard library import path
     * @return the {@link GoDependency} for the package import path
     */
    public static GoDependency stdlib(String importPath) {
        return GoDependency.standardLibraryDependency(importPath, Versions.GO_STDLIB);
    }

    /**
     * Get a {@link GoDependency} representing the standard library package import path with the given alias.
     *
     * @param importPath standard library package import path
     * @param alias      the package alias
     * @return the {@link GoDependency} for the package import path
     */
    public static GoDependency stdlib(String importPath, String alias) {
        return GoDependency.standardLibraryDependency(importPath, Versions.GO_STDLIB, alias);
    }

    private static GoDependency smithy(String relativePath) {
        return smithy(relativePath, null);
    }

    private static GoDependency smithy(String relativePath, String alias) {
        return relativePackage(SMITHY_SOURCE_PATH, relativePath, Versions.SMITHY_GO, alias);
    }

    private static GoDependency goCmp(String relativePath) {
        return relativePackage(GO_CMP_SOURCE_PATH, relativePath, Versions.GO_CMP, null);
    }

    private static GoDependency goJmespath(String relativePath) {
        return relativePackage(GO_JMESPATH_SOURCE_PATH, relativePath, Versions.GO_JMESPATH, null);
    }

    private static GoDependency relativePackage(
            String moduleImportPath,
            String relativePath,
            String version,
            String alias
    ) {
        String importPath = moduleImportPath;
        if (relativePath != null) {
            importPath = importPath + "/" + relativePath;
        }
        return GoDependency.moduleDependency(moduleImportPath, importPath, version, alias);
    }

    private static final class Versions {
        private static final String GO_STDLIB = "1.15";
        private static final String GO_CMP = "v0.5.4";
        private static final String SMITHY_GO = "v1.4.0";
        private static final String GO_JMESPATH = "v0.4.0";
    }
}
