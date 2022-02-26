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

import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;
import java.util.Optional;
import java.util.StringJoiner;
import java.util.function.BiFunction;
import java.util.logging.Logger;
import software.amazon.smithy.codegen.core.CodegenException;
import software.amazon.smithy.codegen.core.Symbol;
import software.amazon.smithy.codegen.core.SymbolContainer;
import software.amazon.smithy.codegen.core.SymbolDependency;
import software.amazon.smithy.codegen.core.SymbolDependencyContainer;
import software.amazon.smithy.codegen.core.SymbolReference;
import software.amazon.smithy.go.codegen.knowledge.GoUsageIndex;
import software.amazon.smithy.model.Model;
import software.amazon.smithy.model.shapes.MemberShape;
import software.amazon.smithy.model.shapes.Shape;
import software.amazon.smithy.model.traits.DeprecatedTrait;
import software.amazon.smithy.model.traits.DocumentationTrait;
import software.amazon.smithy.model.traits.HttpPrefixHeadersTrait;
import software.amazon.smithy.model.traits.MediaTypeTrait;
import software.amazon.smithy.model.traits.RequiredTrait;
import software.amazon.smithy.model.traits.StringTrait;
import software.amazon.smithy.utils.CodeWriter;
import software.amazon.smithy.utils.StringUtils;

/**
 * Specialized code writer for managing Go dependencies.
 *
 * <p>Use the {@code $T} formatter to refer to {@link Symbol}s without using pointers.
 *
 * <p>Use the {@code $P} formatter to refer to {@link Symbol}s using pointers where appropriate.
 */
public final class GoWriter extends CodeWriter {

    private static final Logger LOGGER = Logger.getLogger(GoWriter.class.getName());
    private static final int DEFAULT_DOC_WRAP_LENGTH = 80;

    private final String fullPackageName;
    private final ImportDeclarations imports = new ImportDeclarations();
    private final List<SymbolDependency> dependencies = new ArrayList<>();

    private int docWrapLength = DEFAULT_DOC_WRAP_LENGTH;

    private CodeWriter packageDocs;

    public GoWriter(String fullPackageName) {
        this.fullPackageName = fullPackageName;
        trimBlankLines();
        trimTrailingSpaces();
        setIndentText("\t");
        putFormatter('T', new GoSymbolFormatter());
        putFormatter('P', new PointableGoSymbolFormatter());

        packageDocs = new CodeWriter();
        packageDocs.trimBlankLines();
        packageDocs.trimTrailingSpaces();
        packageDocs.setIndentText("\t");
    }

    /**
     * Sets the wrap length of doc strings written.
     *
     * @param wrapLength The wrap length of the doc string.
     * @return Returns the writer.
     */
    public GoWriter setDocWrapLength(final int wrapLength) {
        this.docWrapLength = wrapLength;
        return this;
    }

    /**
     * Sets the wrap length of doc strings written to the default value for the Go writer.
     *
     * @return Returns the writer.
     */
    public GoWriter setDocWrapLength() {
        this.docWrapLength = DEFAULT_DOC_WRAP_LENGTH;
        return this;
    }

    /**
     * Imports one or more symbols if necessary, using the name of the
     * symbol and only "USE" references.
     *
     * @param container Container of symbols to add.
     * @return Returns the writer.
     */
    public GoWriter addUseImports(SymbolContainer container) {
        for (Symbol symbol : container.getSymbols()) {
            addImport(symbol,
                    CodegenUtils.getSymbolNamespaceAlias(symbol),
                    SymbolReference.ContextOption.USE);
        }
        return this;
    }

    /**
     * Imports a symbol reference if necessary, using the alias of the
     * reference and only associated "USE" references.
     *
     * @param symbolReference Symbol reference to import.
     * @return Returns the writer.
     */
    public GoWriter addUseImports(SymbolReference symbolReference) {
        return addImport(symbolReference.getSymbol(), symbolReference.getAlias(), SymbolReference.ContextOption.USE);
    }

    /**
     * Adds and imports the given dependency.
     *
     * @param goDependency The GoDependency to import.
     * @return Returns the writer.
     */
    public GoWriter addUseImports(GoDependency goDependency) {
        dependencies.addAll(goDependency.getDependencies());
        return addImport(goDependency.getImportPath(), goDependency.getAlias());
    }

    /**
     * Imports a symbol if necessary using a package alias and list of context options.
     *
     * @param symbol       Symbol to optionally import.
     * @param packageAlias The alias to refer to the symbol's package by.
     * @param options      The list of context options (e.g., is it a USE or DECLARE symbol).
     * @return Returns the writer.
     */
    public GoWriter addImport(Symbol symbol, String packageAlias, SymbolReference.ContextOption... options) {
        LOGGER.finest(() -> {
            StringJoiner stackTrace = new StringJoiner("\n");
            for (StackTraceElement element : Thread.currentThread().getStackTrace()) {
                stackTrace.add(element.toString());
            }
            return String.format(
                    "Adding Go import %s as `%s` (%s); Stack trace: %s",
                    symbol.getNamespace(), packageAlias, Arrays.toString(options), stackTrace);
        });

        // Always add dependencies.
        dependencies.addAll(symbol.getDependencies());

        if (isExternalNamespace(symbol.getNamespace())) {
            addImport(symbol.getNamespace(), packageAlias);
        }

        // Just because the direct symbol wasn't imported doesn't mean that the
        // symbols it needs to be declared don't need to be imported.
        addImportReferences(symbol, options);

        return this;
    }

    private boolean isExternalNamespace(String namespace) {
        return !StringUtils.isBlank(namespace) && !namespace.equals(fullPackageName);
    }

    void addImportReferences(Symbol symbol, SymbolReference.ContextOption... options) {
        for (SymbolReference reference : symbol.getReferences()) {
            for (SymbolReference.ContextOption option : options) {
                if (reference.hasOption(option)) {
                    addImport(reference.getSymbol(), reference.getAlias(), options);
                    break;
                }
            }
        }
    }

    /**
     * Imports a package using an alias if necessary.
     *
     * @param packageName Package to import.
     * @param as          Alias to refer to the package as.
     * @return Returns the writer.
     */
    public GoWriter addImport(String packageName, String as) {
        imports.addImport(packageName, as);
        return this;
    }

    /**
     * Adds one or more dependencies to the generated code.
     *
     * <p>The dependencies of all writers created by the {@link GoDelegator}
     * are merged together to eventually generate a go.mod file.
     *
     * @param dependencies Go dependency to add.
     * @return Returns the writer.
     */
    public GoWriter addDependency(SymbolDependencyContainer dependencies) {
        this.dependencies.addAll(dependencies.getDependencies());
        return this;
    }

    Collection<SymbolDependency> getDependencies() {
        return dependencies;
    }

    /**
     * Writes documentation comments.
     *
     * @param runnable Runnable that handles actually writing docs with the writer.
     * @return Returns the writer.
     */
    private void writeDocs(CodeWriter writer, Runnable runnable) {
        writer.pushState("docs");
        writer.setNewlinePrefix("// ");
        runnable.run();
        writer.setNewlinePrefix("");
        writer.popState();
    }

    private void writeDocs(CodeWriter writer, int docWrapLength, String docs) {
        String wrappedDoc = StringUtils.wrap(DocumentationConverter.convert(docs), docWrapLength);
        writeDocs(writer, () -> writer.write(wrappedDoc.replace("$", "$$")));
    }

    /**
     * Writes documentation comments from a string.
     *
     * <p>This function escapes "$" characters so formatters are not run.
     *
     * @param docs Documentation to write.
     * @return Returns the writer.
     */
    public GoWriter writeDocs(String docs) {
        writeDocs(this, docWrapLength, docs);
        return this;
    }

    /**
     * Writes the doc to the Go package docs that are written prior to the go package statement.
     *
     * @param docs documentation to write to package doc.
     * @return writer
     */
    public GoWriter writePackageDocs(String docs) {
        writeDocs(packageDocs, docWrapLength, docs);
        return this;
    }

    /**
     * Writes the doc to the Go package docs that are written prior to the go package statement. This does not perform
     * line wrapping and the provided formatting must be valid Go doc.
     *
     * @param docs documentation to write to package doc.
     * @return writer
     */
    public GoWriter writeRawPackageDocs(String docs) {
        writeDocs(packageDocs, () -> {
            packageDocs.write(docs);
        });
        return this;
    }

    /**
     * Writes shape documentation comments if docs are present.
     *
     * @param shape Shape to write the documentation of.
     * @return Returns true if docs were written.
     */
    boolean writeShapeDocs(Shape shape) {
        return shape.getTrait(DocumentationTrait.class)
                .map(DocumentationTrait::getValue)
                .map(docs -> {
                    writeDocs(docs);
                    return true;
                }).orElse(false);
    }

    /**
     * Writes shape documentation comments to the writer's package doc if docs are present.
     *
     * @param shape Shape to write the documentation of.
     * @return Returns true if docs were written.
     */
    boolean writePackageShapeDocs(Shape shape) {
        return shape.getTrait(DocumentationTrait.class)
                .map(DocumentationTrait::getValue)
                .map(docs -> {
                    writePackageDocs(docs);
                    return true;
                }).orElse(false);
    }

    /**
     * Writes member shape documentation comments if docs are present.
     *
     * @param model  Model used to dereference targets.
     * @param member Shape to write the documentation of.
     * @return Returns true if docs were written.
     */
    boolean writeMemberDocs(Model model, MemberShape member) {
        boolean hasDocs;

        hasDocs = member.getMemberTrait(model, DocumentationTrait.class)
                .map(DocumentationTrait::getValue)
                .map(docs -> {
                    writeDocs(docs);
                    return true;
                }).orElse(false);

        Optional<String> stringOptional = member.getMemberTrait(model, MediaTypeTrait.class)
                .map(StringTrait::getValue);
        if (stringOptional.isPresent()) {
            if (hasDocs) {
                writeDocs("");
            }
            writeDocs("This value conforms to the media type: " + stringOptional.get());
            hasDocs = true;
        }

        GoUsageIndex usageIndex = GoUsageIndex.of(model);
        if (usageIndex.isUsedForOutput(member)) {
            if (member.getMemberTrait(model,
                    HttpPrefixHeadersTrait.class).isPresent()) {
                if (hasDocs) {
                    writeDocs("");
                }
                writeDocs("Map keys will be normalized to lower-case.");
                hasDocs = true;
            }
        }

        if (member.getMemberTrait(model, RequiredTrait.class).isPresent()) {
            if (hasDocs) {
                writeDocs("");
            }
            writeDocs("This member is required.");
            hasDocs = true;
        }

        Optional<DeprecatedTrait> deprecatedTrait = member.getMemberTrait(model, DeprecatedTrait.class);
        if (deprecatedTrait.isPresent()) {
            if (hasDocs) {
                writeDocs("");
            }
            final String defaultMessage = "This member has been deprecated.";
            writeDocs("Deprecated: " + deprecatedTrait.get().getMessage().map(s -> {
                if (s.length() == 0) {
                    return defaultMessage;
                }
                return s;
            }).orElse(defaultMessage));
        }

        return hasDocs;
    }

    @Override
    public String toString() {
        String contents = super.toString();
        String[] packageParts = fullPackageName.split("/");
        String header = String.format("// Code generated by smithy-go-codegen DO NOT EDIT.%n%n");

        String packageName = packageParts[packageParts.length - 1];
        if (packageName.startsWith("v") && packageParts.length >= 2) {
            String remaining = packageName.substring(1);
            try {
                int value = Integer.parseInt(remaining);
                packageName = packageParts[packageParts.length - 2];
                if (value == 0 || value == 1) {
                    throw new CodegenException("module paths vN version component must only be N >= 2");
                }
            } catch (NumberFormatException ne) {
                // Do nothing
            }
        }

        String packageDocs = this.packageDocs.toString();
        String packageStatement = String.format("package %s%n%n", packageName);

        String importString = imports.toString();
        String strippedContents = StringUtils.stripStart(contents, null);
        String strippedImportString = StringUtils.strip(importString, null);

        // Don't add an additional new line between explicit imports and managed imports.
        if (!strippedImportString.isEmpty() && strippedContents.startsWith("import ")) {
            return header + strippedImportString + "\n" + strippedContents;
        }

        return header + packageDocs + packageStatement + importString + contents;
    }

    /**
     * Implements Go symbol formatting for the {@code $T} formatter.
     */
    private class GoSymbolFormatter implements BiFunction<Object, String, String> {
        @Override
        public String apply(Object type, String indent) {
            if (type instanceof Symbol) {
                Symbol resolvedSymbol = (Symbol) type;
                String literal = resolvedSymbol.getName();

                boolean isSlice = resolvedSymbol.getProperty(SymbolUtils.GO_SLICE, Boolean.class).orElse(false);
                boolean isMap = resolvedSymbol.getProperty(SymbolUtils.GO_MAP, Boolean.class).orElse(false);
                if (isSlice || isMap) {
                    resolvedSymbol = resolvedSymbol.getProperty(SymbolUtils.GO_ELEMENT_TYPE, Symbol.class)
                            .orElseThrow(() -> new CodegenException("Expected go element type property to be defined"));
                    literal = new PointableGoSymbolFormatter().apply(resolvedSymbol, "nested");
                } else if (!SymbolUtils.isUniverseType(resolvedSymbol)
                        && isExternalNamespace(resolvedSymbol.getNamespace())) {
                    literal = formatWithNamespace(resolvedSymbol);
                }
                addUseImports(resolvedSymbol);

                if (isSlice) {
                    return "[]" + literal;
                } else if (isMap) {
                    return "map[string]" + literal;
                } else {
                    return literal;
                }
            } else if (type instanceof SymbolReference) {
                SymbolReference typeSymbol = (SymbolReference) type;
                addImport(typeSymbol.getSymbol(), typeSymbol.getAlias(), SymbolReference.ContextOption.USE);
                return typeSymbol.getAlias();
            } else {
                throw new CodegenException(
                        "Invalid type provided to $T. Expected a Symbol, but found `" + type + "`");
            }
        }

        private String formatWithNamespace(Symbol symbol) {
            if (StringUtils.isEmpty(symbol.getNamespace())) {
                return symbol.getName();
            }
            return String.format("%s.%s", CodegenUtils.getSymbolNamespaceAlias(symbol), symbol.getName());
        }
    }

    /**
     * Implements Go symbol formatting for the {@code $P} formatter. This is identical to the $T
     * formatter, except that it will add a * to symbols that can be pointers.
     */
    private class PointableGoSymbolFormatter extends GoSymbolFormatter {
        @Override
        public String apply(Object type, String indent) {
            String formatted = super.apply(type, indent);
            if (isPointer(type)) {
                formatted = "*" + formatted;
            }
            return formatted;
        }

        private boolean isPointer(Object type) {
            if (type instanceof Symbol) {
                Symbol typeSymbol = (Symbol) type;
                return typeSymbol.getProperty(SymbolUtils.POINTABLE, Boolean.class).orElse(false);
            } else if (type instanceof SymbolReference) {
                SymbolReference typeSymbol = (SymbolReference) type;
                return typeSymbol.getProperty(SymbolUtils.POINTABLE, Boolean.class).orElse(false)
                        || typeSymbol.getSymbol().getProperty(SymbolUtils.POINTABLE, Boolean.class).orElse(false);
            } else {
                throw new CodegenException(
                        "Invalid type provided to $P. Expected a Symbol, but found `" + type + "`");
            }
        }
    }
}
