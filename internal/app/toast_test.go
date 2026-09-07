package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestToastRendersMaterialSnackbarAnatomy(t *testing.T) {
	rendered := renderToast(t, toastView{ID: "toast-1", Type: "success", Role: "status", Message: "Saved", IconSVG: toastIcons["success"], Dismiss: true})
	for _, contract := range []string{
		`<div class="ui-toast ui-toast-success" id="toast-1" role="status">`,
		`aria-hidden="true"`,
		`focusable="false"`,
		`<span class="ui-toast-message">Saved</span>`,
		`<button class="ui-toast-action" type="button" data-gelium-toast-dismiss aria-label="Dismiss notification">Dismiss</button>`,
	} {
		if !strings.Contains(rendered, contract) {
			t.Errorf("toast is missing %q", contract)
		}
	}
	if strings.Contains(rendered, "&lt;svg") {
		t.Error("fixed internal icon must render as trusted SVG markup, not escaped text")
	}
}

func TestToastRoleIsAssertiveOnlyForErrors(t *testing.T) {
	for _, tt := range []struct{ typ, role string }{
		{"info", "status"}, {"success", "status"}, {"warning", "status"}, {"error", "alert"},
	} {
		rendered := renderToast(t, newToast(tt.typ, "x", "m"))
		role := regexp.MustCompile(`role="[^"]+"`).FindString(rendered)
		if role != `role="`+tt.role+`"` {
			t.Errorf("toast type %q role = %q, want role=%q", tt.typ, role, tt.role)
		}
	}
}

func TestSanitizeToastTypeFallsBackToClosedVocabulary(t *testing.T) {
	for _, in := range []string{"info", "success", "warning", "error"} {
		if got := sanitizeToastType(in); got != in {
			t.Errorf("sanitizeToastType(%q) = %q, want %q", in, got, in)
		}
	}
	for _, in := range []string{"", "fatal", "SUCCESS", "danger", "success alert"} {
		if got := sanitizeToastType(in); got != "info" {
			t.Errorf("sanitizeToastType(%q) = %q, want info", in, got)
		}
	}
}

func TestToastDocsRenderVariantsAndLiveRegion(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/toast", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`id="gelium-toast-region"`,
		`aria-live="polite"`,
		`class="ui-toast ui-toast-success"`,
		`class="ui-toast ui-toast-error"`,
		`name="type"`,
		`hx-post="/examples/toast/demo"`,
		`/components/toast`,
		`?v=0.6.7`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("toast docs page is missing %q", contract)
		}
	}
}

func TestToastDemoNoJSRendersPersistentInlineFeedback(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/examples/toast/demo", strings.NewReader("message=Record+updated&type=success"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	New().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	if !strings.Contains(body, `<div class="ui-toast ui-toast-success" id="toast-demo-result" role="status">`) {
		t.Error("no-JS toast demo must render a persistent inline toast")
	}
	if !strings.Contains(body, "Record updated") {
		t.Error("inline toast must carry the submitted message")
	}
	if got := res.Header().Get("HX-Trigger"); got != "" {
		t.Errorf("no-JS response must not rely on HX-Trigger, got %q", got)
	}
}

func TestToastDemoNoJSEmptyMessageRejects422WithoutToast(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/examples/toast/demo", strings.NewReader("message=&type=error"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	New().ServeHTTP(res, req)
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnprocessableEntity)
	}
	body := res.Body.String()
	if !strings.Contains(body, "Message is required") {
		t.Error("empty message must render an inline validation error")
	}
	if strings.Contains(body, `id="toast-demo-result"`) {
		t.Error("validation failures must not be announced as toast notifications")
	}
}
func TestToastDemoHXReturnsTriggerAndFragment(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/examples/toast/demo", strings.NewReader("message=Saved&type=success"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	New().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	trigger := res.Header().Get("HX-Trigger")
	if trigger == "" {
		t.Fatal("HTMX toast demo must set HX-Trigger")
	}
	var parsed struct {
		Toast struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"gelium:toast"`
	}
	if err := json.Unmarshal([]byte(trigger), &parsed); err != nil {
		t.Fatalf("HX-Trigger is not valid JSON: %v", err)
	}
	if parsed.Toast.Type != "success" || parsed.Toast.Message != "Saved" {
		t.Errorf("HX-Trigger = %q, want type success and message Saved", trigger)
	}
	body := res.Body.String()
	if !strings.Contains(body, `<form class="toast-demo-form"`) {
		t.Error("HTMX response must be the toast demo form fragment, not a full page")
	}
	if strings.Contains(body, "<html") {
		t.Error("HTMX response must not contain a full document")
	}
	if strings.Contains(body, `id="toast-demo-result"`) {
		t.Error("HTMX demo must rely on the live region, not prerender an inline toast")
	}
}

func TestToastDemoHXEmptyMessageRejects422WithoutTrigger(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/examples/toast/demo", strings.NewReader("message=&type=success"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	New().ServeHTTP(res, req)
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnprocessableEntity)
	}
	if got := res.Header().Get("X-Gelium-Validation"); got != "true" {
		t.Errorf("X-Gelium-Validation = %q, want true", got)
	}
	if got := res.Header().Get("HX-Trigger"); got != "" {
		t.Errorf("validation failure must not raise gelium:toast, got %q", got)
	}
	body := res.Body.String()
	if !strings.Contains(body, "Message is required") {
		t.Error("fragment must render the inline validation error")
	}
	if !strings.Contains(body, `<form class="toast-demo-form"`) {
		t.Error("HTMX 422 must return the toast demo form fragment")
	}
}

func TestToastTriggerJSONEscapesMessage(t *testing.T) {
	trigger, err := toastTriggerJSON("success", `He said "hi" <b> & "bye"`)
	if err != nil {
		t.Fatalf("toastTriggerJSON: %v", err)
	}
	if strings.Contains(trigger, "<b>") {
		t.Error("toast message HTML must be escaped inside the JSON header")
	}
	var parsed struct {
		Toast struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"gelium:toast"`
	}
	if err := json.Unmarshal([]byte(trigger), &parsed); err != nil {
		t.Fatalf("escaped header must remain valid JSON: %v", err)
	}
	if parsed.Toast.Message != `He said "hi" <b> & "bye"` {
		t.Errorf("round-trip message = %q, want original", parsed.Toast.Message)
	}
}

func TestAppJSWiresHTMXToastTriggerEvent(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/static/gelium.js", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("gelium.js status = %d, want %d", res.Code, http.StatusOK)
	}
	js := res.Body.String()
	for _, contract := range []string{
		`addEventListener("gelium:toast"`,
		`#gelium-toast-region`,
		`ui-toast-show`,
		`Dismiss notification`,
	} {
		if !strings.Contains(js, contract) {
			t.Errorf("gelium.js is missing toast enhancement contract %q", contract)
		}
	}
}
