package main

import (
	"log"
	"testing"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
)

func TestGettingUrlFromRequestNoParam(t *testing.T) {
	host := "http://localhost:8080/"

	qStr := "p=v1&p2=v2"
	req, err := http.NewRequest("GET", host + "?" + qStr, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = getTrgURL(req)
	descr := "The query '" + qStr + "' doesn't contain 'url' parameter"
	assert.Equal(t, err.(ProcessingError).Descr, ErrorNoURL)
	assert.Equal(t, err.(ProcessingError).InitErr.Error(), descr)
}

func TestGettingUrlFromRequestWrongVal(t *testing.T) {
	host := "http://localhost:8080/"

	urlVal := "aaaa"
	qStr := "url=" + urlVal + "&p2=v2"
	req, err := http.NewRequest("GET", host + "?" + qStr, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = getTrgURL(req)
	descr := "Parse '" + urlVal + "' into URL: invalid URI"
	assert.Equal(t, err.(ProcessingError).Descr, ErrorURLParsing)
	assert.Equal(t, err.(ProcessingError).InitErr.Error(), descr)
}

func TestGettingUrlFromRequestCorrectVal(t *testing.T) {
	host := "http://localhost:8080/"

	urlVal := "http%3A%2F%2Fexample.com%2Fdir%2F"
	qStr := "url=" + urlVal + "&p2=v2"
	req, err := http.NewRequest("GET", host + "?" + qStr, nil)
	if err != nil {
		log.Fatal(err)
	}

	trgUrl, _ := getTrgURL(req)
	assert.Equal(t, err, nil)
	assert.Equal(t, trgUrl.String(), "http://example.com/dir/")
}

func TestGettingSrcByInvalidUrl(t *testing.T) {
	trgUrl := "http:///dir/"
	url, err := url.ParseRequestURI(trgUrl)
	if err != nil {
		log.Fatal(err)
	}

	_, err = getOkHttpSrc(url)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, err.(ProcessingError).Descr, ErrorGetInt)
	msg := "Get " + trgUrl + ": http: no Host in request URL"
	assert.Equal(t, err.(ProcessingError).InitErr.Error(), msg)
}

func TestGettingNonexistedSrc(t *testing.T) {
	trgUrl := "http://vk.com/aaaaaaaaaaaa/"
	url, err := url.ParseRequestURI(trgUrl)
	if err != nil {
		log.Fatal(err)
	}

	_, err = getOkHttpSrc(url)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, err.(ForwardedError).StatusCode, 404)
}
