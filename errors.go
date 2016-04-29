package main

import "net/http"

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

type ErrorTemplateFiller struct {
	StatusCode int
	Title      string
	Header     string
	Descr      string
}

func handleErrors(w http.ResponseWriter) {
	if handlingError := recover(); handlingError != nil {
		switch err := handlingError.(type) {
		case ProcessingError:
			flr := ErrorTemplateFiller{0, "", "", ""}
			if err.Descr == ErrorNoURL || err.Descr == ErrorURLParsing {
				flr = ErrorTemplateFiller {
					StatusCode: 422,
					Title: "422 Unprocessable Entity",
					Header: "Unprocessable Entity",
					Descr: err.InitErr.Error(),
				}
			} else if err.Descr == ErrorInvalidFileType {
				// TODO: find better error code
				flr = ErrorTemplateFiller {
					StatusCode: 422,
					Title: "422 Unprocessable Entity",
					Header: "Unprocessable Entity",
					Descr: err.InitErr.Error(),
				}
			} else {
				s := string(http.StatusInternalServerError)
				s = s + " Internal Server Error"
				flr = ErrorTemplateFiller {
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
