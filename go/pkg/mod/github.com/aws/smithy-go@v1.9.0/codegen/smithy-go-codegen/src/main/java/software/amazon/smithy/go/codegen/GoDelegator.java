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

import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.function.Consumer;
import software.amazon.smithy.build.FileManifest;
import software.amazon.smithy.codegen.core.Symbol;
import software.amazon.smithy.codegen.core.SymbolDependency;
import software.amazon.smithy.codegen.core.SymbolProvider;
import software.amazon.smithy.codegen.core.SymbolReference;
import software.amazon.smithy.model.Model;
import software.amazon.smithy.model.shapes.Shape;

/**
 * Manages writers for Go files.
 */
public final class GoDelegator {

    private final GoSettings settings;
    private final Model model;
    private final FileManifest fileManifest;
    private final SymbolProvider symbolProvider;
    private final Map<String, GoWriter> writers = new HashMap<>();

    GoDelegator(GoSettings settings, Model model, FileManifest fileManifest, SymbolProvider symbolProvider) {
        this.settings = settings;
        this.model = model;
        this.fileManifest = fileManifest;
        this.symbolProvider = symbolProvider;
    }

    /**
     * Writes all pending writers to disk and then clears them out.
     */
    void flushWriters() {
        writers.forEach((filename, writer) -> fileManifest.writeFile(filename, writer.toString()));
        writers.clear();
    }

    /**
     * Gets all of the dependencies that have been registered in writers owned by the
     * delegator.
     *
     * @return Returns all the dependencies.
     */
    List<SymbolDependency> getDependencies() {
        List<SymbolDependency> resolved = new ArrayList<>();
        writers.values().forEach(s -> resolved.addAll(s.getDependencies()));
        return resolved;
    }

    /**
     * Gets a previously created writer or creates a new one if needed.
     *
     * @param shape          Shape to create the writer for.
     * @param writerConsumer Consumer that accepts and works with the file.
     */
    public void useShapeWriter(Shape shape, Consumer<GoWriter> writerConsumer) {
        Symbol symbol = symbolProvider.toSymbol(shape);
        useShapeWriter(symbol, writerConsumer);
    }

    /**
     * Gets a previously created writer or creates a new one for the a Go test file for the associated shape.
     *
     * @param shape          Shape to create the writer for.
     * @param writerConsumer Consumer that accepts and works with the file.
     */
    public void useShapeTestWriter(Shape shape, Consumer<GoWriter> writerConsumer) {
        Symbol symbol = symbolProvider.toSymbol(shape);
        String filename = symbol.getDefinitionFile();

        StringBuilder b = new StringBuilder(filename);
        b.insert(filename.lastIndexOf(".go"), "_test");
        filename = b.toString();

        symbol = symbol.toBuilder()
                .definitionFile(filename)
                .build();

        useShapeWriter(symbol, writerConsumer);
    }

    /**
     * Gets a previously created writer or creates a new one for the a Go public package test file for the associated
     * shape.
     *
     * @param shape          Shape to create the writer for.
     * @param writerConsumer Consumer that accepts and works with the file.
     */
    public void useShapeExportedTestWriter(Shape shape, Consumer<GoWriter> writerConsumer) {
        Symbol symbol = symbolProvider.toSymbol(shape);
        String filename = symbol.getDefinitionFile();

        StringBuilder b = new StringBuilder(filename);
        b.insert(filename.lastIndexOf(".go"), "_exported_test");
        filename = b.toString();

        symbol = symbol.toBuilder()
                .definitionFile(filename)
                .namespace(symbol.getNamespace() + "_test", symbol.getNamespaceDelimiter())
                .build();

        useShapeWriter(symbol, writerConsumer);
    }

    /**
     * Gets a previously created writer or creates a new one if needed.
     *
     * @param symbol         symbol to create the writer for.
     * @param writerConsumer Consumer that accepts and works with the file.
     */
    private void useShapeWriter(Symbol symbol, Consumer<GoWriter> writerConsumer) {
        GoWriter writer = checkoutWriter(symbol.getDefinitionFile(), symbol.getNamespace());

        // Add any needed DECLARE symbols.
        writer.addImportReferences(symbol, SymbolReference.ContextOption.DECLARE);
        symbol.getDependencies().forEach(writer::addDependency);

        writer.pushState();
        writerConsumer.accept(writer);
        writer.popState();
    }

    /**
     * Gets a previously created writer or creates a new one if needed
     * and adds a new line if the writer already exists.
     *
     * @param filename       Name of the file to create.
     * @param writerConsumer Consumer that accepts and works with the file.
     */
    void useFileWriter(String filename, String namespace, Consumer<GoWriter> writerConsumer) {
        writerConsumer.accept(checkoutWriter(filename, namespace));
    }

    private GoWriter checkoutWriter(String filename, String namespace) {
        String formattedFilename = Paths.get(filename).normalize().toString();
        boolean needsNewline = writers.containsKey(formattedFilename);

        GoWriter writer = writers.computeIfAbsent(formattedFilename, f -> new GoWriter(namespace));

        if (needsNewline) {
            writer.write("\n");
        }

        return writer;
    }
}
