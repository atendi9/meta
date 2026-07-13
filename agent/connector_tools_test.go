package agent

import (
	"io"
	"net/http"
	"testing"

	"github.com/atendi9/capivara/assert"
)

func TestConnectorTools_Create_Success(t *testing.T) {
	mock := mockOK(`{"id":"t1","name":"check_order","description":"look up an order"}`)
	tools := newTestConfigurator(mock).Connectors().Tools("c1")

	tool, err := tools.Create(ConnectorToolRequest{
		Name:        "check_order",
		Description: "look up an order",
		RequestDefinition: ToolRequestDefinition{
			Method: http.MethodGet,
			Path:   "/orders/{id}",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, "t1", tool.ID)
	assert.Equal(t, "check_order", tool.Name)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1/tools", mock.Calls[0].URL)
}

func TestConnectorTools_List_Success(t *testing.T) {
	mock := mockOK(`{"root":[{"id":"t1","name":"check_order","description":"d"}]}`)
	tools := newTestConfigurator(mock).Connectors().Tools("c1")

	list, err := tools.List()

	assert.NoError(t, err)
	assert.LengthSlice(t, 1, list)
	assert.Equal(t, "t1", list[0].ID)
	assert.Equal(t, http.MethodGet, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1/tools", mock.Calls[0].URL)
}

func TestConnectorTools_Get_Success(t *testing.T) {
	mock := mockOK(`{"id":"t1","name":"check_order","description":"d"}`)
	tools := newTestConfigurator(mock).Connectors().Tools("c1")

	tool, err := tools.Get("t1")

	assert.NoError(t, err)
	assert.Equal(t, "t1", tool.ID)
	assert.Equal(t, testBase+"/agent_connectors/c1/tools/t1", mock.Calls[0].URL)
}

func TestConnectorTools_Update_Success(t *testing.T) {
	mock := mockOK(`{"id":"t1","name":"check_order2","description":"d"}`)
	tools := newTestConfigurator(mock).Connectors().Tools("c1")

	tool, err := tools.Update("t1", ConnectorToolRequest{Name: "check_order2", Description: "d"})

	assert.NoError(t, err)
	assert.Equal(t, "check_order2", tool.Name)
	assert.Equal(t, http.MethodPut, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1/tools/t1", mock.Calls[0].URL)
}

func TestConnectorTools_Delete_Success(t *testing.T) {
	mock := mockOK(`{}`)
	tools := newTestConfigurator(mock).Connectors().Tools("c1")

	err := tools.Delete("t1")

	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1/tools/t1", mock.Calls[0].URL)
}

func TestConnectorTools_Run_Success(t *testing.T) {
	mock := mockOK(`{"output":"{\"ok\":true}","status":"success"}`)
	tools := newTestConfigurator(mock).Connectors().Tools("c1")

	res, err := tools.Run("t1", ToolRunRequest{Input: `{"id":"A1"}`})

	assert.NoError(t, err)
	assert.Equal(t, "success", res.Status)
	assert.Equal(t, http.MethodPost, mock.Calls[0].Method)
	assert.Equal(t, testBase+"/agent_connectors/c1/tools/t1/run", mock.Calls[0].URL)

	sent, _ := io.ReadAll(mock.Calls[0].Options.B())
	assert.True(t, contains(string(sent), `"input"`))
}

func TestConnectorTools_TransportErrors(t *testing.T) {
	tools := newTestConfigurator(mockErr(io.ErrClosedPipe)).Connectors().Tools("c1")

	_, err := tools.Create(ConnectorToolRequest{})
	assert.Error(t, err)
	_, err = tools.List()
	assert.Error(t, err)
	_, err = tools.Get("t1")
	assert.Error(t, err)
	_, err = tools.Update("t1", ConnectorToolRequest{})
	assert.Error(t, err)
	err = tools.Delete("t1")
	assert.Error(t, err)
	_, err = tools.Run("t1", ToolRunRequest{})
	assert.Error(t, err)
}
