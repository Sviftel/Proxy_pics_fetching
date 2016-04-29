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

    tree, err := html.Parse(bytes.NewReader(src))
    if err != nil {
        // panic(ProcessingError{Descr: ErrorHtmlParsing, InitErr: err})
        return nil, ProcessingError{Descr: ErrorHtmlParsing, InitErr: err}
    }

    return tree, nil
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
