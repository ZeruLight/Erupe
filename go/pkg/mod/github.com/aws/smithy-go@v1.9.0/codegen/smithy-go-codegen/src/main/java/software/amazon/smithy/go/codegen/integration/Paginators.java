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

package software.amazon.smithy.go.codegen.integration;

import java.util.List;
import java.util.Optional;
import java.util.function.Consumer;
import software.amazon.smithy.codegen.core.Symbol;
import software.amazon.smithy.codegen.core.SymbolProvider;
import software.amazon.smithy.go.codegen.CodegenUtils;
import software.amazon.smithy.go.codegen.GoDelegator;
import software.amazon.smithy.go.codegen.GoSettings;
import software.amazon.smithy.go.codegen.GoValueAccessUtils;
import software.amazon.smithy.go.codegen.GoWriter;
import software.amazon.smithy.go.codegen.SmithyGoDependency;
import software.amazon.smithy.go.codegen.SymbolUtils;
import software.amazon.smithy.go.codegen.knowledge.GoPointableIndex;
import software.amazon.smithy.go.codegen.trait.PagingExtensionTrait;
import software.amazon.smithy.model.Model;
import software.amazon.smithy.model.knowledge.PaginatedIndex;
import software.amazon.smithy.model.knowledge.PaginationInfo;
import software.amazon.smithy.model.knowledge.TopDownIndex;
import software.amazon.smithy.model.shapes.BooleanShape;
import software.amazon.smithy.model.shapes.MemberShape;
import software.amazon.smithy.model.shapes.OperationShape;
import software.amazon.smithy.model.shapes.ServiceShape;
import software.amazon.smithy.model.traits.DocumentationTrait;

/**
 * Implements support for PaginatedTrait.
 */
public class Paginators implements GoIntegration {
    @Override
    public void writeAdditionalFiles(
            GoSettings settings,
            Model model,
            SymbolProvider symbolProvider,
            GoDelegator goDelegator
    ) {
        ServiceShape serviceShape = settings.getService(model);

        PaginatedIndex paginatedIndex = PaginatedIndex.of(model);

        TopDownIndex topDownIndex = TopDownIndex.of(model);

        topDownIndex.getContainedOperations(serviceShape).stream()
                .map(operationShape -> paginatedIndex.getPaginationInfo(serviceShape, operationShape))
                .filter(Optional::isPresent)
                .map(Optional::get)
                .forEach(paginationInfo -> {
                    goDelegator.useShapeWriter(paginationInfo.getOperation(), writer -> {
                        generateOperationPaginator(model, symbolProvider, writer, paginationInfo);
                    });
                });
    }

    private void generateOperationPaginator(
            Model model,
            SymbolProvider symbolProvider,
            GoWriter writer,
            PaginationInfo paginationInfo
    ) {
        Symbol operationSymbol = symbolProvider.toSymbol(paginationInfo.getOperation());

        Symbol interfaceSymbol = SymbolUtils.createValueSymbolBuilder(
                OperationInterfaceGenerator.getApiClientInterfaceName(operationSymbol)
        ).build();
        Symbol paginatorSymbol = SymbolUtils.createPointableSymbolBuilder(String.format("%sPaginator",
                operationSymbol.getName())).build();
        Symbol optionsSymbol = SymbolUtils.createPointableSymbolBuilder(String.format("%sOptions",
                paginatorSymbol.getName())).build();

        writePaginatorOptions(writer, model, symbolProvider, paginationInfo, operationSymbol, optionsSymbol);
        writePaginator(writer, model, symbolProvider, paginationInfo, interfaceSymbol, paginatorSymbol, optionsSymbol);
    }

    private void writePaginator(
            GoWriter writer,
            Model model,
            SymbolProvider symbolProvider,
            PaginationInfo paginationInfo,
            Symbol interfaceSymbol,
            Symbol paginatorSymbol,
            Symbol optionsSymbol
    ) {
        String inputMember = symbolProvider.toMemberName(paginationInfo.getInputTokenMember());

        OperationShape operation = paginationInfo.getOperation();
        Optional<PagingExtensionTrait> pagingExtensionTrait = operation.getTrait(PagingExtensionTrait.class);

        Symbol operationSymbol = symbolProvider.toSymbol(operation);
        Symbol inputSymbol = symbolProvider.toSymbol(paginationInfo.getInput());
        Symbol inputTokenSymbol = symbolProvider.toSymbol(paginationInfo.getInputTokenMember());

        GoPointableIndex pointableIndex = GoPointableIndex.of(model);

        writer.writeDocs(String.format("%s is a paginator for %s", paginatorSymbol, operationSymbol.getName()));
        writer.openBlock("type $T struct {", "}", paginatorSymbol, () -> {
            writer.write("options $T", optionsSymbol);
            writer.write("client $T", interfaceSymbol);
            writer.write("params $P", inputSymbol);
            writer.write("nextToken $P", inputTokenSymbol);
            writer.write("firstPage bool");
        });
        writer.write("");

        Symbol newPagiantor = SymbolUtils.createValueSymbolBuilder(String.format("New%s",
                paginatorSymbol.getName())).build();
        writer.writeDocs(String.format("%s returns a new %s", newPagiantor.getName(), paginatorSymbol.getName()));
        writer.openBlock("func $T(client $T, params $P, optFns ...func($P)) $P {", "}",
                newPagiantor, interfaceSymbol, inputSymbol, optionsSymbol, paginatorSymbol, () -> {
                    writer.openBlock("if params == nil {", "}", () -> writer.write("params = &$T{}", inputSymbol));
                    writer.write("");
                    writer.write("options := $T{}", optionsSymbol);
                    paginationInfo.getPageSizeMember().ifPresent(memberShape -> {
                        GoValueAccessUtils.writeIfNonZeroValueMember(model, symbolProvider, writer, memberShape,
                                "params", op -> {
                                    op = CodegenUtils.getAsValueIfDereferencable(pointableIndex, memberShape, op);
                                    writer.write("options.Limit = $L", op);
                                });

                    });
                    writer.write("");
                    writer.openBlock("for _, fn := range optFns {", "}", () -> {
                        writer.write("fn(&options)");
                    });
                    writer.write("");
                    writer.openBlock("return &$T{", "}", paginatorSymbol, () -> {
                        writer.write("options: options,");
                        writer.write("client: client,");
                        writer.write("params: params,");
                        writer.write("firstPage: true,");
                    });
                });
        writer.write("");

        writer.writeDocs("HasMorePages returns a boolean indicating whether more pages are available");
        writer.openBlock("func (p $P) HasMorePages() bool {", "}", paginatorSymbol, () -> {
            writer.write("return p.firstPage || p.nextToken != nil");
        });
        writer.write("");

        Symbol contextSymbol = SymbolUtils.createValueSymbolBuilder("Context", SmithyGoDependency.CONTEXT)
                .build();
        Symbol outputSymbol = symbolProvider.toSymbol(paginationInfo.getOutput());
        Optional<MemberShape> pageSizeMember = paginationInfo.getPageSizeMember();

        writer.writeDocs(String.format("NextPage retrieves the next %s page.", operationSymbol.getName()));
        writer.openBlock("func (p $P) NextPage(ctx $T, optFns ...func(*Options)) ($P, error) {", "}",
                paginatorSymbol, contextSymbol, outputSymbol, () -> {
                    writer.addUseImports(SmithyGoDependency.FMT);
                    writer.openBlock("if !p.HasMorePages() {", "}", () -> {
                        writer.write("return nil, fmt.Errorf(\"no more pages available\")");
                    });

                    writer.write("");
                    writer.write("params := *p.params");
                    writer.write("params.$L = p.nextToken", inputMember);
                    pageSizeMember.ifPresent(memberShape -> {
                        writer.write("");
                        if (pointableIndex.isPointable(model.expectShape(memberShape.getTarget()))) {
                            writer.write("var limit $P", symbolProvider.toSymbol(memberShape));
                            writer.openBlock("if p.options.Limit > 0 {", "}", () -> {
                                writer.write("limit = &p.options.Limit");
                            });
                            writer.openBlock("params.$L = limit", symbolProvider.toMemberName(memberShape));
                        } else {
                            writer.openBlock("params.$L = p.options.Limit", symbolProvider.toMemberName(memberShape));
                        }
                    });

                    writer.write("");
                    writer.write("result, err := p.client.$L(ctx, &params, optFns...)",
                            operationSymbol.getName());
                    writer.openBlock("if err != nil {", "}", () -> {
                        writer.write("return nil, err");
                    });
                    writer.write("p.firstPage = false");
                    writer.write("");

                    List<MemberShape> outputMemberPath = paginationInfo.getOutputTokenMemberPath();
                    MemberShape tokenMember = outputMemberPath.get(outputMemberPath.size() - 1);
                    Consumer<String> setNextTokenFromOutput = (container) -> {
                        writer.write("p.nextToken = $L", container + "."
                                + symbolProvider.toMemberName(tokenMember));
                    };

                    for (int i = outputMemberPath.size() - 2; i >= 0; i--) {
                        MemberShape memberShape = outputMemberPath.get(i);
                        Consumer<String> inner = setNextTokenFromOutput;
                        setNextTokenFromOutput = (container) -> {
                            GoValueAccessUtils.writeIfNonZeroValueMember(model, symbolProvider, writer, memberShape,
                                    container, inner);
                        };
                    }

                    {
                        final Consumer<String> inner = setNextTokenFromOutput;
                        setNextTokenFromOutput = s -> {
                            if (outputMemberPath.size() > 1) {
                                writer.write("p.nextToken = nil");
                            }
                            inner.accept(s);
                        };
                    }

                    {
                        final Consumer<String> setToken = setNextTokenFromOutput;
                        writer.write("prevToken := p.nextToken");
                        Optional<MemberShape> moreResults = pagingExtensionTrait
                                .flatMap(PagingExtensionTrait::getMoreResults);

                        if (moreResults.isPresent()) {
                            MemberShape memberShape = moreResults.get();
                            model.expectShape(memberShape.getTarget(), BooleanShape.class); // Must be boolean
                            writer.write("p.nextToken = nil");
                            String memberName = symbolProvider.toMemberName(memberShape);
                            if (pointableIndex.isNillable(memberShape.getTarget())) {
                                writer.openBlock("if result.$L != nil && *result.$L {", "}", memberName, memberName,
                                        () -> setToken.accept("result"));
                            } else {
                                writer.openBlock("if result.$L {", "}", memberName, () -> setToken.accept("result"));
                            }
                        } else {
                            setToken.accept("result");
                        }
                    }
                    writer.write("");

                    if (model.expectShape(paginationInfo.getInputTokenMember().getTarget()).isStringShape()) {
                        writer.openBlock("if p.options.StopOnDuplicateToken && "
                                + "prevToken != nil && p.nextToken != nil && "
                                + "*prevToken == *p.nextToken {", "}", () -> {
                            writer.write("p.nextToken = nil");
                        });
                    } else {
                        writer.write("_ = prevToken");
                    }

                    writer.write("");
                    writer.write("return result, nil");
                });
    }

    private void writePaginatorOptions(
            GoWriter writer,
            Model model,
            SymbolProvider symbolProvider,
            PaginationInfo paginationInfo,
            Symbol operationSymbol,
            Symbol optionsSymbol
    ) {
        writer.writeDocs(String.format("%s is the paginator options for %s", optionsSymbol.getName(),
                operationSymbol.getName()));
        writer.openBlock("type $T struct {", "}", optionsSymbol, () -> {
            paginationInfo.getPageSizeMember().ifPresent(memberShape -> {
                memberShape.getMemberTrait(model, DocumentationTrait.class).ifPresent(documentationTrait -> {
                    writer.writeDocs(documentationTrait.getValue());
                });
                writer.write("Limit $T", symbolProvider.toSymbol(memberShape));
                writer.write("");
            });
            if (model.expectShape(paginationInfo.getInputTokenMember().getTarget()).isStringShape()) {
                writer.writeDocs("Set to true if pagination should stop if the service returns a pagination token that "
                        + "matches the most recent token provided to the service.");
                writer.write("StopOnDuplicateToken bool");
            }
        });
        writer.write("");
    }
}
