package main

import (
	"net/http"
	"net/url"
)

const (
	ErrorNoURL                 = "NO_URL"
	ErrorURLParsing            = "URL_PARSING_FAILED"
	ErrorGetInt                = "INT_GET_FAILED"
	ErrorReadResp              = "READ_FAILED"
	ErrorHtmlParsing           = "HTML_PARSING_FAILED"
	// TODO: find better error code
	ErrorInvalidFileType       = "INVALID_FILE_TYPE_IN_IMG_SRC"
	ErrorUnsupportedDataScheme = "UNSUPPTRED_DATA_SCHEME_FOR_IMGS"
)

type ForwardedError struct {
	StatusCode int
	Body       []byte
}

func (e ForwardedError) Error() string {
	return string(e.Body)
}

type ProcessingError struct {
	Descr   string
	InitErr error
}

func (e ProcessingError) Error() string {
	return e.InitErr.Error()
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
			switch {
			case err.Descr == ErrorNoURL || err.Descr == ErrorURLParsing:
				flr = ErrorTemplateFiller {
					StatusCode: 422,
					Title: "422 Unprocessable Entity",
					Header: "Unprocessable Entity",
					Descr: err.InitErr.Error(),
				}
			case err.Descr == ErrorGetInt:
				switch err.InitErr.(type) {
				case *url.Error:
					// TODO: check for better error handling
					flr = ErrorTemplateFiller {
						StatusCode: 404,
						Title: "404 Not Found",
						Header: "Not Found",
						Descr: err.InitErr.Error(),
					}
				default:
					flr = ErrorTemplateFiller {
						StatusCode: 422,
						Title: "422 Unprocessable Entity",
						Header: "Unprocessable Entity",
						Descr: err.InitErr.Error(),
					}
				}
			default:
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
