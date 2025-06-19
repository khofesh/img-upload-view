package middlewares

import (
	errres "github.com/khofesh/img-upload-view/pkg/errors"
	"github.com/rs/zerolog"
)

type Middlewares[T any] struct {
	errorResponse  errres.ErrorResponse
	trustedOrigins []string
	logger         *zerolog.Logger
}

type Option[T any] func(*Middlewares[T])

func WithZerolog[T any](logger *zerolog.Logger) Option[T] {
	return func(m *Middlewares[T]) {
		m.logger = logger
	}
}

func WithErrorResponse[T any](errResponse errres.ErrorResponse) Option[T] {
	return func(m *Middlewares[T]) {
		m.errorResponse = errResponse
	}
}

func WithTrustedOrigins[T any](trustedOrigins []string) Option[T] {
	return func(m *Middlewares[T]) {
		m.trustedOrigins = trustedOrigins
	}
}

func New[T any](opts ...Option[T]) Middlewares[T] {
	m := Middlewares[T]{}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}
