package apiresponse

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type ErrorDocument struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Title  string `json:"title"`
	Detail string `json:"detail,omitempty"`
}

func Write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, code, title, detail string) {
	Write(w, status, ErrorDocument{Errors: []Error{{
		Status: strconv.Itoa(status),
		Code:   code,
		Title:  title,
		Detail: detail,
	}}})
}
