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

import java.util.Set;
import org.commonmark.node.BlockQuote;
import org.commonmark.node.FencedCodeBlock;
import org.commonmark.node.Heading;
import org.commonmark.node.HtmlBlock;
import org.commonmark.node.ListBlock;
import org.commonmark.node.ThematicBreak;
import org.commonmark.parser.Parser;
import org.commonmark.renderer.html.HtmlRenderer;
import org.jsoup.Jsoup;
import org.jsoup.nodes.Node;
import org.jsoup.nodes.TextNode;
import org.jsoup.safety.Safelist;
import org.jsoup.select.NodeTraversor;
import org.jsoup.select.NodeVisitor;
import software.amazon.smithy.utils.CodeWriter;
import software.amazon.smithy.utils.SetUtils;
import software.amazon.smithy.utils.StringUtils;

/**
 * Converts commonmark-formatted documentation into godoc format.
 */
public final class DocumentationConverter {
    // godoc only supports text blocks, root-level non-inline code blocks, headers, and links.
    // This allowlist strips out anything we can't reasonably convert, vastly simplifying the
    // node tree we end up having to crawl through.
    private static final Safelist GODOC_ALLOWLIST = new Safelist()
            .addTags("code", "pre", "ul", "ol", "li", "a", "br", "h1", "h2", "h3", "h4", "h5", "h6")
            .addAttributes("a", "href")
            .addProtocols("a", "href", "http", "https", "mailto");

    // Construct a markdown parser that specifically ignores parsing indented code blocks. This
    // is because HTML blocks can have really wonky formatting that can be mis-attributed to an
    // indented code blocks. We may need to add a configuration option to re-enable this.
    private static final Parser MARKDOWN_PARSER = Parser.builder()
            .enabledBlockTypes(SetUtils.of(
                    Heading.class, HtmlBlock.class, ThematicBreak.class, FencedCodeBlock.class,
                    BlockQuote.class, ListBlock.class))
            .build();

    private DocumentationConverter() {}

    /**
     * Converts a commonmark formatted string into a godoc formatted string.
     *
     * @param docs commonmark formatted documentation
     * @return godoc formatted documentation
     */
    public static String convert(String docs) {
        // Smithy's documentation format is commonmark, which can inline html. So here we convert
        // to html so we have a single known format to work with.
        String htmlDocs = HtmlRenderer.builder().escapeHtml(false).build().render(MARKDOWN_PARSER.parse(docs));

        // Strip out tags and attributes we can't reasonably convert to godoc.
        htmlDocs = Jsoup.clean(htmlDocs, GODOC_ALLOWLIST);

        // Now we parse the html and visit the resultant nodes to render the godoc.
        FormattingVisitor formatter = new FormattingVisitor();
        Node body = Jsoup.parse(htmlDocs).body();
        NodeTraversor.traverse(formatter, body);
        return formatter.toString();
    }

    private static class FormattingVisitor implements NodeVisitor {
        private static final Set<String> TEXT_BLOCK_NODES = SetUtils.of(
                "br", "p", "h1", "h2", "h3", "h4", "h5", "h6"
        );
        private static final Set<String> LIST_BLOCK_NODES = SetUtils.of("ul", "ol");
        private static final Set<String> CODE_BLOCK_NODES = SetUtils.of("pre", "code");
        private final CodeWriter writer;

        private boolean needsListPrefix = false;
        private boolean shouldStripPrefixWhitespace = false;

        FormattingVisitor() {
            writer = new CodeWriter();
            writer.trimTrailingSpaces(false);
            writer.trimBlankLines();
            writer.insertTrailingNewline(false);
        }

        @Override
        public void head(Node node, int depth) {
            String name = node.nodeName();
            if (isTopLevelCodeBlock(node, depth)) {
                writer.indent();
            }

            if (node instanceof TextNode) {
                writeText((TextNode) node);
            } else if (TEXT_BLOCK_NODES.contains(name) || isTopLevelCodeBlock(node, depth)) {
                writeNewline();
                writeIndent();
            } else if (LIST_BLOCK_NODES.contains(name)) {
                writeNewline();
            } else if (name.equals("li")) {
                // We don't actually write out the list prefix here in case the list element
                // starts with one or more text blocks. By deferring writing those out until
                // the first bit of actual text, we can ensure that no intermediary newlines
                // are kept. It also has the added benefit of eliminating empty list elements.
                needsListPrefix = true;
            }
        }

        private void writeNewline() {
            // While jsoup will strip out redundant whitespace, it will still leave some. If we
            // start a new line then we want to make sure we don't keep any prefixing whitespace.
            shouldStripPrefixWhitespace = true;
            writer.write("");
        }

        private void writeText(TextNode node) {
            if (node.isBlank()) {
                return;
            }

            // Docs can have valid $ characters that shouldn't run through formatters.
            String text = node.text().replace("$", "$$");

            if (shouldStripPrefixWhitespace) {
                shouldStripPrefixWhitespace = false;
                text = StringUtils.stripStart(text, " \t");
            }

            if (needsListPrefix) {
                needsListPrefix = false;
                writer.write("");
                writeIndent();
                text = "* " + StringUtils.stripStart(text, " \t");
            }
            writer.writeInline(text);
        }

        void writeIndent() {
            writer.setNewline("").write("").setNewline("\n");
        }

        private boolean isTopLevelCodeBlock(Node node, int depth) {
            // The node must be a code block node
            if (!CODE_BLOCK_NODES.contains(node.nodeName())) {
                return false;
            }

            // It must either have no siblings or its siblings must be separate blocks.
            if (!allSiblingsAreBlocks(node)) {
                return false;
            }

            // Depth 0 will always be a "body" element, so depth 1 means it's top level.
            if (depth == 1) {
                return true;
            }

            // If its depth is 2, it could still be effectively top level if its parent is a p
            // node whose siblings are all blocks.
            Node parent = node.parent();
            return depth == 2 && parent.nodeName().equals("p") && allSiblingsAreBlocks(parent);
        }

        /**
         * Determines whether a given node's siblings are all text blocks, code blocks, or lists.
         *
         * <p>Siblings that are blank text nodes are skipped.
         *
         * @param node The node whose siblings should be checked.
         * @return true if the node's siblings are blocks, otherwise false.
         */
        private boolean allSiblingsAreBlocks(Node node) {
            // Find the nearest sibling to the left which is not a blank text node.
            Node previous = node.previousSibling();
            while (true) {
                if (previous instanceof TextNode) {
                    if (((TextNode) previous).isBlank()) {
                        previous = previous.previousSibling();
                        continue;
                    }
                }
                break;
            }

            // Find the nearest sibling to the right which is not a blank text node.
            Node next = node.nextSibling();
            while (true) {
                if (next instanceof TextNode) {
                    if (((TextNode) next).isBlank()) {
                        next = next.nextSibling();
                        continue;
                    }
                }
                break;
            }

            return (previous == null || isBlockNode(previous)) && (next == null || isBlockNode(next));
        }

        private boolean isBlockNode(Node node) {
            String name = node.nodeName();
            return TEXT_BLOCK_NODES.contains(name) || LIST_BLOCK_NODES.contains(name)
                    || CODE_BLOCK_NODES.contains(name);
        }

        @Override
        public void tail(Node node, int depth) {
            String name = node.nodeName();
            if (isTopLevelCodeBlock(node, depth)) {
                writer.dedent();
            }

            if (TEXT_BLOCK_NODES.contains(name) || isTopLevelCodeBlock(node, depth)
                    || LIST_BLOCK_NODES.contains(name)) {
                writeNewline();
                writeNewline();
            } else if (name.equals("a")) {
                String url = node.absUrl("href");
                if (!url.isEmpty()) {
                    // godoc can't render links with text bodies, so we simply append the link.
                    // Full links do get rendered.
                    writer.writeInline(" ($L)", url);
                }
            } else if (name.equals("li")) {
                // Clear out the expectation of a list element if the element's body is empty.
                needsListPrefix = false;
                writer.write("");
            }
        }

        @Override
        public String toString() {
            String result = writer.toString();
            if (StringUtils.isBlank(result)) {
                return "";
            }

            // Strip trailing whitespace from every line. We can't use the codewriter for this due to
            // not knowing when a line will end, as we typically build them up over many elements.
            String[] lines = result.split("\n", -1);
            for (int i = 0; i < lines.length; i++) {
                lines[i] = StringUtils.stripEnd(lines[i], " \t");
            }
            result = String.join("\n", lines);

            // Strip out leading and trailing newlines.
            return StringUtils.strip(result, "\n");
        }
    }
}
