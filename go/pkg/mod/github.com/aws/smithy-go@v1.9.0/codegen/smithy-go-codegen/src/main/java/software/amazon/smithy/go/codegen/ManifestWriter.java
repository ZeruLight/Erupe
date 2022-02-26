/*
 * Copyright 2021 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.TreeMap;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import software.amazon.smithy.build.FileManifest;
import software.amazon.smithy.codegen.core.CodegenException;
import software.amazon.smithy.codegen.core.SymbolDependency;
import software.amazon.smithy.model.Model;
import software.amazon.smithy.model.node.ArrayNode;
import software.amazon.smithy.model.node.BooleanNode;
import software.amazon.smithy.model.node.Node;
import software.amazon.smithy.model.node.ObjectNode;
import software.amazon.smithy.model.node.StringNode;
import software.amazon.smithy.model.traits.UnstableTrait;

/**
 * Generates a manifest description of the generated code, minimum go version, and minimum dependencies required.
 */
public final class ManifestWriter {
    private static final String GENERATED_JSON = "generated.json";

    private ManifestWriter() {
    }

    /**
     * Write the manifest description of the generated code.
     *
     * @param settings     the go settings
     * @param model        the smithy model
     * @param fileManifest the file manifest
     * @param dependencies the list of symbol dependencies
     */
    public static void writeManifest(
            GoSettings settings,
            Model model,
            FileManifest fileManifest,
            List<SymbolDependency> dependencies
    ) {
        Path manifestFile = fileManifest.getBaseDir().resolve(GENERATED_JSON);

        if (Files.exists(manifestFile)) {
            try {
                Files.delete(manifestFile);
            } catch (IOException e) {
                throw new CodegenException("Failed to delete existing " + GENERATED_JSON + " file", e);
            }
        }
        fileManifest.addFile(manifestFile);

        Node generatedJson = buildManifestFile(settings, model, fileManifest, dependencies);
        fileManifest.writeFile(manifestFile.toString(), Node.prettyPrintJson(generatedJson) + "\n");
    }

    private static Node buildManifestFile(
            GoSettings settings,
            Model model,
            FileManifest fileManifest,
            List<SymbolDependency> dependencies
    ) {

        List<SymbolDependency> nonStdLib = new ArrayList<>();
        Optional<SymbolDependency> minStandard = Optional.empty();

        for (SymbolDependency dependency : dependencies) {
            if (!dependency.getDependencyType().equals(GoDependency.Type.STANDARD_LIBRARY.toString())) {
                nonStdLib.add(dependency);
            } else {
                if (minStandard.isPresent()) {
                    if (minStandard.get().getVersion().compareTo(dependency.getVersion()) < 0) {
                        minStandard = Optional.of(dependency);
                    }
                } else {
                    minStandard = Optional.of(dependency);
                }
            }
        }

        Map<StringNode, Node> manifestNodes = new HashMap<>();

        Map<String, String> minimumDependencies = gatherMinimumDependencies(nonStdLib.stream());

        Map<StringNode, Node> dependencyNodes = new HashMap<>();
        for (Map.Entry<String, String> entry : minimumDependencies.entrySet()) {
            dependencyNodes.put(StringNode.from(entry.getKey()),
                    StringNode.from(entry.getValue()));
        }

        Collection<String> generatedFiles = new ArrayList<>();
        Path baseDir = fileManifest.getBaseDir();
        for (Path filePath : fileManifest.getFiles()) {
            generatedFiles.add(baseDir.relativize(filePath).toString());
        }
        generatedFiles = generatedFiles.stream().sorted().collect(Collectors.toList());

        manifestNodes.put(StringNode.from("module"), StringNode.from(settings.getModuleName()));
        minStandard.ifPresent(symbolDependency ->
                manifestNodes.put(StringNode.from("go"), StringNode.from(symbolDependency.getVersion())));
        manifestNodes.put(StringNode.from("dependencies"), ObjectNode.objectNode(dependencyNodes));
        manifestNodes.put(StringNode.from("files"), ArrayNode.fromStrings(generatedFiles));
        manifestNodes.put(StringNode.from("unstable"),
                BooleanNode.from(settings.getService(model).getTrait(UnstableTrait.class).isPresent()));

        return ObjectNode.objectNode(manifestNodes).withDeepSortedKeys();
    }

    private static Map<String, String> gatherMinimumDependencies(
            Stream<SymbolDependency> symbolStream
    ) {
        return SymbolDependency.gatherDependencies(symbolStream, GoDependency::mergeByMinimumVersionSelection)
                .entrySet().stream()
                .flatMap(entry -> entry.getValue().entrySet().stream())
                .collect(Collectors.toMap(
                        Map.Entry::getKey, entry -> entry.getValue().getVersion(), (a, b) -> b, TreeMap::new));
    }

}
