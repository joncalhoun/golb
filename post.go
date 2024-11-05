package golb

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

type MetadataQuerier interface {
	Query() ([]PostMetadata, error)
}

type PostMetadata struct {
	Slug        string
	Title       string    `toml:"title"`
	Author      Author    `toml:"author"`
	Description string    `toml:"description"`
	Date        time.Time `toml:"date"`
}

type SlugReader interface {
	Read(slug string) (string, error)
}

type FileReader struct {
	// Directory to find blog posts in.
	Dir string
}

func (fr FileReader) Read(slug string) (string, error) {
	slugPath := filepath.Join(fr.Dir, slug+".md")
	f, err := os.Open(slugPath)
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

func (fr FileReader) Query() ([]PostMetadata, error) {
	postsPath := filepath.Join(fr.Dir, "*.md")
	filenames, err := filepath.Glob(postsPath)
	if err != nil {
		return nil, fmt.Errorf("querying for files: %w", err)
	}
	var posts []PostMetadata
	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("opening file %s: %w", filename, err)
		}
		defer f.Close()
		var post PostMetadata
		_, err = frontmatter.Parse(f, &post)
		if err != nil {
			return nil, fmt.Errorf("parsing frontmatter for file %s: %w", filename, err)
		}
		post.Slug = strings.TrimSuffix(filepath.Base(filename), ".md")
		posts = append(posts, post)
	}
	return posts, nil
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

func PostHandler(sl SlugReader, tpl *template.Template) http.HandlerFunc {
	mdRenderer := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
			AsideBlockExtension{},
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

type IndexData struct {
	Posts []PostMetadata
}

func IndexHandler(mq MetadataQuerier, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := mq.Query()
		if err != nil {
			http.Error(w, "Error querying posts", http.StatusInternalServerError)
			return
		}
		data := IndexData{
			Posts: posts,
		}
		err = tpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
	}
}
