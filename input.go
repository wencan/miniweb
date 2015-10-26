package miniweb

import (
    "net/http"
)

type Input struct {
    Request *http.Request

    Fields       map[string]string
    QueryStrings map[string][]string
}
