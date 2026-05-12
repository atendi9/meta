package xhttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"strings"
)

// BasicAuth holds the credentials for HTTP Basic Authentication.
type BasicAuth struct {
	Username string
	Password string
}

// HTTPOptions defines the interface for HTTP request options, providing methods
// to retrieve headers, the request body, query parameters, cookies, context, and auth.
type HTTPOptions interface {
	// H returns a slice of HTTPData representing the headers.
	H() []HTTPData
	// B returns an [io.Reader] representing the request body.
	B() io.Reader
	// Q returns a slice of HTTPData representing the query parameters.
	Q() []HTTPData
	// C returns a slice of [http.Cookie] pointers representing the cookies.
	C() []*http.Cookie
	// Ctx returns the [context.Context] for managing request lifecycles.
	Ctx() context.Context
	// Auth returns the [BasicAuth] credentials for the request.
	Auth() *BasicAuth
}

// Options implements the [HTTPOptions] interface and holds headers,
// the request body, query parameters, cookies, context, and authentication.
type Options struct {
	Headers     []HTTPData
	Body        io.Reader
	QueryParams []HTTPData
	Cookies     []*http.Cookie
	Context     context.Context
	BasicAuth   *BasicAuth
}

// H returns the headers defined in the [Options] struct.
func (o *Options) H() []HTTPData {
	return o.Headers
}

// Q returns the query parameters defined in the [Options] struct.
func (o *Options) Q() []HTTPData {
	return o.QueryParams
}

// B returns the request body defined in the [Options] struct.
func (o *Options) B() io.Reader {
	return o.Body
}

// C returns the cookies defined in the [Options] struct.
func (o *Options) C() []*http.Cookie {
	return o.Cookies
}

// Ctx returns the context defined in the [Options] struct. If none is set, it returns nil.
func (o *Options) Ctx() context.Context {
	return o.Context
}

// Auth returns the [BasicAuth] credentials defined in the [Options] struct.
func (o *Options) Auth() *BasicAuth {
	return o.BasicAuth
}

// Header is an alias for [Data] representing an HTTP header.
type Header = Data

// QueryParam is an alias for [Data] representing an HTTP query parameter.
type QueryParam = Data

// NewData initializes or updates an [HTTPData] instance with the provided key and value.
func NewData(d HTTPData, key string, value any) {
	d.Set(key, value)
}

// SetQueryParams appends a list of [HTTPData] query parameters to the given URL string.
// It properly escapes the parameter keys and values and returns the complete URL string.
func SetQueryParams(
	url string,
	queryParams ...HTTPData,
) string {
	if len(queryParams) == 0 {
		return url
	}

	builder := new(strings.Builder)
	builder.WriteString(url)
	for i, param := range queryParams {
		key := netUrl.QueryEscape(param.Data().Key)
		value := netUrl.QueryEscape(fmt.Sprintf("%v", param.Data().Value))
		if i == 0 {
			builder.WriteString("?")
		} else {
			builder.WriteString("&")
		}
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(value)
	}

	return builder.String()
}

// SetHeaders applies a list of [HTTPData] headers to the provided [http.Request].
func SetHeaders(req *http.Request, headers []HTTPData) {
	for _, header := range headers {
		var value string
		if v, ok := header.Data().Value.(string); ok {
			value = v
		} else {
			value = fmt.Sprint(header.Data().Value)
		}
		req.Header.Set(header.Data().Key, value)
	}
}

// SetCookies applies a list of [http.Cookie] pointers to the provided [http.Request].
func SetCookies(req *http.Request, cookies []*http.Cookie) {
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
}

// SetBasicAuth applies Basic Authentication credentials to the provided [http.Request] if auth is not nil.
func SetBasicAuth(req *http.Request, auth *BasicAuth) {
	if auth != nil {
		req.SetBasicAuth(auth.Username, auth.Password)
	}
}

// HTTPData defines an interface for key-value pair data used in HTTP requests,
// such as headers or query parameters.
type HTTPData interface {
	// Set assigns a key and a value to the underlying data structure.
	Set(key string, value any)
	// Data retrieves the underlying [Data] representation.
	Data() Data
}

// Data represents a basic key-value pair used for headers and query parameters.
type Data struct {
	Key   string
	Value any
}

// Data returns a copy of the [Data] struct.
func (d *Data) Data() Data {
	return Data{
		Key:   d.Key,
		Value: d.Value,
	}
}

// Set assigns the given key and value to the [Data] instance.
func (d *Data) Set(key string, value any) {
	d.Key = key
	d.Value = value
}
