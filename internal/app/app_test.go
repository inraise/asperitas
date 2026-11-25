package app_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/inraise/asperitas/internal/app"
	"github.com/inraise/asperitas/internal/domain"
)

var secret = []byte("test-secret-test-secret-test-secret")

type api struct {
	t    *testing.T
	base string
}

func newAPI(t *testing.T) *api {
	t.Helper()
	dir := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(dir, "html"), 0o755))
	for name, content := range map[string]string{
		"index.html":    "<html>app</html>",
		"favicon.ico":   "ico",
		"manifest.json": "{}",
	} {
		must(t, os.WriteFile(filepath.Join(dir, "html", name), []byte(content), 0o644))
	}
	srv := httptest.NewServer(app.New(secret, dir))
	t.Cleanup(srv.Close)
	return &api{t: t, base: srv.URL}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func (a *api) do(method, path, token string, body any) (int, []byte) {
	a.t.Helper()
	var rd io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		rd = strings.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		must(a.t, err)
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, a.base+path, rd)
	must(a.t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	must(a.t, err)
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	must(a.t, err)
	return resp.StatusCode, raw
}

func (a *api) expect(want int, method, path, token string, body any) []byte {
	a.t.Helper()
	got, raw := a.do(method, path, token, body)
	if got != want {
		a.t.Fatalf("%s %s: status %d, want %d; body=%s", method, path, got, want, raw)
	}
	return raw
}

func (a *api) register(name string) string {
	a.t.Helper()
	raw := a.expect(201, "POST", "/api/register", "", map[string]string{"username": name, "password": "example-password"})
	var r domain.AuthenticationResponse
	must(a.t, json.Unmarshal(raw, &r))
	return r.Token
}

func (a *api) post(want int, method, path, token string, body any) domain.Post {
	a.t.Helper()
	var p domain.Post
	must(a.t, json.Unmarshal(a.expect(want, method, path, token, body), &p))
	return p
}

func (a *api) list(path string) []domain.Post {
	a.t.Helper()
	raw := a.expect(200, "GET", path, "", nil)
	if strings.TrimSpace(string(raw)) == "null" {
		a.t.Fatalf("%s returned null instead of []", path)
	}
	var ps []domain.Post
	must(a.t, json.Unmarshal(raw, &ps))
	return ps
}

func (a *api) textPost(token, title, category string) domain.Post {
	a.t.Helper()
	return a.post(201, "POST", "/api/posts", token, map[string]string{
		"title": title, "category": category, "type": "text", "text": "some text",
	})
}

func message(t *testing.T, raw []byte) string {
	t.Helper()
	var m struct {
		Message string `json:"message"`
	}
	must(t, json.Unmarshal(raw, &m))
	return m.Message
}

func validationParams(t *testing.T, raw []byte) []string {
	t.Helper()
	var v struct {
		Errors []struct{ Param, Msg string } `json:"errors"`
	}
	must(t, json.Unmarshal(raw, &v))
	if len(v.Errors) == 0 {
		t.Fatalf("expected non-empty errors array, got %s", raw)
	}
	out := []string{}
	for _, e := range v.Errors {
		out = append(out, e.Param)
	}
	return out
}

func signedToken(t *testing.T, method jwt.SigningMethod, key any, user domain.User, exp time.Time) string {
	t.Helper()
	claims := domain.JwtClaims{User: user, RegisteredClaims: jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		ExpiresAt: jwt.NewNumericDate(exp),
	}}
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	must(t, err)
	return s
}

func TestWebRouting(t *testing.T) {
	a := newAPI(t)
	for _, p := range []string{"/", "/login", "/signup", "/createpost", "/a/music", "/a/music/abc123", "/u/alice", "/favicon.ico", "/manifest.json"} {
		if code, _ := a.do("GET", p, "", nil); code != 200 {
			t.Errorf("GET %s = %d, want 200", p, code)
		}
	}
	code, raw := a.do("GET", "/api/nope", "", nil)
	if code != 404 || message(t, raw) == "" {
		t.Errorf("unknown /api path: %d %s, want JSON 404", code, raw)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	a := newAPI(t)
	token := a.register("alice")

	parts := strings.Split(token, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	must(t, err)
	var claims struct {
		User     map[string]any `json:"user"`
		Iat, Exp int64
	}
	must(t, json.Unmarshal(payload, &claims))
	if claims.User["username"] != "alice" || claims.User["id"] == "" || claims.Exp-claims.Iat != 7*24*3600 {
		t.Errorf("unexpected claims: %s", payload)
	}
	if bytes.Contains(payload, []byte("assword")) || bytes.Contains(payload, []byte("hash")) {
		t.Errorf("secrets in JWT payload: %s", payload)
	}

	for _, name := range []string{"alice", "ALICE"} {
		raw := a.expect(422, "POST", "/api/register", "", map[string]string{"username": name, "password": "example-password"})
		if p := validationParams(t, raw); p[0] != "username" {
			t.Errorf("duplicate %q: params %v", name, p)
		}
	}

	a.expect(400, "POST", "/api/register", "", "{not json")
	for name, body := range map[string]map[string]string{
		"short password":   {"username": "bob", "password": "short"},
		"bad username":     {"username": "bo b", "password": "example-password"},
		"empty username":   {"username": "", "password": "example-password"},
		"73-byte password": {"username": "bob", "password": strings.Repeat("a", 73)},
		"2-byte chars >72": {"username": "bob", "password": strings.Repeat("я", 37)}, // 37 символов = 74 байта
	} {
		raw := a.expect(422, "POST", "/api/register", "", body)
		if bytes.Contains(raw, []byte(body["password"])) && body["password"] != "" {
			t.Errorf("%s: password echoed in error", name)
		}
	}

	a.expect(200, "POST", "/api/login", "", map[string]string{"username": "Alice", "password": "example-password"})
	raw := a.expect(401, "POST", "/api/login", "", map[string]string{"username": "alice", "password": "wrong-password"})
	if m := message(t, raw); m != "invalid password" {
		t.Errorf("wrong password message = %q", m)
	}
	raw = a.expect(401, "POST", "/api/login", "", map[string]string{"username": "nobody", "password": "example-password"})
	if m := message(t, raw); m != "user not found" {
		t.Errorf("unknown user message = %q", m)
	}
}

func TestAuthMiddlewareRejectsBadTokens(t *testing.T) {
	a := newAPI(t)
	u := domain.User{Id: "u1", Username: "mallory"}
	valid := signedToken(t, jwt.SigningMethodHS256, secret, u, time.Now().Add(time.Hour))
	body := map[string]string{"title": "t", "category": "c", "type": "text", "text": "some text"}

	bad := map[string]string{
		"no token":     "",
		"garbage":      "abc.def.ghi",
		"expired":      signedToken(t, jwt.SigningMethodHS256, secret, u, time.Now().Add(-time.Minute)),
		"wrong secret": signedToken(t, jwt.SigningMethodHS256, []byte("another-secret-another-secret-123"), u, time.Now().Add(time.Hour)),
		"HS512":        signedToken(t, jwt.SigningMethodHS512, secret, u, time.Now().Add(time.Hour)),
		"alg none":     signedToken(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, u, time.Now().Add(time.Hour)),
		"tampered":     valid[:len(valid)-2] + "xx",
	}
	for name, tok := range bad {
		raw := a.expect(401, "POST", "/api/posts", tok, body)
		if message(t, raw) != "unauthorized" {
			t.Errorf("%s: body %s", name, raw)
		}
	}

	if n := len(a.list("/api/posts")); n != 0 {
		t.Fatalf("anonymous request created %d posts", n)
	}
	a.expect(201, "POST", "/api/posts", valid, body) // валидный токен проходит
}

func TestCreatePost(t *testing.T) {
	a := newAPI(t)
	tok := a.register("alice")

	p := a.textPost(tok, "Learning Go", "programming")
	if p.Id == "" || p.Author.Username != "alice" || p.Score != 1 || p.UpvotePercentage != 100 || p.Views != 0 ||
		len(p.Votes) != 1 || p.Votes[0].User != p.Author.Id || p.Votes[0].Vote != 1 || p.Comments == nil || len(p.Comments) != 0 ||
		p.Text == nil || p.Url != nil {
		t.Fatalf("unexpected post: %+v", p)
	}

	forged := a.post(201, "POST", "/api/posts", tok, map[string]any{
		"title": "x", "category": "c", "type": "text", "text": "some text",
		"author": map[string]string{"id": "evil", "username": "evil"},
	})
	if forged.Author.Username != "alice" {
		t.Fatalf("author taken from body: %+v", forged.Author)
	}

	l := a.post(201, "POST", "/api/posts/", tok, map[string]string{
		"title": "Go", "category": "programming", "type": "link", "url": "https://go.dev/", "text": "stale",
	})
	if l.Url == nil || *l.Url != "https://go.dev/" || l.Text != nil {
		t.Fatalf("link post: %+v", l)
	}

	invalid := map[string]map[string]string{
		"title":    {"title": "", "category": "c", "type": "text", "text": "some text"},
		"category": {"title": "t", "category": "", "type": "text", "text": "some text"},
		"type":     {"title": "t", "category": "c", "type": "video"},
		"text":     {"title": "t", "category": "c", "type": "text", "text": "abc"},
		"url":      {"title": "t", "category": "c", "type": "link", "url": "go.dev"},
	}
	for param, body := range invalid {
		raw := a.expect(422, "POST", "/api/posts", tok, body)
		if got := validationParams(t, raw); got[0] != param {
			t.Errorf("%s: params %v", param, got)
		}
	}
	a.expect(422, "POST", "/api/posts", tok, map[string]string{"title": " padded", "category": "c", "type": "text", "text": "some text"})
	a.expect(422, "POST", "/api/posts", tok, map[string]string{"title": strings.Repeat("a", 101), "category": "c", "type": "text", "text": "some text"})
	a.expect(422, "POST", "/api/posts", tok, map[string]string{"title": "t", "category": "c", "type": "link", "url": "javascript://x.com/%0Aalert(1)"})
	a.expect(400, "POST", "/api/posts", tok, "{")
}

func TestListsAndViews(t *testing.T) {
	a := newAPI(t)
	if got := a.list("/api/posts"); len(got) != 0 {
		t.Fatalf("empty feed: %v", got)
	}
	tok := a.register("alice")
	other := a.register("bob")

	first := a.textPost(tok, "first", "music")
	second := a.textPost(tok, "second", "funny")
	third := a.textPost(other, "third", "music")

	feed := a.list("/api/posts/")
	if len(feed) != 3 || feed[0].Id != third.Id || feed[2].Id != first.Id {
		t.Fatalf("feed order: %v", ids(feed))
	}
	a.post(200, "GET", "/api/post/"+first.Id+"/upvote", other, nil)
	if feed = a.list("/api/posts"); feed[0].Id != first.Id {
		t.Fatalf("score order: %v", ids(feed))
	}

	if got := a.list("/api/posts/music"); len(got) != 2 {
		t.Fatalf("category music: %v", ids(got))
	}
	if got := a.list("/api/posts/nothing"); len(got) != 0 {
		t.Fatalf("empty category: %v", got)
	}

	mine := a.list("/api/user/ALICE")
	if len(mine) != 2 || mine[0].Id != second.Id || mine[1].Id != first.Id {
		t.Fatalf("user posts: %v", ids(mine))
	}
	a.register("carol")
	if got := a.list("/api/user/carol"); len(got) != 0 {
		t.Fatalf("user without posts: %v", got)
	}
	raw := a.expect(404, "GET", "/api/user/ghost", "", nil)
	if message(t, raw) != "user not found" {
		t.Errorf("unknown user: %s", raw)
	}
	a.expect(400, "GET", "/api/user/bad%20name", "", nil)

	for want := 1; want <= 2; want++ {
		if p := a.post(200, "GET", "/api/post/"+second.Id, "", nil); p.Views != want {
			t.Fatalf("views = %d, want %d", p.Views, want)
		}
		a.list("/api/posts")
	}
	raw = a.expect(404, "GET", "/api/post/missing", "", nil)
	if message(t, raw) != "post not found" {
		t.Errorf("missing post: %s", raw)
	}
}

func ids(ps []domain.Post) []string {
	out := []string{}
	for _, p := range ps {
		out = append(out, p.Title)
	}
	return out
}

func TestDeletePost(t *testing.T) {
	a := newAPI(t)
	alice, bob := a.register("alice"), a.register("bob")
	p := a.textPost(alice, "mine", "c")
	path := "/api/post/" + p.Id

	a.expect(401, "DELETE", path, "", nil)
	a.expect(401, "DELETE", path, bob, nil)
	a.post(200, "GET", path, "", nil)

	raw := a.expect(200, "DELETE", path, alice, nil)
	if message(t, raw) != "success" {
		t.Errorf("delete body: %s", raw)
	}
	a.expect(404, "GET", path, "", nil)
	if n := len(a.list("/api/posts")); n != 0 {
		t.Fatalf("post still in feed")
	}
	a.expect(404, "DELETE", path, alice, nil)

	p2 := a.textPost(alice, "other", "c")
	admin := signedToken(t, jwt.SigningMethodHS256, secret, domain.User{Id: "admin1", Username: "root", Admin: true}, time.Now().Add(time.Hour))
	a.expect(200, "DELETE", "/api/post/"+p2.Id, admin, nil)
}

func TestComments(t *testing.T) {
	a := newAPI(t)
	alice, bob, carol := a.register("alice"), a.register("bob"), a.register("carol")
	p := a.textPost(alice, "post", "c")
	path := "/api/post/" + p.Id

	a.expect(401, "POST", path, "", map[string]string{"comment": "hi"})
	a.expect(404, "POST", "/api/post/missing", bob, map[string]string{"comment": "hi"})
	a.expect(400, "POST", path, bob, "{")
	a.expect(422, "POST", path, bob, map[string]string{"comment": ""})
	a.expect(422, "POST", path, bob, map[string]string{"comment": strings.Repeat("a", 2001)})

	withComment := a.post(201, "POST", path, bob, map[string]string{"comment": "Thanks for sharing!"})
	if len(withComment.Comments) != 1 {
		t.Fatalf("comments: %+v", withComment.Comments)
	}
	c := withComment.Comments[0]
	if c.Body != "Thanks for sharing!" || c.Author.Username != "bob" || c.Id == "" || c.Created == "" {
		t.Fatalf("comment: %+v", c)
	}
	if withComment.Id != p.Id || withComment.Score != 1 {
		t.Fatalf("response is not the full post: %+v", withComment)
	}

	cpath := path + "/" + c.Id
	a.expect(401, "DELETE", cpath, "", nil)
	a.expect(401, "DELETE", cpath, carol, nil)
	a.expect(404, "DELETE", path+"/nope", bob, nil)
	a.expect(404, "DELETE", "/api/post/missing/"+c.Id, bob, nil)

	after := a.post(200, "DELETE", cpath, bob, nil)
	if after.Comments == nil || len(after.Comments) != 0 {
		t.Fatalf("after delete: %+v", after.Comments)
	}

	c2 := a.post(201, "POST", path, carol, map[string]string{"comment": "second"}).Comments[0]
	a.post(200, "DELETE", path+"/"+c2.Id, alice, nil)

	c3 := a.post(201, "POST", path, carol, map[string]string{"comment": "third"}).Comments[0]
	admin := signedToken(t, jwt.SigningMethodHS256, secret, domain.User{Id: "admin1", Username: "root", Admin: true}, time.Now().Add(time.Hour))
	a.post(200, "DELETE", path+"/"+c3.Id, admin, nil)
}

func TestVotes(t *testing.T) {
	a := newAPI(t)
	alice, bob, carol := a.register("alice"), a.register("bob"), a.register("carol")
	p := a.textPost(alice, "post", "c")
	path := "/api/post/" + p.Id

	for _, action := range []string{"upvote", "downvote", "unvote"} {
		a.expect(401, "GET", path+"/"+action, "", nil)
		a.expect(404, "GET", "/api/post/missing/"+action, bob, nil)
	}

	check := func(got domain.Post, score, pct, nvotes int) {
		t.Helper()
		if got.Score != score || got.UpvotePercentage != pct || len(got.Votes) != nvotes || got.Votes == nil {
			t.Fatalf("score=%d pct=%d votes=%d; want %d/%d/%d", got.Score, got.UpvotePercentage, len(got.Votes), score, pct, nvotes)
		}
	}

	check(a.post(200, "GET", path+"/upvote", bob, nil), 2, 100, 2)
	check(a.post(200, "GET", path+"/upvote", bob, nil), 2, 100, 2)
	check(a.post(200, "GET", path+"/downvote", carol, nil), 1, 66, 3)
	check(a.post(200, "GET", path+"/downvote", bob, nil), -1, 33, 3)
	check(a.post(200, "GET", path+"/unvote", bob, nil), 0, 50, 2)
	check(a.post(200, "GET", path+"/unvote", bob, nil), 0, 50, 2)
	check(a.post(200, "GET", path+"/unvote", alice, nil), -1, 0, 1)
	check(a.post(200, "GET", path+"/unvote", carol, nil), 0, 0, 0)

	if got := a.post(200, "GET", path+"/upvote", alice, nil); got.Views != 0 {
		t.Fatalf("views changed by vote: %d", got.Views)
	}
}

func TestConcurrentVotesAndComments(t *testing.T) {
	a := newAPI(t)
	owner := a.register("owner")
	p := a.textPost(owner, "popular", "c")
	path := "/api/post/" + p.Id

	const n = 15
	tokens := make([]string, n)
	for i := range tokens {
		tokens[i] = a.register(fmt.Sprintf("user%d", i))
	}

	var wg sync.WaitGroup
	for _, tok := range tokens {
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.expect(200, "GET", path+"/upvote", tok, nil)
			a.expect(201, "POST", path, tok, map[string]string{"comment": "hello"})
			a.expect(200, "GET", path, "", nil)
			a.expect(200, "GET", "/api/posts", "", nil)
		}()
	}
	wg.Wait()

	got := a.post(200, "GET", path, "", nil)
	if got.Score != n+1 || len(got.Votes) != n+1 || len(got.Comments) != n || got.Views != n+1 {
		t.Fatalf("score=%d votes=%d comments=%d views=%d; want %d/%d/%d/%d",
			got.Score, len(got.Votes), len(got.Comments), got.Views, n+1, n+1, n, n+1)
	}
}

func TestConcurrentVoteFlipping(t *testing.T) {
	a := newAPI(t)
	owner := a.register("owner")
	p := a.textPost(owner, "flip", "c")
	path := "/api/post/" + p.Id

	const n = 5
	tokens := make([]string, n)
	for i := range tokens {
		tokens[i] = a.register(fmt.Sprintf("flipper%d", i))
	}

	var wg sync.WaitGroup
	for _, tok := range tokens {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				a.expect(200, "GET", path+"/upvote", tok, nil)
				a.expect(200, "GET", path+"/downvote", tok, nil)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				a.expect(200, "GET", path, "", nil)
				a.expect(200, "GET", "/api/posts", "", nil)
			}
		}()
	}
	wg.Wait()

	got := a.post(200, "GET", path, "", nil)
	if got.Score != 1-n || len(got.Votes) != n+1 {
		t.Fatalf("score=%d votes=%d; want %d/%d", got.Score, len(got.Votes), 1-n, n+1)
	}
}
