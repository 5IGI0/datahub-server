package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

func ApiDecorator(fnc func(http.ResponseWriter, *http.Request) (any, int, string, error), flags int) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var response struct {
			Success      bool    `json:"success"`
			Error        *string `json:"error"`
			ErrorMessage *string `json:"error_message"`
			Data         any     `json:"data"`
		}

		w.Header().Add("Access-Control-Allow-Origin", "*")

		if err := ValidateToken(w, r, flags); err != nil {
			response.Success = false
			err_msg := err.Error()
			response.ErrorMessage = &err_msg
			w.Header().Add("Content-Type", "application/json")
			if errors.Is(err, ApiAccessDeniedErr) {
				error_str := "ACCESS_DENIED"
				response.Error = &error_str
				w.WriteHeader(http.StatusUnauthorized)
			} else if errors.Is(err, ApiRateLimitErr) {
				error_str := "RATE_LIMIT_EXCEEDED"
				response.Error = &error_str
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				error_str := "INTERNAL_ERROR"
				response.Error = &error_str
				w.WriteHeader(http.StatusInternalServerError)
			}

			encoded_response, _ := json.Marshal(response) // TODO: error
			w.Write(encoded_response)
			return
		}

		data, status_code, error_code, err := fnc(w, r)

		if err != nil {
			response.Data = nil
			response.Success = false
			response.Error = &error_code
			tmp := err.Error()
			response.ErrorMessage = &tmp
		} else {
			response.Data = data
			response.Success = true
			response.Error = nil
			response.ErrorMessage = nil
		}

		encoded_response, _ := json.Marshal(response) // TODO: error
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(status_code)
		w.Write(encoded_response)
	}
}

func ApiPostDecorator(fnc func(http.ResponseWriter, *http.Request) (any, int, string, error), flags int) func(http.ResponseWriter, *http.Request) {
	return ApiDecorator(func(w http.ResponseWriter, r *http.Request) (any, int, string, error) {
		if r.Method != "POST" {
			return nil, http.StatusMethodNotAllowed, "BAD_METHOD", errors.New("this endpoint only use POST")
		}
		if r.ContentLength == 0 && r.Body != nil {
			return nil, http.StatusLengthRequired, "NO_LENGHT_PROVIDED", errors.New("your request doesn't provide a Content-Length header")
		}
		if r.ContentLength > 50_000_000 {
			return nil, http.StatusRequestEntityTooLarge, "CONTENT_TOO_LARGE", errors.New("content too large")
		}
		return fnc(w, r)
	}, flags)
}
