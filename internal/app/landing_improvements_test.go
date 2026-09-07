package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLandingExplainsGeliumForLLMUsers(t *testing.T) {
	body := getOKBody(t, "/")
	for _, want := range []string{
		"Build server-rendered UI your LLM can understand.",
		"A better default for LLM-built interfaces",
		"Install in under a minute.",
		"Give your LLM the right instructions.",
		"Open agent workflow",
		"npm install gelium-ui",
		"gelium-ui/dist/gelium.css",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("landing missing %q", want)
		}
	}
}

func TestLandingOmitsOutOfScopeSections(t *testing.T) {
	body := getOKBody(t, "/")
	for _, absent := range []string{
		"Frequently asked questions",
		"Launch live demo",
		"Screen recipes",
	} {
		if strings.Contains(body, absent) {
			t.Errorf("landing must not render out-of-scope content %q", absent)
		}
	}
}

func TestLandingClaimsStrip(t *testing.T) {
	body := getOKBody(t, "/")
	for _, claim := range []string{
		"HTML-first",
		"No-JS baseline",
		"Server-first state",
		"Themes without forks",
		"Agent-ready guidance",
	} {
		if !strings.Contains(body, claim) {
			t.Errorf("landing claims strip missing %q", claim)
		}
	}
	if got := strings.Count(body, "✓"); got != 5 {
		t.Errorf("landing claims strip rendered %d checkmarks, want 5", got)
	}
}

func TestLandingUnchangedStatus(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := strings.Count(res.Body.String(), "<h1"); got != 1 {
		t.Errorf("landing rendered %d <h1> elements, want 1", got)
	}
}
