+++
title = "Go's io.Reader"
description = "A quick exploration of Go's io.Reader interface."
date = 2024-03-10

[author]
name = "Jon Calhoun"
email = "jon@calhoun.io"
+++

Go's io.Reader is defined as:

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

::: This is the title of the aside
This is some extra side content that is relevant, but not necessary to enjoy the blog post.
:::


## Next up

Put more content here under the h2 tag.
