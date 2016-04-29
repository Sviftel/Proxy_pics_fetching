package main

import (
	"bytes"
	html "golang.org/x/net/html"
	"container/list"
	"net/url"
)

func getHtmlTree(URL *url.URL) (*html.Node, error) {
	src, err := getOkHttpSrc(URL)
	if err != nil {
		return nil, err
	}

	oldRoot, err := html.Parse(bytes.NewReader(src))
	if err != nil {
		return nil, ProcessingError{Descr: ErrorHtmlParsing, InitErr: err}
	}

	var buf bytes.Buffer
	err = html.Render(&buf, oldRoot)
	if err != nil {
		return nil, ProcessingError{Descr: ErrorHtmlParsing, InitErr: err}
	}

	newRoot, err := html.Parse(&buf)
	if err != nil {
		return nil, ProcessingError{Descr: ErrorHtmlParsing, InitErr: err}
	}

	return newRoot, nil
}

func findImgNodes(n *html.Node, imgUrlList *list.List) {
	if n.Type == html.ElementNode && n.Data == "img" {
		for _, attr := range n.Attr {
			if attr.Key == "src" {
				imgUrlList.PushBack(attr.Val)
				break
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findImgNodes(c, imgUrlList)
	}
}
