package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"golog/internal/model"
	"golog/internal/store"
)

func newServer(t *testing.T) http.Handler {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	web := fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>GoLog</title>")}}
	return New(st, web, nil)
}

func do(t *testing.T, h http.Handler, method, url, body string, out any) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Host = "127.0.0.1:7341"
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if out != nil {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("%s %s: bad JSON %q", method, url, rec.Body.String())
		}
	}
	return rec
}

func TestCreateEditFlow(t *testing.T) {
	h := newServer(t)
	var p model.Project
	if rec := do(t, h, "POST", "/api/projects", `{"name":"  Field trip "}`, &p); rec.Code != 201 || p.Name != "Field trip" {
		t.Fatalf("create project: %d %+v", rec.Code, p)
	}
	var l model.Log
	if rec := do(t, h, "POST", "/api/projects/"+p.ID+"/logs", `{"name":"Log 1"}`, &l); rec.Code != 201 {
		t.Fatalf("create log: %d", rec.Code)
	}

	l.Beds = []model.Bed{{ID: "a", Thickness: 0.5, Lithology: []model.Lithology{{Pattern: "shale", Percent: 100}}}}
	body, _ := json.Marshal(l)
	var saved model.Log
	if rec := do(t, h, "PUT", "/api/logs/"+l.ID, string(body), &saved); rec.Code != 200 || saved.Revision != 2 || saved.Beds[0].Contact != "unknown" {
		t.Fatalf("save: %d %+v", rec.Code, saved)
	}

	// Saving the same old revision again is a conflict and returns the current log.
	var conflict struct {
		Current model.Log `json:"current"`
	}
	if rec := do(t, h, "PUT", "/api/logs/"+l.ID, string(body), &conflict); rec.Code != 409 || conflict.Current.Revision != 2 {
		t.Fatalf("conflict: %d %+v", rec.Code, conflict)
	}
}

func TestValidation(t *testing.T) {
	h := newServer(t)
	var p model.Project
	do(t, h, "POST", "/api/projects", `{"name":"P"}`, &p)
	var l model.Log
	do(t, h, "POST", "/api/projects/"+p.ID+"/logs", `{"name":"L"}`, &l)
	for name, bed := range map[string]string{
		"zero thickness": `{"id":"a","thickness":0}`,
		"bad contact":    `{"id":"a","thickness":1,"contact":"wobbly"}`,
		"over 100%":      `{"id":"a","thickness":1,"lithology":[{"pattern":"x","percent":60},{"pattern":"y","percent":60}]}`,
		"no id":          `{"thickness":1}`,
	} {
		body := `{"name":"L","revision":1,"units":"m","beds":[` + bed + `]}`
		if rec := do(t, h, "PUT", "/api/logs/"+l.ID, body, nil); rec.Code != 400 {
			t.Errorf("%s: status %d, want 400 (%s)", name, rec.Code, rec.Body.String())
		}
	}
	if rec := do(t, h, "POST", "/api/projects", `{"name":"   "}`, nil); rec.Code != 400 {
		t.Errorf("empty name: status %d", rec.Code)
	}
	if rec := do(t, h, "GET", "/api/logs/missing", "", nil); rec.Code != 404 {
		t.Errorf("missing log: status %d", rec.Code)
	}
}

func TestGuard(t *testing.T) {
	h := newServer(t)
	req := httptest.NewRequest("GET", "/api/projects", nil)
	req.Host = "evil.example:7341" // DNS rebinding
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Errorf("foreign Host: %d, want 403", rec.Code)
	}

	req = httptest.NewRequest("POST", "/api/projects", strings.NewReader(`{"name":"x"}`))
	req.Host = "127.0.0.1:7341"
	req.Header.Set("Content-Type", "text/plain") // a form post from another site
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 415 {
		t.Errorf("text/plain POST: %d, want 415", rec.Code)
	}

	req = httptest.NewRequest("POST", "/api/projects", strings.NewReader(`{"name":"x"}`))
	req.Host = "127.0.0.1:7341"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Errorf("foreign Origin: %d, want 403", rec.Code)
	}
}

func TestStaticFallback(t *testing.T) {
	h := newServer(t)
	rec := do(t, h, "GET", "/projects/abc", "", nil)
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "<title>GoLog") {
		t.Fatalf("app page: %d %q", rec.Code, rec.Body.String())
	}
}

func TestEmptyProjectHasLogList(t *testing.T) {
	h := newServer(t)
	var p model.Project
	do(t, h, "POST", "/api/projects", `{"name":"P"}`, &p)
	rec := do(t, h, "GET", "/api/projects/"+p.ID, "", nil)
	if !strings.Contains(rec.Body.String(), `"logs":[]`) {
		t.Fatalf("empty project should list logs as []: %s", rec.Body.String())
	}
}

func TestImport(t *testing.T) {
	h := newServer(t)
	csv := "THICKNESS (CM),LITHOLOGY,LITHOLOGY %,NOTES COLUMN\n50,Sandstone,100,first\n20,Shale,100,second\n"
	enc := func(s string) string { b, _ := json.Marshal([]byte(s)); return string(b) }

	var res struct {
		Format   string           `json:"format"`
		Content  model.LogContent `json:"content"`
		Warnings []string         `json:"warnings"`
		CSV      struct {
			Mapping map[string]int `json:"mapping"`
		} `json:"csv"`
	}
	if rec := do(t, h, "POST", "/api/import", `{"data":`+enc(csv)+`}`, &res); rec.Code != 200 || res.Format != "csv" || len(res.Content.Beds) != 2 {
		t.Fatalf("import: %d %s", rec.Code, rec.Body.String())
	}
	// Choosing the columns by hand: notes from column 0 is silly but allowed.
	if rec := do(t, h, "POST", "/api/import", `{"data":`+enc(csv)+`,"csv":{"mapping":{"thickness":0,"notes":1},"units":"m"}}`, &res); rec.Code != 200 || res.Content.Beds[0].Notes != "Sandstone" || res.Content.Beds[0].Thickness != 50 {
		t.Fatalf("import with mapping: %d %+v", rec.Code, res.Content.Beds)
	}
	golog := `{"golog":1,"kind":"log","name":"Back-up","content":{"units":"m","scale":20,"paper":"A3","palaeocurrent":"arrows","fill":"white","columns":[],"beds":[{"id":"x","thickness":2,"contact":"sharp","lithology":[{"pattern":"coal","percent":100}],"grain":{}}]}}`
	if rec := do(t, h, "POST", "/api/import", `{"data":`+enc(golog)+`}`, &res); rec.Code != 200 || res.Format != "golog" || res.Content.Scale != 20 || len(res.Content.Columns) != 10 {
		t.Fatalf("GoLog file: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, h, "POST", "/api/import", `{"data":`+enc(`{"golog":9,"kind":"log"}`)+`}`, nil); rec.Code != 422 || !strings.Contains(rec.Body.String(), "newer GoLog") {
		t.Errorf("file from a newer GoLog: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, h, "POST", "/api/import", `{"data":`+enc("just text")+`}`, nil); rec.Code != 422 {
		t.Errorf("not a SedLog file: %d", rec.Code)
	}

	// Create a log from the imported content.
	var p model.Project
	do(t, h, "POST", "/api/projects", `{"name":"P"}`, &p)
	do(t, h, "POST", "/api/import", `{"data":`+enc(csv)+`}`, &res)
	content, _ := json.Marshal(res.Content)
	var l model.Log
	if rec := do(t, h, "POST", "/api/projects/"+p.ID+"/logs", `{"name":"Imported","content":`+string(content)+`}`, &l); rec.Code != 201 || len(l.Beds) != 2 {
		t.Fatalf("create from import: %d %s", rec.Code, rec.Body.String())
	}
	var proj model.Project
	do(t, h, "GET", "/api/projects/"+p.ID, "", &proj)
	if proj.Logs[0].BedCount != 2 {
		t.Errorf("bed count in project list: %d", proj.Logs[0].BedCount)
	}
	bad := strings.Replace(string(content), `"scale":50`, `"scale":33`, 1)
	if rec := do(t, h, "POST", "/api/projects/"+p.ID+"/logs", `{"name":"Bad","content":`+bad+`}`, nil); rec.Code != 400 {
		t.Errorf("invalid content accepted: %d", rec.Code)
	}
}

func TestStatistics(t *testing.T) {
	h := newServer(t)
	c := model.NewLogContent()
	for i, f := range []string{"1", "2", "3", "4", "4", "4", "2", "5", "2"} {
		c.Beds = append(c.Beds, model.Bed{ID: f + string(rune('a'+i)), Thickness: []float64{1, 0.45, 0.45, 1.35, 0.65, 0.5, 0.6, 0.35, 0.65}[i], Facies: f})
	}
	body, _ := json.Marshal(map[string]any{"content": c, "shuffles": 300})
	var res struct {
		Facies struct {
			Codes []string `json:"codes"`
			M     struct {
				Observed float64 `json:"observed"`
				Shuffles int     `json:"shuffles"`
			} `json:"m"`
		} `json:"facies"`
		Runs struct {
			R struct {
				Observed float64 `json:"observed"`
			} `json:"r"`
		} `json:"runs"`
	}
	if rec := do(t, h, "POST", "/api/stats", string(body), &res); rec.Code != 200 || res.Facies.M.Observed != 0.5 || res.Facies.M.Shuffles != 300 || len(res.Facies.Codes) != 5 {
		t.Fatalf("stats: %d %s", rec.Code, rec.Body.String())
	}
	if r := res.Runs.R.Observed; r < 0.888 || r > 0.889 {
		t.Errorf("r = %v, want 8/9", r)
	}
	if rec := do(t, h, "POST", "/api/stats", `{"content":{"beds":[{"id":"a","thickness":-1}]}}`, nil); rec.Code != 400 {
		t.Errorf("a bad log: %d", rec.Code)
	}
}

func TestStop(t *testing.T) {
	h := newServer(t)
	if rec := do(t, h, "POST", "/api/quit", "", nil); rec.Code != 404 {
		t.Errorf("stop without a quit function: %d", rec.Code)
	}
	st, err := store.Open(filepath.Join(t.TempDir(), "q.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	stopped := make(chan bool, 1)
	h = New(st, fstest.MapFS{}, func() { stopped <- true })
	if rec := do(t, h, "POST", "/api/quit", "", nil); rec.Code != 202 {
		t.Fatalf("stop: %d", rec.Code)
	}
	<-stopped
	// Another website may not stop GoLog.
	req := httptest.NewRequest("POST", "/api/quit", nil)
	req.Host = "127.0.0.1:7341"
	req.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 403 {
		t.Errorf("stop from another website: %d", rec.Code)
	}
}
