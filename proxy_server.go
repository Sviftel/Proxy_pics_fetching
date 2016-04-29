package main

import (
	// "errors"
	// "fmt"
	// "mime"
	// "strings"
	"container/list"
	// "encoding/base64"
	// "io/ioutil"
	"net/http"
	// "net/url"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	defer handleErrors(w)

	trgURL := getTrgURL(r)
	tree := getHtmlTree(trgURL)

	imgUrlList := list.New()
	findImgNodes(tree, imgUrlList)

	imgSrcs := make([]string, imgUrlList.Len())
	imgTagUnion := ""
	for e, i := imgUrlList.Front(), 0; e != nil; e, i = e.Next(), i + 1 {
		savePic(trgURL, e.Value.(string), &(imgSrcs[i]))
		imgTagUnion = imgTagUnion + imgSrcs[i]
		// fmt.Println(i, "Src:", e.Value.(string), ",len:", len(imgSrcs[i]))
	}

	fillRespTemplate(w, imgTagUnion)
}

func main() {
	http.HandleFunc("/", proxyHandler)
	http.ListenAndServe(":8080", nil)
}
