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

import java.util.Map;
import java.util.TreeMap;
import software.amazon.smithy.codegen.core.CodegenException;
import software.amazon.smithy.utils.StringUtils;

/**
 * Container and formatter for go imports.
 */
final class ImportDeclarations {

    private final Map<String, String> imports = new TreeMap<>();

    ImportDeclarations addImport(String packageName, String alias) {
        String importedName = CodegenUtils.getDefaultPackageImportName(packageName);
        if (!StringUtils.isBlank(alias)) {
            if (alias.equals(".")) {
                // Global imports are generally a bad practice.
                throw new CodegenException("Globally importing packages is forbidden: " + packageName);
            }
            importedName = alias;
        }
        imports.putIfAbsent(importedName, packageName);
        return this;
    }

    @Override
    public String toString() {
        if (imports.isEmpty()) {
            return "";
        }

        StringBuilder builder = new StringBuilder("import (\n");
        for (Map.Entry<String, String> entry : imports.entrySet()) {
            builder.append('\t');
            builder.append(createImportStatement(entry));
            builder.append('\n');
        }
        builder.append(")\n\n");
        return builder.toString();
    }

    private String createImportStatement(Map.Entry<String, String> entry) {
        String formattedPackageName = "\"" + entry.getValue() + "\"";
        return CodegenUtils.getDefaultPackageImportName(entry.getValue()).equals(entry.getKey())
                ? formattedPackageName
                : entry.getKey() + " " + formattedPackageName;
    }
}
