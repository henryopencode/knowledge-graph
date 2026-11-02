package content

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v3"
)

func Parse(raw []byte) (Document, error) {
	normalized := strings.ReplaceAll(strings.TrimPrefix(string(raw), "\ufeff"), "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return Document{}, Fail("INVALID_DOCUMENT", 400)
	}
	end := strings.Index(normalized[4:], "\n---\n")
	if end < 0 {
		return Document{}, Fail("INVALID_DOCUMENT", 400)
	}
	var metadata Metadata
	decoder := yaml.NewDecoder(strings.NewReader(normalized[4 : 4+end]))
	decoder.KnownFields(true)
	if err := decoder.Decode(&metadata); err != nil {
		return Document{}, fmt.Errorf("frontmatter: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Document{}, Fail("INVALID_DOCUMENT", 400)
	}
	d := Document{Metadata: metadata, Content: strings.TrimLeft(normalized[4+end+5:], "\n")}
	if d.Tags == nil {
		d.Tags = []string{}
	}
	if err := validate(d); err != nil {
		return Document{}, err
	}
	sum := sha256.Sum256(raw)
	d.Version = hex.EncodeToString(sum[:])
	return d, nil
}

func Encode(d Document) ([]byte, error) {
	if err := validate(d); err != nil {
		return nil, err
	}
	var header bytes.Buffer
	encoder := yaml.NewEncoder(&header)
	encoder.SetIndent(2)
	if err := encoder.Encode(d.Metadata); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return []byte("---\n" + header.String() + "---\n\n" + strings.TrimRight(d.Content, "\n") + "\n"), nil
}

type WikiLink struct {
	ast.BaseInline
	Target, Label string
	Start, End    int
}

var wikiKind = ast.NewNodeKind("DocumentLink")

func (w *WikiLink) Kind() ast.NodeKind            { return wikiKind }
func (w *WikiLink) Dump(source []byte, level int) { ast.DumpHelper(w, source, level, nil, nil) }

type wikiParser struct{}

func (*wikiParser) Trigger() []byte { return []byte{'['} }
func (*wikiParser) Parse(parent ast.Node, reader text.Reader, context parser.Context) ast.Node {
	line, segment := reader.PeekLine()
	if !bytes.HasPrefix(line, []byte("[[")) {
		return nil
	}
	end := bytes.Index(line[2:], []byte("]]"))
	if end < 0 {
		return nil
	}
	end += 2
	value := string(line[2:end])
	parts := strings.SplitN(value, "|", 2)
	target := strings.TrimSpace(strings.SplitN(parts[0], "#", 2)[0])
	target = strings.TrimSuffix(strings.TrimPrefix(target, "doc:"), ".md")
	if target == "" {
		return nil
	}
	node := &WikiLink{Target: target, Start: segment.Start, End: segment.Start + end + 2}
	if len(parts) > 1 {
		node.Label = parts[1]
	}
	reader.Advance(end + 2)
	return node
}

func markdownEngine() goldmark.Markdown {
	return goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithInlineParsers(util.Prioritized(&wikiParser{}, 150))))
}
func WikiLinks(body string) []WikiLink {
	source := []byte(body)
	tree := markdownEngine().Parser().Parse(text.NewReader(source))
	links := []WikiLink{}
	_ = ast.Walk(tree, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if link, ok := n.(*WikiLink); ok {
				links = append(links, *link)
			}
		}
		return ast.WalkContinue, nil
	})
	return links
}
