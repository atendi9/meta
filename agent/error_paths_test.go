package agent

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/atendi9/capivara/assert"
)

// badJSON returns a Configurator whose every response is a 200 carrying an
// invalid JSON body, driving each method into its decode-error branch.
func badJSON() *Configurator {
	return newTestConfigurator(mockOK(`not-json`))
}

// errReader is an io.Reader that always fails, used to force the multipart
// builder in FileRequest down its error path.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestDecodeErrors_AllowList(t *testing.T) {
	c := newTestClient(mockOK(`not-json`))
	_, err := c.AllowList.Get()
	assert.Error(t, err)
	_, err = c.AllowList.Add("551199")
	assert.Error(t, err)
}

func TestDecodeErrors_Settings(t *testing.T) {
	c := newTestClient(mockOK(`not-json`))
	_, err := c.Onboard.Settings()
	assert.Error(t, err)
	_, err = c.Onboard.UpdateSettings(SettingsRequest{})
	assert.Error(t, err)
}

func TestDecodeErrors_BusinessInfo(t *testing.T) {
	b := badJSON().BusinessInfo()
	_, err := b.Get()
	assert.Error(t, err)
	_, err = b.Update(BusinessInfoRequest{})
	assert.Error(t, err)
	_, err = b.Delete()
	assert.Error(t, err)
}

func TestDecodeErrors_FAQs(t *testing.T) {
	f := badJSON().FAQs()
	_, err := f.Create(FAQRequest{})
	assert.Error(t, err)
	_, err = f.List()
	assert.Error(t, err)
	_, err = f.Get("f1")
	assert.Error(t, err)
	_, err = f.Update("f1", FAQRequest{})
	assert.Error(t, err)
}

func TestDecodeErrors_Skills(t *testing.T) {
	s := badJSON().Skills()
	_, err := s.Create(SkillRequest{})
	assert.Error(t, err)
	_, err = s.List()
	assert.Error(t, err)
	_, err = s.Get("s1")
	assert.Error(t, err)
	_, err = s.Update("s1", SkillRequest{})
	assert.Error(t, err)
}

func TestDecodeErrors_Websites(t *testing.T) {
	w := badJSON().Websites()
	_, err := w.Create(WebsiteRequest{})
	assert.Error(t, err)
	_, err = w.List()
	assert.Error(t, err)
	_, err = w.Get("w1")
	assert.Error(t, err)
	_, err = w.Update("w1", WebsiteRequest{})
	assert.Error(t, err)
}

func TestDecodeErrors_Files(t *testing.T) {
	f := badJSON().Files()
	_, err := f.Upload(FileRequest{FileName: "doc.pdf", File: strings.NewReader("hi")})
	assert.Error(t, err)
	_, err = f.List()
	assert.Error(t, err)
	_, err = f.Get("file_1")
	assert.Error(t, err)
}

func TestDecodeErrors_Connectors(t *testing.T) {
	c := badJSON().Connectors()
	_, err := c.Create(ConnectorRequest{})
	assert.Error(t, err)
	_, err = c.List()
	assert.Error(t, err)
	_, err = c.Get("c1")
	assert.Error(t, err)
	_, err = c.Update("c1", ConnectorRequest{})
	assert.Error(t, err)
	_, err = c.UpsertAPIKey("c1", UpsertAPIKeyRequest{})
	assert.Error(t, err)
	_, err = c.UpsertOAuth("c1", UpsertOAuthRequest{})
	assert.Error(t, err)
	_, err = c.UpsertCertificate("c1", CertificateRequest{})
	assert.Error(t, err)
	_, err = c.Logs("c1", ConnectorLogsOptions{})
	assert.Error(t, err)
}

func TestDecodeErrors_ConnectorTools(t *testing.T) {
	tools := badJSON().Connectors().Tools("c1")
	_, err := tools.Create(ConnectorToolRequest{})
	assert.Error(t, err)
	_, err = tools.List()
	assert.Error(t, err)
	_, err = tools.Get("t1")
	assert.Error(t, err)
	_, err = tools.Update("t1", ConnectorToolRequest{})
	assert.Error(t, err)
	_, err = tools.Run("t1", ToolRunRequest{})
	assert.Error(t, err)
}

func TestDecodeErrors_AgentEvent(t *testing.T) {
	ev := badJSON().Operate().AgentEvent()
	_, err := ev.Send(AgentEventRequest{})
	assert.Error(t, err)
}

func TestDecodeErrors_AgentEval(t *testing.T) {
	ev := badJSON().Operate().AgentEval()
	_, err := ev.Details("e1")
	assert.Error(t, err)
	_, err = ev.Summary("s1")
	assert.Error(t, err)
	_, err = ev.RunStatus("job_1")
	assert.Error(t, err)
	_, err = ev.Run("c1")
	assert.Error(t, err)
}

func TestFileRequest_BuildError(t *testing.T) {
	// A file whose contents fail to read propagates the error through Read,
	// ContentType and Upload.
	_, err := io.ReadAll(&FileRequest{FileName: "doc.pdf", File: errReader{}})
	assert.Error(t, err)

	_, err = (&FileRequest{FileName: "doc.pdf", File: errReader{}}).ContentType()
	assert.Error(t, err)

	f := newTestConfigurator(mockOK(`{}`)).Files()
	_, err = f.Upload(FileRequest{FileName: "doc.pdf", File: errReader{}})
	assert.Error(t, err)
}
