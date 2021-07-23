package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	adminToken  = "a@b.com"
	memberToken = "test@unittest.com"
)

var svr *devserver

func init() {
	svr = &devserver{
		db: make(map[string][]map[string]interface{}),
	}

	// create an admin user
	users := make([]map[string]interface{}, 0)
	newUser := make(map[string]interface{})
	newUser["accountId"] = 1
	newUser["userId"] = 1
	newUser["email"] = adminToken
	newUser["password"] = "test123"
	newUser["role"] = 100

	member := make(map[string]interface{})
	member["accountId"] = 2
	member["userId"] = 2
	member["email"] = memberToken
	member["password"] = "test123"
	member["role"] = 0

	users = append(users, newUser)
	users = append(users, member)
	svr.db["sb_users"] = users
}

func newRecorder(t *testing.T, method, url, typ string, body io.Reader, h http.HandlerFunc) *httptest.ResponseRecorder {
	return newAuthRecorder(t, method, url, typ, "", body, h)
}

func newAuthRecorder(t *testing.T, method, url, typ, tok string, body io.Reader, h http.HandlerFunc) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", typ)
	if len(tok) > 0 {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	handler := http.HandlerFunc(h)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	return rec
}

func TestDevServerUserRegisterLogin(t *testing.T) {
	data := strings.NewReader(`{"email": "unit@test.com", "password": "pw123"}`)
	rec := newRecorder(t, "POST", "/register", "application/json", data, svr.register)

	if rec.Code != http.StatusOK {
		t.Errorf("[register] got status %d, expected %d", rec.Code, http.StatusOK)
	} else if body := rec.Body.String(); body != `"unit@test.com"` {
		t.Errorf("[register] got body %s, expected unit@test.com", body)
	}

	logData := strings.NewReader(`{"email": "unit@test.com", "password": "pw123"}`)
	logRec := newRecorder(t, "POST", "/login", "application/json", logData, svr.login)

	if logRec.Code != http.StatusOK {
		t.Errorf("[login] got status %d, expected %d -> %s", logRec.Code, http.StatusOK, logRec.Body.String())
	} else if body := logRec.Body.String(); body != `"unit@test.com"` {
		t.Errorf("[login] got body %s, expected unit@test.com", body)
	}
}

func TestDevServerDatabase(t *testing.T) {
	data := strings.NewReader(`{"email": "db@unittest.com", "password": "pw123"}`)
	rec := newRecorder(t, "POST", "/register", "application/json", data, svr.register)

	if rec.Code != http.StatusOK {
		t.Errorf("[register] got status %d, expected %d", rec.Code, http.StatusOK)
	} else if body := rec.Body.String(); body != `"db@unittest.com"` {
		t.Errorf("[register] got body %s, expected db@unittest.com", body)
	}

	tok := "db@unittest.com"

	cData := strings.NewReader(`{"my": "unit test", "works": 123}`)
	cRec := newAuthRecorder(t, "POST", "/db/unittest", "application/json", tok, cData, svr.database)

	if cRec.Code != http.StatusCreated {
		t.Errorf("[create] got status %d, expected %d", cRec.Code, http.StatusCreated)
	} else if x := len(svr.db["unittest"]); x != 1 {
		t.Errorf("[create] got %d table len, expected 1", x)
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(cRec.Body.Bytes(), &doc); err != nil {
		t.Error(err)
	}

	fRec := newAuthRecorder(t, "GET", "/db/unittest/1", "application/json", tok, nil, svr.database)

	if fRec.Code != http.StatusOK {
		t.Errorf("[fetch] got status %d, expected %d", fRec.Code, http.StatusOK)
	} else if body := fRec.Body.String(); strings.Index(body, `"my":"unit test"`) < 0 {
		t.Errorf(`[fetch] got %s as body, expected "my": "unit test"`, body)
	}

	uData := strings.NewReader(`{"my": "updated value", "works": 654}`)
	uRec := newAuthRecorder(t, "PUT", "/db/unittest/1", "application/json", tok, uData, svr.database)

	if uRec.Code != http.StatusOK {
		t.Errorf("[update] got status %d, expected %d", uRec.Code, http.StatusOK)
	} else {
		cond := func(v map[string]interface{}) bool {
			return v["accountId"] == fmt.Sprintf("%v", doc["accountId"])
		}
		tmp, err := svr.fetch("unittest", fmt.Sprintf("%v", doc["id"]), cond)
		if err != nil {
			t.Error(err)
		} else if tmp["my"] != "updated value" {
			t.Errorf("[update] got my %v, expected updated value", tmp["my"])
		}
	}

	lRec := newAuthRecorder(t, "GET", "/db/unittest", "application/json", tok, nil, svr.database)

	if lRec.Code != http.StatusOK {
		t.Errorf("[list] got status %d, expected %d", lRec.Code, http.StatusOK)
	}

	dRec := newAuthRecorder(t, "DELETE", "/db/unittest/1", "application/json", tok, nil, svr.database)

	if dRec.Code != http.StatusOK {
		t.Errorf("[delete] got status %d, expected %d", dRec.Code, http.StatusOK)
	} else if x := len(svr.db["unittest"]); x != 0 {
		t.Errorf("[delete] table len %d, expected to be empty", x)
	}
}

type permtest struct {
	status int
	action string
	doc    map[string]interface{}
}

func TestDatabasePermission(t *testing.T) {
	doc := strings.NewReader(`{"title": "sample document", "done": false}`)

	// testing public repo
	// public repo can be read from all, written from owner
	rec := newAuthRecorder(t, "POST", "/db/pub_repo", "application/json", memberToken, doc, svr.database)

	if rec.Code > 299 {
		t.Fatalf("cannot add test document: %v", rec.Code)
	}

	eRec := newRecorder(t, "GET", "/db/pub_repo", "application/json", nil, svr.database)
	if eRec.Code > 299 {
		t.Fatalf("public user cannot read public repo: %v", eRec.Code)
	}

	wRec := newRecorder(t, "POST", "/db/pub_repo", "application/json", doc, svr.database)
	if wRec.Code < 300 {
		t.Fatal("public user could write on repo")
	}

	// testing auth can read, account can write
	doc = strings.NewReader(`{"title": "sample document", "done": false}`)
	rec = newAuthRecorder(t, "POST", "/db/repo_764_", "application/json", adminToken, doc, svr.database)

	if rec.Code > 299 {
		t.Fatalf("cannot add test document: %d %s", rec.Code, rec.Body.String())
	}

	eRec = newAuthRecorder(t, "GET", "/db/repo_764_", "application/json", memberToken, nil, svr.database)
	if eRec.Code > 299 {
		t.Fatalf("public user cannot read public repo: %v", eRec.Code)
	}

	wRec = newAuthRecorder(t, "PUT", "/db/repo_764_/1", "application/json", memberToken, doc, svr.database)
	if wRec.Code < 300 {
		t.Fatal("public user could write on repo")
	}

	doc = strings.NewReader(`{"title": "sample document", "done": true}`)
	wRec = newAuthRecorder(t, "PUT", "/db/repo_764_/1", "application/json", adminToken, doc, svr.database)
	if wRec.Code > 299 {
		t.Fatal("public user could write on repo")
	} else if strings.Index(wRec.Body.String(), "true") == -1 {
		t.Errorf("expected to have true got %s", wRec.Body.String())
	}
}

func TestDevServerPostForm(t *testing.T) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	field1, err := w.CreateFormField("field1")
	if err != nil {
		t.Error(err)
	}
	field1.Write([]byte("field1 value"))

	field2, err := w.CreateFormField("field2")
	if err != nil {
		t.Error(err)
	}
	field2.Write([]byte("another value"))

	w.Close()

	rec := newRecorder(t, "POST", "/postform/form-unittest", w.FormDataContentType(), bytes.NewReader(b.Bytes()), svr.postForm)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, expected %d -> %s", rec.Code, http.StatusOK, rec.Body.String())
	} else if x := len(svr.db["sb_forms"]); x != 1 {
		t.Errorf("got sb_forms len of %d, expected 1", x)
	}

}
