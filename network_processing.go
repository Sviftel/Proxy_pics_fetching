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

func getTrgURL(r *http.Request) (*url.URL, error) {
	queryVals := r.URL.Query()
	if val, ok := queryVals["url"]; ok {
		url, err := url.ParseRequestURI(val[0])
		if err == nil {
			return url, nil
		}

		s := "Parse '" + val[0] + "' into URL: invalid URI"
		nErr := ProcessingError{Descr: ErrorURLParsing, InitErr: errors.New(s)}
		return nil, nErr
	}

	s := "The query '" + string(r.URL.RawQuery)
	s = s + "' doesn't contain 'url' parameter"
	return nil, ProcessingError{Descr: ErrorNoURL, InitErr: errors.New(s)}
}

func getOkHttpSrc(URL *url.URL) ([]byte, error) {
	resp, err := http.Get(URL.String())
	if err != nil {
		return nil, ProcessingError{Descr: ErrorGetInt, InitErr: err}
	}
	defer resp.Body.Close()

	src, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, ProcessingError{Descr: ErrorReadResp, InitErr: err}
	}

	if resp.StatusCode >= 300 {
		return nil, ForwardedError{StatusCode: resp.StatusCode, Body: src}
	}

	return src, nil
}

func savePic(initUrl *url.URL, src string, pic *string, errc chan error) {
	if strings.HasPrefix(src, "data:") {
		*pic = "<img src=\"" + src + "\">"
		errc <- nil
	} else if srcUrl, err := url.Parse(src); err == nil {
		pref := "<img src=\"data:"
		suff := "\">"

		splits := strings.Split(srcUrl.Path, ".")
		ext := "." + splits[len(splits) - 1]
		mimeType := mime.TypeByExtension(ext)

		if !strings.HasPrefix(mimeType, "image/") {
			s := "Source '" + srcUrl.Path + "' has incorrect MIME type"
			err := ProcessingError{
				Descr: ErrorInvalidFileType,
				InitErr: errors.New(s),
			}
			errc <- err
		}

		pref = pref + mimeType + ";base64,"
		data, err := getOkHttpSrc(initUrl.ResolveReference(srcUrl))
		if err != nil {
			errc <- err
		}
		*pic = pref + base64.StdEncoding.EncodeToString(data) + suff
		errc <- nil
	} else {
		s := "Unsupported data scheme: '" + src + "'"
		err := ProcessingError{
			Descr: ErrorUnsupportedDataScheme,
			InitErr: errors.New(s),
		}
		errc <- err
	}
}
