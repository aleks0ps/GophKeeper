package http

import (
	"net/http"
	"strconv"
	"strings"
)

type ContentType uint64

const SContentType string = "Content-Type"

const (
	CTypeNone ContentType = iota
	CTypePlain
	CTypeCSS
	CTypeHTML
	CTypeXML
	CTypeJSON
	CTypeURLEncoded
	CTypeJS
)

const (
	STypeNone       = "none"
	STypePlain      = "text/plain"
	STypeCSS        = "text/css"
	STypeHTML       = "text/html"
	STypeXML        = "text/xml"
	STypeJSON       = "application/json"
	STypeURLEncoded = "application/x-www-form-urlencoded"
	STypeJS         = "application/javascript"
)

var contentTypeMap = map[string]ContentType{
	STypeNone:       CTypeNone,
	STypePlain:      CTypePlain,
	STypeCSS:        CTypeCSS,
	STypeHTML:       CTypeHTML,
	STypeXML:        CTypeXML,
	STypeJSON:       CTypeJSON,
	STypeURLEncoded: CTypeURLEncoded,
	STypeJS:         CTypeJS,
}

var reverseContentTypeMap = map[ContentType]string{
	CTypeNone:       STypeNone,
	CTypePlain:      STypePlain,
	CTypeCSS:        STypeCSS,
	CTypeHTML:       STypeHTML,
	CTypeXML:        STypeXML,
	CTypeJSON:       STypeJSON,
	CTypeURLEncoded: STypeURLEncoded,
	CTypeJS:         STypeJS,
}

func GetContentTypeCode(stype string) ContentType {
	stype = strings.ToLower(stype)
	ctype, ok := contentTypeMap[stype]
	if !ok {
		return CTypeNone
	}
	return ctype

}

func GetContentTypeName(code ContentType) string {
	stype, ok := reverseContentTypeMap[code]
	if !ok {
		return STypeNone
	}
	return stype
}

func WriteResponse(w http.ResponseWriter, t ContentType, status int, data []byte) {
	switch t {
	case CTypeNone:
		w.WriteHeader(status)
		if data != nil {
			w.Write(data)
		}
	default:
		w.Header().Set("Content-Type", GetContentTypeName(t))
		w.Header().Set("Content-Length", strconv.Itoa(len(data)))
		w.WriteHeader(status)
		w.Write(data)
	}
}

func WriteError(w http.ResponseWriter, status int, err error) {
	var errMsg string
	if err != nil {
		errMsg = err.Error()
	}
	http.Error(w, errMsg, status)
}
