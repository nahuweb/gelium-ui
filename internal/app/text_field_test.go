package app

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestTextFieldRendersMaterialControlAnatomy(t *testing.T) {
	rendered := renderTextField(t, textFieldView{
		ID: "account-email", Label: "Email", Value: "ada@example.com", Variant: "outlined",
	})

	control := regexp.MustCompile(`<span class="ui-text-field-control">(?s:.*?)</span>`).FindString(rendered)
	if control == "" {
		t.Fatal("text field must group its floating label and native input in a Material control container")
	}
	labelIndex := strings.Index(control, `<label for="account-email">Email</label>`)
	inputIndex := strings.Index(control, `<input`)
	if labelIndex < 0 || inputIndex < 0 || labelIndex > inputIndex {
		t.Errorf("control = %q, want associated floating label before native input", control)
	}
}

func TestTextFieldErrorRendersTrustedDecorativeTrailingIcon(t *testing.T) {
	for _, tt := range []struct {
		name     string
		textarea bool
		native   string
	}{
		{name: "input", native: "<input"},
		{name: "textarea", textarea: true, native: "<textarea"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rendered := renderTextField(t, textFieldView{
				ID: "invalid-field", Label: "Value", Variant: "outlined", Textarea: tt.textarea, Error: "Invalid value",
			})
			control := regexp.MustCompile(`<span class="ui-text-field-control">(?s:.*?)</span>`).FindString(rendered)
			nativeIndex := strings.Index(control, tt.native)
			iconIndex := strings.Index(control, `<svg class="ui-text-field-error-icon" aria-hidden="true" focusable="false"`)
			if nativeIndex < 0 || iconIndex < nativeIndex {
				t.Errorf("error control = %q, want an aria-hidden, unfocusable decorative trusted SVG after the native %s", control, tt.name)
			}
			if strings.Contains(control, "&lt;svg") {
				t.Error("fixed internal error icon must render as trusted SVG markup")
			}
		})
	}
}

func TestTextFieldNativeControlsExposeEmptyStateForFloatingLabel(t *testing.T) {
	for _, tt := range []struct {
		name     string
		textarea bool
		element  string
	}{
		{name: "input", element: "input"},
		{name: "textarea", textarea: true, element: "textarea"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rendered := renderTextField(t, textFieldView{
				ID: "profile-field", Label: "Profile", Variant: "filled", Textarea: tt.textarea,
			})
			control := openingTagWithID(t, rendered, tt.element, "profile-field")
			if !strings.Contains(control, `placeholder=" "`) {
				t.Errorf("%s = %q, want an invisible placeholder hook for the CSS empty/populated state", tt.element, control)
			}
		})
	}
}

func TestTextFieldDocsRouteDogfoodsNativeAccessibleVariantsAndStates(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/text-field", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`>Text field</h1>`,
		`href="/components/text-field"`,
		`ui-text-field-outlined`,
		`ui-text-field-filled`,
		`<textarea`,
		`disabled`,
		`type="submit"`,
		`<section class="component-preview" aria-label="Text field validation example" id="validation-example">`,
		`action="/examples/text-field/validate#validation-example"`,
		`hx-post="/examples/text-field/validate"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("text field docs do not contain contract %q", contract)
		}
	}

	for id, label := range map[string]string{
		"text-normal":            "Name",
		"text-helper":            "Email",
		"text-error":             "Username",
		"text-disabled":          "Account ID",
		"text-disabled-outlined": "Server path",
		"text-disabled-textarea": "Changelog",
		"text-textarea":          "Biography",
		"text-email":             "Email address",
		"text-search":            "Search the docs",
		"text-constrained":       "Handle",
	} {
		if !strings.Contains(body, `<label for="`+id+`">`+label+`</label>`) {
			t.Errorf("field %q is missing its associated label", id)
		}
	}

	for _, id := range []string{"text-disabled", "text-disabled-outlined"} {
		input := openingTagWithID(t, body, "input", id)
		if !strings.Contains(input, "disabled") {
			t.Errorf("dogfooded disabled input %q is missing the native disabled attribute", id)
		}
	}
	textarea := openingTagWithID(t, body, "textarea", "text-disabled-textarea")
	if !strings.Contains(textarea, "disabled") {
		t.Error("dogfooded disabled textarea is missing the native disabled attribute")
	}

	helperInput := openingTagWithID(t, body, "input", "text-helper")
	if !strings.Contains(helperInput, `aria-describedby="text-helper-help"`) || !strings.Contains(body, `id="text-helper-help"`) {
		t.Error("helper input must reference its unique helper ID")
	}
	errorInput := openingTagWithID(t, body, "input", "text-error")
	for _, attribute := range []string{`aria-invalid="true"`, `aria-describedby="text-error-error"`} {
		if !strings.Contains(errorInput, attribute) {
			t.Errorf("error input is missing %s", attribute)
		}
	}
	if !strings.Contains(body, `id="text-error-error"`) || !strings.Contains(body, `role="alert"`) || !strings.Contains(body, `Error:`) {
		t.Error("error state must be associated and visibly signalled beyond color")
	}
	if strings.Contains(errorInput, `autofocus`) {
		t.Error("the static Username error demo must not steal focus on GET")
	}

	emailInput := openingTagWithID(t, body, "input", "text-email")
	for _, attribute := range []string{`type="email"`, `autocomplete="email"`, `inputmode="email"`, `required`} {
		if !strings.Contains(emailInput, attribute) {
			t.Errorf("dogfooded email field is missing %s", attribute)
		}
	}
	searchInput := openingTagWithID(t, body, "input", "text-search")
	for _, attribute := range []string{`type="search"`, `inputmode="search"`, `placeholder="Search components…"`, `maxlength="60"`} {
		if !strings.Contains(searchInput, attribute) {
			t.Errorf("dogfooded search field is missing %s", attribute)
		}
	}
	if strings.Contains(searchInput, `placeholder=" "`) {
		t.Error("a set placeholder must replace the invisible space hook")
	}
	if !strings.Contains(body, `ui-text-field-has-placeholder`) {
		t.Error("a placeholder field must float its label via the has-placeholder modifier")
	}
	constrainedInput := openingTagWithID(t, body, "input", "text-constrained")
	// html/template escapes "+" as "&#43;" inside attribute values; browsers
	// decode it back to the literal regex when matching.
	for _, attribute := range []string{`minlength="3"`, `maxlength="20"`, `pattern="[a-z0-9_]&#43;"`} {
		if !strings.Contains(constrainedInput, attribute) {
			t.Errorf("dogfooded constrained field is missing %s", attribute)
		}
	}
}

func TestTextFieldDocsExplainHTTP422AndHTMXSwapContract(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/text-field", nil))
	body := res.Body.String()
	for _, contract := range []string{"HTTP 422", "X-Gelium-Validation", "htmx:before:swap", "shouldSwap", "isError", "outerHTML", "without JavaScript", "complete documentation page"} {
		if !strings.Contains(body, contract) {
			t.Errorf("text field docs are missing 422/HTMX explanation %q", contract)
		}
	}
}

func TestTextFieldValidationWithHXRejectsWhitespaceWithCompleteAccessibleFormFragment(t *testing.T) {
	form := strings.NewReader("name=+++++")
	req := httptest.NewRequest(http.MethodPost, "/examples/text-field/validate", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	res := httptest.NewRecorder()
	New().ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnprocessableEntity)
	}
	if got := res.Header().Get("X-Gelium-Validation"); got != "true" {
		t.Errorf("X-Gelium-Validation = %q, want %q", got, "true")
	}
	body := res.Body.String()
	if strings.Contains(body, `<!doctype html>`) || strings.Contains(body, `<title>`) {
		t.Error("HX 422 response must be a form fragment, not a complete document")
	}
	if !strings.Contains(body, `<form`) || !strings.Contains(body, `Name is required`) {
		t.Error("422 response must return the complete form and required error")
	}
	input := openingTagWithID(t, body, "input", "validation-name")
	for _, contract := range []string{`name="name"`, `value="     "`, `aria-invalid="true"`, `aria-describedby="validation-name-error"`} {
		if !strings.Contains(input, contract) {
			t.Errorf("422 input is missing %s", contract)
		}
	}
	if !strings.Contains(body, `id="validation-name-error"`) || !strings.Contains(body, `role="alert"`) {
		t.Error("422 error must be associated and announced")
	}
	if strings.Contains(input, `autofocus`) {
		t.Error("HX 422 fragment must not add autofocus")
	}
}

func TestTextFieldValidationWithoutHXRejectsWhitespaceInCompleteDocumentationPage(t *testing.T) {
	form := strings.NewReader("name=+++++")
	req := httptest.NewRequest(http.MethodPost, "/examples/text-field/validate", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	New().ServeHTTP(res, req)

	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnprocessableEntity)
	}
	if got := res.Header().Get("X-Gelium-Validation"); got != "" {
		t.Errorf("non-HX X-Gelium-Validation = %q, want no fragment-only validation header", got)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<!doctype html>`,
		`<title>Text field · Gelium UI</title>`,
		`<nav aria-label="Docs">`,
		`<article class="prose"><p class="article-provenance">`,
		`>Text field</h1>`,
		`aria-label="Text field examples"`,
		`aria-label="Text field validation example"`,
		`Name is required`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("full-page 422 response is missing %q", contract)
		}
	}
	input := openingTagWithID(t, body, "input", "validation-name")
	for _, contract := range []string{`value="     "`, `aria-invalid="true"`, `aria-describedby="validation-name-error"`, `autofocus`} {
		if !strings.Contains(input, contract) {
			t.Errorf("full-page 422 input is missing %s", contract)
		}
	}
	demoInput := openingTagWithID(t, body, "input", "text-error")
	if strings.Contains(demoInput, `autofocus`) {
		t.Error("full-page 422 must focus only the submitted validation field, not the static Username demo")
	}
}

func TestTextFieldValidationWithHXAcceptsValueAndReturnsAccessibleSuccessFormFragment(t *testing.T) {
	form := strings.NewReader("name=Ada+Lovelace")
	req := httptest.NewRequest(http.MethodPost, "/examples/text-field/validate", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	res := httptest.NewRecorder()
	New().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	if strings.Contains(body, `<!doctype html>`) || strings.Contains(body, `<title>`) {
		t.Error("HX 200 response must be a form fragment, not a complete document")
	}
	input := openingTagWithID(t, body, "input", "validation-name")
	for _, contract := range []string{`name="name"`, `value="Ada Lovelace"`, `aria-describedby="validation-name-help"`} {
		if !strings.Contains(input, contract) {
			t.Errorf("200 input is missing %s", contract)
		}
	}
	if !strings.Contains(body, `id="validation-name-help"`) || !strings.Contains(body, `role="status"`) || !strings.Contains(body, `Name accepted`) {
		t.Error("200 form must include an associated accessible success message")
	}
	if strings.Contains(input, `aria-invalid="true"`) {
		t.Error("valid input must not be marked invalid")
	}
}

func TestTextFieldValidationWithoutHXAcceptsValueInCompleteDocumentationPage(t *testing.T) {
	form := strings.NewReader("name=Ada+Lovelace")
	req := httptest.NewRequest(http.MethodPost, "/examples/text-field/validate", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	New().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<!doctype html>`,
		`<title>Text field · Gelium UI</title>`,
		`<nav aria-label="Docs">`,
		`<article class="prose"><p class="article-provenance">`,
		`>Text field</h1>`,
		`aria-label="Text field examples"`,
		`aria-label="Text field validation example"`,
		`role="status"`,
		`Name accepted`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("full-page 200 response is missing %q", contract)
		}
	}
	input := openingTagWithID(t, body, "input", "validation-name")
	for _, contract := range []string{`value="Ada Lovelace"`, `aria-describedby="validation-name-help"`} {
		if !strings.Contains(input, contract) {
			t.Errorf("full-page 200 input is missing %s", contract)
		}
	}
	if strings.Contains(input, `aria-invalid="true"`) {
		t.Error("valid full-page input must not be marked invalid")
	}
	if strings.Contains(input, `autofocus`) {
		t.Error("valid full-page input must not autofocus")
	}
}

func TestOnlyTrueHXRequestHeaderReceivesValidationFragment(t *testing.T) {
	tests := []struct {
		name         string
		header       string
		wantDocument bool
	}{
		{name: "missing", wantDocument: true},
		{name: "false", header: "false", wantDocument: true},
		{name: "surrounded true", header: " true ", wantDocument: true},
		{name: "true", header: "true"},
		{name: "case insensitive true", header: "TRUE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/examples/text-field/validate", strings.NewReader("name=Ada"))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tt.header != "" {
				req.Header.Set("HX-Request", tt.header)
			}
			res := httptest.NewRecorder()
			New().ServeHTTP(res, req)

			gotDocument := strings.Contains(res.Body.String(), `<!doctype html>`)
			if gotDocument != tt.wantDocument {
				t.Errorf("HX-Request %q document = %t, want %t", tt.header, gotDocument, tt.wantDocument)
			}
		})
	}
}

func TestTextFieldValidationRouteUsesModernMuxMethodSemantics(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/examples/text-field/validate", nil))
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET validation status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}

	res = httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/examples/text-field/missing", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("unknown route status = %d, want %d", res.Code, http.StatusNotFound)
	}
}

func TestTextFieldFormAndLocalScriptImplementHTMX422SwapContract(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/text-field", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("docs status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	// Scope to the validation form specifically: the docs chrome now renders
	// its own GET forms (theme select / scheme switch) earlier in the body.
	form := regexp.MustCompile(`<form\b[^>]*hx-post="/examples/text-field/validate"[^>]*>`).FindString(body)
	if form == "" {
		t.Fatal("docs are missing validation form")
	}
	for _, attribute := range []string{
		`novalidate`,
		`hx-post="/examples/text-field/validate"`,
		`hx-target="this"`,
		`hx-swap="outerHTML"`,
	} {
		if !strings.Contains(form, attribute) {
			t.Errorf("validation form is missing %s", attribute)
		}
	}
	htmxIndex := strings.Index(body, `src="/static/htmx.min.js?v=0.6.7"`)
	geliumIndex := strings.Index(body, `src="/static/gelium.js?v=0.6.7"`)
	appIndex := strings.Index(body, `src="/static/app.js?v=0.6.7"`)
	if htmxIndex < 0 || geliumIndex < 0 || appIndex < 0 || geliumIndex < htmxIndex || appIndex < geliumIndex {
		t.Error("script order must be htmx → gelium.js (consumer) → app.js (chrome)")
	}

	asset := httptest.NewRecorder()
	New().ServeHTTP(asset, httptest.NewRequest(http.MethodGet, "/static/gelium.js", nil))
	if asset.Code != http.StatusOK {
		t.Fatalf("gelium.js status = %d, want %d", asset.Code, http.StatusOK)
	}
	if got := asset.Header().Get("Content-Type"); got != "text/javascript; charset=utf-8" {
		t.Errorf("gelium.js Content-Type = %q", got)
	}
	js := asset.Body.String()
	for _, contract := range []string{
		`htmx:before:swap`,
		`ctx.response`,
		`shouldSwap = true`,
		`isError = false`,
	} {
		if !strings.Contains(js, contract) {
			t.Errorf("gelium.js is missing 422 hook contract %q", contract)
		}
	}
	validation422 := regexp.MustCompile(`response\.status\s*===\s*422\s*&&\s*response\.headers\.get\("X-Gelium-Validation"\)\s*===\s*"true"`)
	if !validation422.MatchString(js) {
		t.Error("gelium.js must only swap a 422 when X-Gelium-Validation is true")
	}
}

// TestTextFieldDisabledStateKeepsBlockedAttributeAndPrecedesError locks the
// explicit state precedence contract demanded by the roadmap: disabled wins
// over error. A disabled field must carry the real disabled attribute, must
// never advertise aria-invalid or an error alert, and must not expose an error
// icon even when the view model carries both flags.
func TestTextFieldDisabledStateKeepsBlockedAttributeAndPrecedesError(t *testing.T) {
	rendered := renderTextField(t, textFieldView{ID: "d", Label: "Account", Name: "account", Value: "ACCT-1", Variant: "filled", Disabled: true, Error: "Should never surface"})

	input := openingTagWithID(t, rendered, "input", "d")
	if !strings.Contains(input, "disabled") {
		t.Error("disabled field must carry the native disabled attribute")
	}
	for _, forbidden := range []string{`aria-invalid`, `aria-describedby`, `role="alert"`, `<p class="ui-text-field-message"`} {
		if strings.Contains(rendered, forbidden) {
			t.Errorf("disabled state must precede error state, but rendered %q", forbidden)
		}
	}
	if strings.Contains(rendered, "ui-text-field-error") {
		t.Error("disabled field must not expose the error state class")
	}
	if strings.Contains(rendered, "invalid") {
		t.Error("disabled field must not render the error message text")
	}
}

// TestTextFieldDocsDisabledFieldCarriesBlockingAttribute isolates the dogfooded
// disabled preview: the exact <input> with id "text-disabled" must carry the
// native disabled attribute and must not be announced as invalid.
func TestTextFieldDocsDisabledFieldCarriesBlockedAttribute(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/text-field", nil))
	body := res.Body.String()
	disabledInput := openingTagWithID(t, body, "input", "text-disabled")
	if !strings.Contains(disabledInput, "disabled") {
		t.Error("dogfooded disabled input is missing the native disabled attribute")
	}
	if strings.Contains(disabledInput, "aria-invalid") {
		t.Error("dogfooded disabled input must not be announced as invalid")
	}
}

// TestTextFieldRendersInputSurfaceAttributesWhenSet locks the full input
// surface: every new native attribute renders from its view field, and a set
// placeholder replaces the invisible space hook while floating the label via
// the has-placeholder modifier class (no JS involved).
func TestTextFieldRendersInputSurfaceAttributesWhenSet(t *testing.T) {
	rendered := renderTextField(t, textFieldView{
		ID: "account-email", Label: "Email", Name: "email", Variant: "outlined",
		Type: "email", Placeholder: "you@example.com", Pattern: "[^@]+@[^@]+",
		Autocomplete: "email", InputMode: "email", MaxLength: 254, MinLength: 5,
		Required: true, ReadOnly: true,
	})

	input := openingTagWithID(t, rendered, "input", "account-email")
	// html/template escapes "+" as "&#43;" inside attribute values; browsers
	// decode it back to the literal regex when matching.
	for _, attribute := range []string{
		`type="email"`, `placeholder="you@example.com"`, `pattern="[^@]&#43;@[^@]&#43;"`,
		`autocomplete="email"`, `inputmode="email"`, `maxlength="254"`, `minlength="5"`,
		`required`, `readonly`,
	} {
		if !strings.Contains(input, attribute) {
			t.Errorf("input surface is missing attribute %s in %q", attribute, input)
		}
	}
	if strings.Contains(input, `placeholder=" "`) {
		t.Error("a set placeholder must replace the invisible space hook")
	}
	if !strings.Contains(rendered, "ui-text-field-has-placeholder") {
		t.Error("a placeholder field must float its label via the has-placeholder modifier")
	}
}

// TestTextFieldOmitsInputSurfaceAttributesWhenZero locks backward
// compatibility: empty/zero values must keep the template's defaults (type
// text, the invisible space placeholder hook) and must not emit any of the
// new attributes.
func TestTextFieldOmitsInputSurfaceAttributesWhenZero(t *testing.T) {
	rendered := renderTextField(t, textFieldView{ID: "plain", Label: "Name", Variant: "outlined"})

	input := openingTagWithID(t, rendered, "input", "plain")
	if !strings.Contains(input, `type="text"`) {
		t.Errorf("default type must remain text, got %q", input)
	}
	if !strings.Contains(input, `placeholder=" "`) {
		t.Error("empty placeholder must keep the invisible space hook")
	}
	for _, forbidden := range []string{`required`, `maxlength=`, `minlength=`, `pattern=`, `autocomplete=`, `inputmode=`, `readonly`} {
		if strings.Contains(input, forbidden) {
			t.Errorf("zero-valued input surface must not render %q in %q", forbidden, input)
		}
	}
	if strings.Contains(rendered, "ui-text-field-has-placeholder") {
		t.Error("a placeholder-less field must not carry the has-placeholder modifier")
	}
}

// TestTextFieldPassesThroughSupportedNativeTypes covers the documented type
// vocabulary: email, url, tel, number, password, and search all render as the
// native type attribute on the input.
func TestTextFieldPassesThroughSupportedNativeTypes(t *testing.T) {
	for _, inputType := range []string{"email", "url", "tel", "number", "password", "search"} {
		t.Run(inputType, func(t *testing.T) {
			rendered := renderTextField(t, textFieldView{ID: "t", Label: "Field", Variant: "outlined", Type: inputType})
			input := openingTagWithID(t, rendered, "input", "t")
			if !strings.Contains(input, `type="`+inputType+`"`) {
				t.Errorf("type %q not rendered, got %q", inputType, input)
			}
		})
	}
}

// TestTextFieldTextareaSupportsConstraintsButNotTypeOrPattern locks the
// textarea branch of the input surface: constraints and hints apply, while
// type and pattern stay input-only and the value lives in the element body.
func TestTextFieldTextareaSupportsConstraintsButNotTypeOrPattern(t *testing.T) {
	rendered := renderTextField(t, textFieldView{
		ID: "bio", Label: "Biography", Name: "bio", Variant: "outlined", Textarea: true,
		Placeholder: "Tell people about yourself…", Required: true, MaxLength: 400,
		MinLength: 10, Autocomplete: "off", InputMode: "text", ReadOnly: true,
	})

	textarea := openingTagWithID(t, rendered, "textarea", "bio")
	for _, attribute := range []string{
		`placeholder="Tell people about yourself…"`, `required`, `maxlength="400"`,
		`minlength="10"`, `autocomplete="off"`, `inputmode="text"`, `readonly`,
	} {
		if !strings.Contains(textarea, attribute) {
			t.Errorf("textarea is missing %s in %q", attribute, textarea)
		}
	}
	for _, forbidden := range []string{`type=`, `pattern=`, `value="`} {
		if strings.Contains(textarea, forbidden) {
			t.Errorf("textarea must not render %q in %q", forbidden, textarea)
		}
	}
}
