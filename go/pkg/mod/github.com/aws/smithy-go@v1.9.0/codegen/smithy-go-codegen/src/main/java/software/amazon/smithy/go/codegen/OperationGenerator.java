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
import java.util.Collection;
import java.util.List;
import java.util.stream.Stream;
import software.amazon.smithy.codegen.core.CodegenException;
import software.amazon.smithy.codegen.core.Symbol;
import software.amazon.smithy.codegen.core.SymbolProvider;
import software.amazon.smithy.go.codegen.integration.MiddlewareRegistrar;
import software.amazon.smithy.go.codegen.integration.ProtocolGenerator;
import software.amazon.smithy.go.codegen.integration.RuntimeClientPlugin;
import software.amazon.smithy.model.Model;
import software.amazon.smithy.model.knowledge.OperationIndex;
import software.amazon.smithy.model.shapes.OperationShape;
import software.amazon.smithy.model.shapes.ServiceShape;
import software.amazon.smithy.model.shapes.StructureShape;
import software.amazon.smithy.model.traits.DeprecatedTrait;
import software.amazon.smithy.model.traits.StreamingTrait;

/**
 * Generates a client operation and associated custom shapes.
 */
public final class OperationGenerator implements Runnable {

    private final GoSettings settings;
    private final Model model;
    private final SymbolProvider symbolProvider;
    private final GoWriter writer;
    private final ServiceShape service;
    private final OperationShape operation;
    private final Symbol operationSymbol;
    private final ApplicationProtocol applicationProtocol;
    private final ProtocolGenerator protocolGenerator;
    private final List<RuntimeClientPlugin> runtimeClientPlugins;

    OperationGenerator(
            GoSettings settings,
            Model model,
            SymbolProvider symbolProvider,
            GoWriter writer,
            ServiceShape service,
            OperationShape operation,
            Symbol operationSymbol,
            ApplicationProtocol applicationProtocol,
            ProtocolGenerator protocolGenerator,
            List<RuntimeClientPlugin> runtimeClientPlugins
    ) {
        this.settings = settings;
        this.model = model;
        this.symbolProvider = symbolProvider;
        this.writer = writer;
        this.service = service;
        this.operation = operation;
        this.operationSymbol = operationSymbol;
        this.applicationProtocol = applicationProtocol;
        this.protocolGenerator = protocolGenerator;
        this.runtimeClientPlugins = runtimeClientPlugins;
    }

    @Override
    public void run() {
        OperationIndex operationIndex = OperationIndex.of(model);
        Symbol serviceSymbol = symbolProvider.toSymbol(service);

        if (!operationIndex.getInput(operation).isPresent()) {
            // Theoretically this shouldn't ever get hit since we automatically insert synthetic inputs / outputs.
            throw new CodegenException(
                    "Operations are required to have input shapes in order to allow for future evolution.");
        }
        StructureShape inputShape = operationIndex.getInput(operation).get();
        Symbol inputSymbol = symbolProvider.toSymbol(inputShape);

        if (!operationIndex.getOutput(operation).isPresent()) {
            throw new CodegenException(
                    "Operations are required to have output shapes in order to allow for future evolution.");
        }
        StructureShape outputShape = operationIndex.getOutput(operation).get();
        Symbol outputSymbol = symbolProvider.toSymbol(outputShape);

        // Generate operation method
        final boolean hasDocs = writer.writeShapeDocs(operation);
        operation.getTrait(DeprecatedTrait.class)
                .ifPresent(trait -> {
                    if (hasDocs) {
                        writer.writeDocs("");
                    }
                    final String defaultMessage = "This operation has been deprecated.";
                    writer.writeDocs("Deprecated: " + trait.getMessage().map(s -> {
                        if (s.length() == 0) {
                            return defaultMessage;
                        }
                        return s;
                    }).orElse(defaultMessage));
                });
        Symbol contextSymbol = SymbolUtils.createValueSymbolBuilder("Context", SmithyGoDependency.CONTEXT).build();
        writer.openBlock("func (c $P) $T(ctx $T, params $P, optFns ...func(*Options)) ($P, error) {", "}",
                serviceSymbol, operationSymbol, contextSymbol, inputSymbol, outputSymbol, () -> {
                    writer.write("if params == nil { params = &$T{} }", inputSymbol);
                    writer.write("");

                    writer.write("result, metadata, err := c.invokeOperation(ctx, $S, params, optFns, c.$L)",
                            operationSymbol.getName(), getAddOperationMiddlewareFuncName(operationSymbol));
                    writer.write("if err != nil { return nil, err }");
                    writer.write("");

                    writer.write("out := result.($P)", outputSymbol);
                    writer.write("out.ResultMetadata = metadata");
                    writer.write("return out, nil");
                }).write("");

        // Write out the input and output structures. These are written out here to prevent naming conflicts with other
        // shapes in the model.
        new StructureGenerator(model, symbolProvider, writer, service, inputShape, inputSymbol, protocolGenerator)
                .renderStructure(() -> {
                }, true);

        // The output structure gets a metadata member added.
        Symbol metadataSymbol = SymbolUtils.createValueSymbolBuilder("Metadata", SmithyGoDependency.SMITHY_MIDDLEWARE)
                .build();

        boolean hasEventStream = Stream.concat(inputShape.members().stream(),
                        outputShape.members().stream())
                .anyMatch(memberShape -> StreamingTrait.isEventStream(model, memberShape));

        new StructureGenerator(model, symbolProvider, writer, service, outputShape, outputSymbol, protocolGenerator)
                .renderStructure(() -> {
                    if (outputShape.getMemberNames().size() != 0) {
                        writer.write("");
                    }

                    if (hasEventStream) {
                        writer.write("eventStream $P",
                                        EventStreamGenerator.getEventStreamOperationStructureSymbol(service, operation))
                                .write("");
                    }

                    writer.writeDocs("Metadata pertaining to the operation's result.");
                    writer.write("ResultMetadata $T", metadataSymbol);
                });

        if (hasEventStream) {
            writer.write("""
                         // GetStream returns the type to interact with the event stream.
                         func (o $P) GetStream() $P {
                             return o.eventStream
                         }
                         """, outputSymbol, EventStreamGenerator.getEventStreamOperationStructureSymbol(
                    service, operation));
        }

        // Generate operation protocol middleware helper function
        generateAddOperationMiddleware();
    }

    /**
     * Adds middleware to the operation middleware stack.
     */
    private void generateAddOperationMiddleware() {
        Symbol stackSymbol = SymbolUtils.createPointableSymbolBuilder("Stack", SmithyGoDependency.SMITHY_MIDDLEWARE)
                .build();

        writer.openBlock("func (c *Client) $L(stack $P, options Options) (err error) {", "}",
                getAddOperationMiddlewareFuncName(operationSymbol), stackSymbol,
                () -> {
                    generateOperationProtocolMiddlewareAdders();

                    // Populate middleware's from runtime client plugins
                    runtimeClientPlugins.forEach(runtimeClientPlugin -> {
                        if (!runtimeClientPlugin.matchesService(model, service)
                            && !runtimeClientPlugin.matchesOperation(model, service, operation)) {
                            return;
                        }

                        if (!runtimeClientPlugin.registerMiddleware().isPresent()) {
                            return;
                        }

                        MiddlewareRegistrar middlewareRegistrar = runtimeClientPlugin.registerMiddleware().get();
                        Collection<Symbol> functionArguments = middlewareRegistrar.getFunctionArguments();

                        // TODO these functions do not all return err like they should. This should be fixed.
                        // TODO Must be fixed for all public functions.
                        if (middlewareRegistrar.getInlineRegisterMiddlewareStatement() != null) {
                            String registerStatement = String.format("if err = stack.%s",
                                    middlewareRegistrar.getInlineRegisterMiddlewareStatement());
                            writer.writeInline(registerStatement);
                            writer.writeInline("$T(", middlewareRegistrar.getResolvedFunction());
                            if (functionArguments != null) {
                                List<Symbol> args = new ArrayList<>(functionArguments);
                                for (Symbol arg : args) {
                                    writer.writeInline("$P, ", arg);
                                }
                            }
                            writer.writeInline(")");
                            writer.write(", $T); err != nil {\nreturn err\n}",
                                    middlewareRegistrar.getInlineRegisterMiddlewarePosition());
                        } else {
                            writer.writeInline("if err = $T(stack", middlewareRegistrar.getResolvedFunction());
                            if (functionArguments != null) {
                                List<Symbol> args = new ArrayList<>(functionArguments);
                                for (Symbol arg : args) {
                                    writer.writeInline(", $P", arg);
                                }
                            }
                            writer.write("); err != nil {\nreturn err\n}");
                        }
                    });

                    writer.write("return nil");
                });
    }

    /**
     * Generate operation protocol middleware helper.
     */
    private void generateOperationProtocolMiddlewareAdders() {
        if (protocolGenerator == null) {
            return;
        }
        writer.addUseImports(SmithyGoDependency.SMITHY_MIDDLEWARE);

        // Add request serializer middleware
        String serializerMiddlewareName = ProtocolGenerator.getSerializeMiddlewareName(
                operation.getId(), service, protocolGenerator.getProtocolName());
        writer.write("err = stack.Serialize.Add(&$L{}, middleware.After)", serializerMiddlewareName);
        writer.write("if err != nil { return err }");

        // Adds response deserializer middleware
        String deserializerMiddlewareName = ProtocolGenerator.getDeserializeMiddlewareName(
                operation.getId(), service, protocolGenerator.getProtocolName());
        writer.write("err = stack.Deserialize.Add(&$L{}, middleware.After)", deserializerMiddlewareName);
        writer.write("if err != nil { return err }");
    }

    /**
     * Returns the name of the operation's middleware mutator function, that adds all middleware for the operation to
     * the stack.
     *
     * @param operation symbol for operation
     * @return name of function
     */
    public static String getAddOperationMiddlewareFuncName(Symbol operation) {
        return String.format("addOperation%sMiddlewares", operation.getName());
    }
}
