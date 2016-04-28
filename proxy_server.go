package main

import (
	"bytes"
	"errors"
	// "fmt"
	"mime"
	"strings"
	html "golang.org/x/net/html"
	"container/list"
	"encoding/base64"
	"io/ioutil"
	"html/template"
	textT "text/template"
	"net/http"
	"net/url"
)

const (
	ErrorNoURL           = "NO_URL"
	ErrorURLParsing      = "URL_PARSING_FAILED"
	ErrorGetInt          = "INT_GET_FAILED"
	ErrorReadResp        = "READ_FAILED"
	ErrorHtmlParsing     = "HTML_PARSING_FAILED"
	ErrorInvalidFileType = "INVALID_FILE_TYPE_IN_IMG_SRC"
)

type ForwardedError struct {
	StatusCode int
	Body       []byte
}

type ProcessingError struct {
	Descr   string
	InitErr error
}

type TemplateFiller struct {
	StatusCode int
	Title      string
	Header     string
	Descr      string
}

var htmlTemplates = template.Must(template.ParseFiles("error_msg_temp.html"))
var textTemplates = textT.Must(textT.ParseFiles("resp_temp.html"))


func fillErrorTemplate(w http.ResponseWriter, flr *TemplateFiller) {
	w.WriteHeader((*flr).StatusCode)
	innerErr := htmlTemplates.ExecuteTemplate(w, "error_msg_temp.html", *flr)
	if innerErr != nil {
		http.Error(w, innerErr.Error(),
			http.StatusInternalServerError)
	}
}

func handleErrors(w http.ResponseWriter) {
	if handlingError := recover(); handlingError != nil {
		switch err := handlingError.(type) {
		case ProcessingError:
			flr := TemplateFiller{0, "", "", ""}
			if err.Descr == ErrorNoURL || err.Descr == ErrorURLParsing {
				flr = TemplateFiller {
					StatusCode: 422,
					Title: "422 Unprocessable Entity",
					Header: "Unprocessable Entity",
					Descr: err.InitErr.Error(),
				}
			} else if err.Descr == ErrorInvalidFileType {
				flr = TemplateFiller {
					StatusCode: 422,
					Title: "422 Unprocessable Entity",
					Header: "Unprocessable Entity",
					Descr: err.InitErr.Error(),
				}
			} else {
				s := string(http.StatusInternalServerError)
				s = s + " Internal Server Error"
				flr = TemplateFiller {
					StatusCode: http.StatusInternalServerError,
					Title: s,
					Header: "Internal Server Error",
					Descr: err.InitErr.Error(),
				}
			}
			fillErrorTemplate(w, &flr)
		case ForwardedError:
			w.WriteHeader(err.StatusCode)
			_, innerErr := w.Write(err.Body)
			if innerErr != nil {
				http.Error(w, innerErr.Error(), http.StatusInternalServerError)
			}
		default:
			panic("Unknown error")
		}
	}
}

func getTrgURL(r *http.Request) *url.URL {
	queryVals := r.URL.Query()
	if val, ok := queryVals["url"]; ok {
		if url, err := url.Parse(val[0]); err == nil {
			return url
		}

		s := "Couldn't parse '" + val[0] + "' into URL"
		panic(ProcessingError{Descr: ErrorURLParsing, InitErr: errors.New(s)})
	}

	s := "The query '" + string(r.URL.RawQuery)
	s = s + "' doesn't contain 'url' parameter"
	panic(ProcessingError{Descr: ErrorNoURL, InitErr: errors.New(s)})
}

func getOkHttpSrc(URL *url.URL) []byte {
	// TODO: think about redirections
	resp, err := http.Get(URL.String())
	if err != nil {
		panic(ProcessingError{Descr: ErrorGetInt, InitErr: err})
	}
	defer resp.Body.Close()

	src, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(ProcessingError{Descr: ErrorReadResp, InitErr: err})
	}

	if resp.StatusCode >= 300 {
		panic(ForwardedError{StatusCode: resp.StatusCode, Body: src})
	}

	return src
}

func getHtmlTree(URL *url.URL) *html.Node {
	src := getOkHttpSrc(URL)

	tree, err := html.Parse(bytes.NewReader(src))
	if err != nil {
		panic(ProcessingError{Descr: ErrorHtmlParsing, InitErr: err})
	}

	return tree
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

func savePic(initUrl *url.URL, src string, pic *string) {
	if srcUrl, err := url.Parse(src); err == nil {
		pref := "<img src=\"data:"
		suff := "\">\n"

		splits := strings.Split(srcUrl.Path, ".")
		ext := "." + splits[len(splits) - 1]
		mimeType := mime.TypeByExtension(ext)
		if !strings.HasPrefix(mimeType, "image/") {
			s := "Source '" + srcUrl.Path + "' has incorrect MIME type"
			panic(ProcessingError{
				Descr: ErrorInvalidFileType,
				InitErr: errors.New(s)})
		}

		pref = pref + mimeType + ";base64,"
		data := getOkHttpSrc(initUrl.ResolveReference(srcUrl))
		*pic = pref + base64.StdEncoding.EncodeToString(data) + suff
	}
}

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

	// fmt.Println("Union len:", len(imgTagUnion))
	innerErr := textTemplates.ExecuteTemplate(w, "resp_temp.html", imgTagUnion)
	if innerErr != nil {
		http.Error(w, innerErr.Error(),
			http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/", proxyHandler)
	http.ListenAndServe(":8080", nil)
}
