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

type IndexData struct {
	Posts []PostMetadata
}

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
		err = tpl.Execute(w, data)
		if err != nil {
			log.Printf("Error: %v", err)
			fmt.Fprint(w, "Something went wrong...")
			return
		}
	}
}
