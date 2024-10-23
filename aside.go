package golb

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type AsideBlockExtension struct{}

// Extend implements goldmark.Extender.
func (ext *AsideBlockExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&AsideBlockParser{}, 101),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&AsideBlockHTMLRenderer{}, 101),
	))
}

var KindAsideBlock = ast.NewNodeKind("AsideBlock")

// This helps us ensure it implements the Node interface.
var _ ast.Node = &AsideBlock{}

type AsideBlock struct {
	ast.BaseBlock
	Title   string
	Content []string
}

func (n *AsideBlock) Kind() ast.NodeKind {
	return KindAsideBlock
}

func (n *AsideBlock) Dump(source []byte, level int) {
	m := map[string]string{
		"Title": fmt.Sprintf("\"%s\"", n.Title),
	}
	ast.DumpHelper(n, source, level, m, nil)
}

// func (n *AsideBlock) Append(line string) {
// 	n.Content = append(n.Content, line)
// }

var _ parser.BlockParser = &AsideBlockParser{}

type AsideBlockParser struct{}

// Trigger returns a list of characters that triggers Parse method of
// this parser.
// If Trigger returns a nil, Open will be called with any lines.
func (ext *AsideBlockParser) Trigger() []byte {
	return []byte{':'}
}

// Open parses the current line and returns a result of parsing.
//
// Open must not parse beyond the current line.
// If Open has been able to parse the current line, Open must advance a reader
// position by consumed byte length.
//
// If Open has not been able to parse the current line, Open should returns
// (nil, NoChildren). If Open has been able to parse the current line, Open
// should returns a new Block node and returns HasChildren or NoChildren.
func (ext *AsideBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	// We need at least ":::" which is 3 bytes, so we check to verify
	// that the line is 3 bytes long first. If not, we return nil.
	if len(line) < 3 {
		return nil, parser.NoChildren
	}
	// Now that we know the line is at least 3 bytes long, we can check if it
	// matches what we want.
	if !bytes.Equal(line[0:3], []byte(":::")) {
		return nil, parser.NoChildren
	}
	// If we get here we have an aside block, but we don't have a type for that yet so we will return nil for now.
	aside := &AsideBlock{
		Title: strings.TrimSpace(string(line[3:])),
	}
	reader.Advance(segment.Len() - 1)
	return aside, parser.HasChildren
}

// Continue parses the current line and returns a result of parsing.
//
// Continue must not parse beyond the current line.
// If Continue has been able to parse the current line, Continue must advance
// a reader position by consumed byte length.
//
// If Continue has not been able to parse the current line, Continue should
// returns Close. If Continue has been able to parse the current line,
// Continue should returns (Continue | NoChildren) or
// (Continue | HasChildren)
func (ext *AsideBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if len(line) >= 3 && bytes.Equal(line[0:3], []byte(":::")) {
		reader.Advance(segment.Len())
		return parser.Close
	}
	return parser.Continue | parser.HasChildren
}

// func (p *AsideBlockParser) Continue(node ast.Node, reader text.Reader, ctx parser.Context) parser.State {

// 	// p.init()
// 	fmt.Println("continue...")

// 	// TODO: Check this kind first so no error
// 	n := node.(*AsideBlock)
// 	line, segment := reader.PeekLine()
// 	fmt.Println("got line:", string(line))
// 	if len(line) >= 3 && bytes.Equal(line[0:3], []byte(":::")) {
// 		reader.Advance(segment.Len())
// 		return parser.Close
// 	}
// 	n.Append(string(line))
// 	// reader.Advance(segment.Len() - 1)
// 	// if delim, count := lineDelim(line); delim != 0 {
// 	// 	if delim == n.Format.Delim && count == n.DelimCount {
// 	// 		reader.Advance(seg.Len())
// 	// 		return parser.Close
// 	// 	}
// 	// }
// 	// n.Segment.Stop = seg.Stop
// 	return parser.Continue | parser.HasChildren
// }

// Close will be called when the parser returns Close.
func (ext *AsideBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// noop
}

// CanInterruptParagraph returns true if the parser can interrupt paragraphs,
// otherwise false.
func (ext *AsideBlockParser) CanInterruptParagraph() bool {
	return false
}

// CanAcceptIndentedLine returns true if the parser can open new node when
// the given line is being indented more than 3 spaces.
func (ext *AsideBlockParser) CanAcceptIndentedLine() bool {
	return false
}

// HTML Renderer stuff
var _ renderer.NodeRenderer = &AsideBlockHTMLRenderer{}

type AsideBlockHTMLRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (rend *AsideBlockHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAsideBlock, rend.render)
}

func (rend *AsideBlockHTMLRenderer) render(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if asideBlock, ok := n.(*AsideBlock); ok && entering {
		// TODO: Watch for injection?
		fmt.Fprintf(w, `<aside><h4>%s</h4>`, asideBlock.Title)
		return ast.WalkContinue, nil
	}
	fmt.Fprintf(w, "</aside>")
	return ast.WalkContinue, nil

	// if !entering {
	// 	return ast.WalkContinue, nil
	// }
	// if n.Kind() != KindAsideBlock {
	// 	return ast.WalkContinue, nil
	// }

	// asideBlock.ChildCount()
	// asideBlock.FirstChild()
	// asideBlock.FirstChild().Kind()
	// asideBlock.FirstChild()
	// fmt.Fprintf(w, "<aside><b>%s</b><p>%v</p>%d %v</aside>\n", asideBlock.Title, asideBlock.Content, asideBlock.ChildCount(), asideBlock.FirstChild())
	// return ast.WalkContinue, nil
}
