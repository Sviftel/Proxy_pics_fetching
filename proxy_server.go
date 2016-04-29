package main

import (
	"strings"
	"container/list"
	"net/http"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	defer handleErrors(w)

	trgURL, err := getTrgURL(r)
	if err != nil {
		panic(err)
	}

	tree, err := getHtmlTree(trgURL)
	if err != nil {
		panic(err)
	}

	imgUrlList := list.New()
	findImgNodes(tree, imgUrlList)

	imgSrcs := make([]string, imgUrlList.Len())
	errc := make(chan error, imgUrlList.Len())
	for e, i := imgUrlList.Front(), 0; e != nil; e, i = e.Next(), i + 1 {
		go savePic(trgURL, e.Value.(string), &(imgSrcs[i]), errc)
	}

	for i := 0; i < imgUrlList.Len(); i++ {
		err := <-errc
		if err != nil {
			panic(err)
		}
	}

	fillRespTemplate(w, strings.Join(imgSrcs, ""))
}

func main() {
	http.HandleFunc("/", proxyHandler)
	http.ListenAndServe(":8080", nil)
}
