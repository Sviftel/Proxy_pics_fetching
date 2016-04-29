package main

import (
	"errors"
	"mime"
	"strings"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"net/url"
)

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
