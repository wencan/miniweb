package miniweb

import (
    "net/http"
)

type Output struct {
    Response http.ResponseWriter
}

func (self Output) Write(body []byte) (int, error) {
    return self.Write(body)
}

func (self Output) Return(status int, body []byte) (int, error) {
    self.Response.WriteHeader(status)
    return self.Response.Write(body)
}

func (self Output) Ok(body []byte) (int, error) {
    self.Response.WriteHeader(http.StatusOK)
    return self.Response.Write(body)
}
