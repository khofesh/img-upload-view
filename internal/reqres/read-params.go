package reqres

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// ReadIDParam - read "id" from parameters (r.Context())
func ReadIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// ReadLimitParam - read "limit" from query parameters
func ReadLimitParam(r *http.Request) (int64, error) {
	limitParam := r.URL.Query().Get("limit")
	if limitParam == "" {
		return -1, nil
	}

	limit, err := strconv.ParseInt(limitParam, 10, 64)
	if err != nil || limit < 1 {
		return 0, errors.New("invalid limit parameter")
	}

	return limit, nil
}

// ReadOffsetParam - read "offset" from query parameters
func ReadOffsetParam(r *http.Request) (int64, error) {
	offsetParam := r.URL.Query().Get("offset")
	if offsetParam == "" {
		return 0, nil
	}

	offset, err := strconv.ParseInt(offsetParam, 10, 64)
	if err != nil || offset < 0 {
		return 0, errors.New("invalid offset parameter")
	}

	return offset, nil
}

func ReadInt(qs url.Values, key string, defaultValue int) (int, error) {
	s := qs.Get(key)

	if s == "" {
		return defaultValue, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue, fmt.Errorf("must be an integer value")
	}

	return i, nil
}

func ReadString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}
