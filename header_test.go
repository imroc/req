/*
  GoLang code created by Jirawat Harnsiriwatanakit https://github.com/kazekim
*/

package req

import "testing"

func TestParseStruct(t *testing.T) {

	type HeaderStruct struct {
		UserAgent     string `json:"User-Agent"`
		Authorization string `json:"Authorization"`
	}

	h := HeaderStruct{
		"V1.0.0",
		"roc",
	}

	var header Header
	header = ParseStruct(header, h)

	if header["User-Agent"] != h.UserAgent && header["Authorization"] != h.Authorization {
		t.Fatal("struct parser for header is not working")
	}

}

func TestHeaderFromStruct(t *testing.T) {

	type HeaderStruct struct {
		UserAgent string `json:"User-Agent"`
		Authorization string `json:"Authorization"`
	}

	h := HeaderStruct{
		"V1.0.0",
		"roc",
	}

	header := HeaderFromStruct(h)

	if header["User-Agent"] != h.UserAgent && header["Authorization"] != h.Authorization {
		t.Fatal("struct parser for header is not working")
	}
}
