package main

import (
	"html/template"
	textT "text/template"
	"net/http"
)

var htmlTemplates = template.Must(template.ParseFiles("error_msg_temp.html"))
var textTemplates = textT.Must(textT.ParseFiles("resp_temp.html"))

func fillErrorTemplate(w http.ResponseWriter, flr *ErrorTemplateFiller) {
	w.WriteHeader((*flr).StatusCode)
	innerErr := htmlTemplates.ExecuteTemplate(w, "error_msg_temp.html", *flr)
	if innerErr != nil {
		http.Error(w, innerErr.Error(),
			http.StatusInternalServerError)
	}
}

func fillRespTemplate(w http.ResponseWriter, res string) {
	innerErr := textTemplates.ExecuteTemplate(w, "resp_temp.html", res)
	if innerErr != nil {
		http.Error(w, innerErr.Error(),
		http.StatusInternalServerError)
	}
}
