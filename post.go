package golb

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

type PostMetadata struct {
	Slug        string
	Title       string    `toml:"title"`
	Author      Author    `toml:"author"`
	Description string    `toml:"description"`
	Date        time.Time `toml:"date"`
}

type PostData struct {
	Content template.HTML
	Title   string `toml:"title"`
	Author  Author `toml:"author"`
}

type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

type MetadataQuerier interface {
	Query() ([]PostMetadata, error)
}

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct{}

func (fr FileReader) Query() ([]PostMetadata, error) {
	filenames, err := filepath.Glob("posts/*.md")
	if err != nil {
		return nil, fmt.Errorf("querying for files: %w", err)
	}
	var posts []PostMetadata
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer f.Close()
		var post PostMetadata
		_, err = frontmatter.Parse(f, &post)
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter: %w", err)
		}
		post.Slug = strings.TrimSuffix(filepath.Base(filename), ".md")
		posts = append(posts, post)
	}
	return posts, nil
}

func (fr FileReader) Read(slug string) (string, error) {
	f, err := os.Open("posts/" + slug + ".md")
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func PostHandler(sl SlugReader, tpl *template.Template) http.HandlerFunc {
	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
			&AsideBlockExtension{},
		),
	)
	return func(w http.ResponseWriter, r *http.Request) {
		slug := r.PathValue("slug")
		postMarkdown, err := sl.Read(slug)
		if err != nil {
			// TODO: Handle different errors in the future
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		var post PostData
		remainingMd, err := frontmatter.Parse(strings.NewReader(postMarkdown), &post)
		if err != nil {
			http.Error(w, "Error parsing frontmatter", http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		err = mdRenderer.Convert([]byte(remainingMd), &buf)
		if err != nil {
			panic(err)
		}
		post.Content = template.HTML(buf.String())

		err = tpl.Execute(w, post)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}

func IndexHandler(mq MetadataQuerier, tpl *template.Template) http.HandlerFunc {
	type Link struct {
		Text string
		Href string
	}
	type PaginationItem struct {
		Link     Link
		Active   bool
		Ellipsis bool
	}
	type IndexData struct {
		Posts      []PostMetadata
		Pagination struct {
			Previous string
			Next     string
			Items    []PaginationItem
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := mq.Query()
		if err != nil {
			log.Println(err)
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
			return
		}
		data := IndexData{
			Posts: posts,
		}
		// Fake this for demo purposes
		data.Pagination.Previous = ""
		data.Pagination.Next = "?page=2"
		data.Pagination.Items = []PaginationItem{
			{Link: Link{Text: "1", Href: "?page=1"}, Active: true},
			{Link: Link{Text: "2", Href: "?page=2"}},
			{Link: Link{Text: "3", Href: "?page=3"}},
			{Link: Link{Text: "4", Href: "?page=4"}},
			{Link: Link{Text: "5", Href: "?page=5"}},
			{Ellipsis: true},
			{Link: Link{Text: "10", Href: "?page=10"}},
		}

		err = tpl.Execute(w, data)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprint(w, "Something went wrong...")
			return
		}
	}
}
