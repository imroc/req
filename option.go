package req

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// http request header param
type Header map[string]string

// http request param
type Param map[string]string

// used to force append http request param to the uri
type QueryParam map[string]string

// used for set request's Host
type Host string

// represents a file to upload
type FileUpload struct {
	// filename in multipart form.
	FileName string
	// form field name
	FieldName string
	// file to uplaod, required
	File io.ReadCloser
}

type bodyJson struct {
	v interface{}
}

// BodyJSON make the object be encoded in json format and set it to the request body
func BodyJSON(v interface{}) *bodyJson {
	return &bodyJson{v: v}
}

type bodyXml struct {
	v interface{}
}

// BodyXML make the object be encoded in xml format and set it to the request body
func BodyXML(v interface{}) *bodyXml {
	return &bodyXml{v: v}
}

// File upload files matching the name pattern such as
// /usr/*/bin/go* (assuming the Separator is '/')
func File(patterns ...string) interface{} {
	matches := []string{}
	for _, pattern := range patterns {
		m, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		matches = append(matches, m...)
	}
	if len(matches) == 0 {
		return errors.New("req: No file have been matched")
	}
	uploads := []FileUpload{}
	for _, match := range matches {
		if s, e := os.Stat(match); e != nil || s.IsDir() {
			continue
		}
		file, _ := os.Open(match)
		uploads = append(uploads, FileUpload{File: file, FileName: filepath.Base(match)})
	}

	return uploads
}
