package main

import (
	"log"
	"testing"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func TestGolangDocProxying (t *testing.T) {
	url := "http://localhost:8080/?url=https%3A%2F%2Fgolang.org%2Fdoc%2F"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
		t.Fail()
	}

	w := httptest.NewRecorder()
	proxyHandler(w, req)

	doc, err := ioutil.ReadFile("./golang_doc/doc.png")
	if err != nil {
		log.Fatal(err)
		t.Fail()
	}

	talks, err1 := ioutil.ReadFile("./golang_doc/talks.png")
	if err1 != nil {
		log.Fatal(err)
		t.Fail()
	}

	pref := "<img src=\"data:image/png;base64,"
	suff := "\">"
	docTag := pref + base64.StdEncoding.EncodeToString(doc) + suff
	talksTag := pref + base64.StdEncoding.EncodeToString(talks) + suff
	beg := "<html><head></head><body>"
	end := "</body></html>"

	v1 := beg + docTag + talksTag + end
	v2 := beg + talksTag + docTag + end

	assert.True(t, w.Body.String() == v1 || w.Body.String() == v2)
}

// Работа также была протестирована на ru.wikipedia.org,
// python.org и др.