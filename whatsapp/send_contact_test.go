package whatsapp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
	"github.com/atendi9/meta"
	"github.com/atendi9/meta/xhttp"
)

func TestSendContact_Success(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages":[{"id": "wamid_contact_123"}]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient:  mockClient,
			ApiVersion:  "v19.0",
			BaseUrl:     "https://graph.facebook.com",
			AccessToken: "valid_token",
		},
	}

	receiver := "5511999999999"
	name := "John Doe da Silva"
	phone := "+55 11 98888-8888"

	id, err := api.SendContact(receiver, name, phone)
	assert.NoError(t, err)
	assert.Equal(t, "wamid_contact_123", id)
	assert.LengthSlice(t, 1, mockClient.Calls)
	assert.Equal(t, http.MethodPost, mockClient.Calls[0].Method)
}

func TestSendContact_InvalidName(t *testing.T) {
	api := Default("3333333333", "api_token")
	api.HttpClient = xhttp.NewMockClient(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"Hello World"}`))),
	}, nil)
	id, err := api.SendContact("123", "John", "555-5555")

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidContactName, err)
	assert.Equal(t, "", id)
}

func TestSendContact_RequestError(t *testing.T) {
	mockClient := xhttp.NewMockClient(nil, io.ErrUnexpectedEOF)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	id, err := api.SendContact("5511999999999", "John Doe", "555-0000")

	assert.Error(t, err)
	assert.Equal(t, "", id)
}

func TestSendContact_EmptyResponse(t *testing.T) {
	mockRes := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"messages":[]}`)),
	}
	mockClient := xhttp.NewMockClient(mockRes, nil)

	api := &Client{
		senderID: "10987654321",
		GraphAPIClient: meta.GraphAPIClient{
			HttpClient: mockClient,
		},
	}

	id, err := api.SendContact("5511999999999", "John Doe", "555-1234")
	assert.Equal(t, ErrMessageNotSent, err)
	assert.Equal(t, "", id)
}

func TestSplitContactName(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Empty string", "", nil},
		{"Only spaces", "   ", nil},
		{"Standard name", "John Doe", []string{"John", "Doe"}},
		{"Multiple spaces", " John   Doe ", []string{"John", "Doe"}},
		{"Camel case name", "JohnDoe", []string{"John", "Doe"}},
		{"Single word", "John", []string{"John"}},
		{"Camel case multiple words", "JohnDoeSmith", []string{"John", "Doe", "Smith"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitContactName(tc.input)
			assert.Equal(t, fmt.Sprint(tc.expected), fmt.Sprint(result))
		})
	}
}

func TestSplitCamelCase(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Empty string", "", nil},
		{"Standard camel case", "JohnDoe", []string{"John", "Doe"}},
		{"Multiple camel case words", "JoaoSilvaSauro", []string{"Joao", "Silva", "Sauro"}},
		{"All lowercase", "john", []string{"john"}},
		{"All uppercase", "JOHN", []string{"JOHN"}},
		{"Single character", "A", []string{"A"}},
		{"Starting with lowercase", "johnDoe", []string{"john", "Doe"}},
		{"Mixed with consecutive uppercase at end", "ParseXML", []string{"Parse", "XML"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitCamelCase(tc.input)
			assert.Equal(t, fmt.Sprint(tc.expected), fmt.Sprint(result))
		})
	}
}

func TestSplitSnakeCase(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected []string
	}{
		{"Empty string", "", nil},
		{"Only underscores", "___", nil},
		{"Standard snake case", "john_doe", []string{"john", "doe"}},
		{"Multiple snake case words", "joao_silva_sauro", []string{"joao", "silva", "sauro"}},
		{"Consecutive underscores", "john__doe", []string{"john", "doe"}},
		{"Leading and trailing underscores", "_john_doe_", []string{"john", "doe"}},
		{"Mixed case", "John_Doe", []string{"John", "Doe"}},
		{"Single word without underscores", "john", []string{"john"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := splitSnakeCase(tc.input)
			assert.Equal(t, fmt.Sprint(tc.expected), fmt.Sprint(result))
		})
	}
}
