package app

import (
	"bytes"
	"html/template"
	"net/http"

	"geliumui/lib"
)

// landingView is the marketing home composition. Every section dogfoods an
// existing Gelium public/content pattern or component — no parallel landing CSS
// system beyond a thin page frame.
type landingView struct {
	Hero            *heroView
	Claims          []string
	FeaturesHeading sectionHeadingView
	Features        []featureCardView
	Split           *splitView
	Agent           *splitView
	Demo            *landingDemoView
	Recipes         *landingRecipesView
	FAQ             *landingFAQView
	CTABand         *landingCTABandView
}

// heroView matches web/templates/hero.html.
type heroView struct {
	Eyebrow  string
	Title    string
	Subtitle string
	CTAs     []buttonView
	Media    template.HTML
}

// featureCardView matches web/templates/feature-card.html.
type featureCardView struct {
	Title string
	Body  string
	CTA   *buttonView
	Media template.HTML
}

// splitView matches web/templates/split.html.
type splitView struct {
	Eyebrow string
	Title   string
	Body    string
	CTA     *buttonView
	Media   template.HTML
}

// sectionHeadingView matches web/templates/section-heading.html.
type sectionHeadingView struct {
	Eyebrow  string
	Title    string
	Centered bool
}

// landingRecipesView is a section of outbound recipe cards.
type landingRecipesView struct {
	Heading sectionHeadingView
	Cards   []featureCardView
}

// landingCTABandView is the closing call-to-action band.
type landingCTABandView struct {
	Heading sectionHeadingView
	CTA     buttonView
	// SecondaryCTA is an optional secondary action rendered beside the primary
	// (e.g. the GitHub source link). It is intentionally NOT chrome-rewritten:
	// external hrefs must never carry ?theme= / ?scheme= query strings.
	SecondaryCTA *buttonView
}

// landingDemoView is the visual demo card (Basecoat pattern): a prominent
// link into the live WhatsApp demo with a token-styled preview.
type landingDemoView struct {
	Heading sectionHeadingView
	Body    string
	CTA     buttonView
	Media   template.HTML
}

// landingFAQView is the closing FAQ section (Base UI pattern): answers
// developer objections with zero-JS native disclosures.
type landingFAQView struct {
	Heading sectionHeadingView
	Items   []faqItemView
}

// faqItemView is a single question/answer pair.
type faqItemView struct {
	Question string
	Answer   string
}

// marketingLanding builds the home page composition from Gelium primitives.
// Persuade mode: exactly one primary button on the page (hero Get started).
func marketingLanding() landingView {
	return landingView{
		Hero: &heroView{
			Eyebrow:  "Gelium UI",
			Title:    "Build UI that works without JavaScript.",
			Subtitle: "Open-code components, semantic HTML, token themes, and server-first flows for HTML-first applications.",
			CTAs: []buttonView{
				{Label: "Install gelium-ui", Variant: "primary", Href: "#landing-install"},
				{Label: "Give this to your LLM", Variant: "secondary", Href: "https://skills.sh/nahuweb/gelium-ui/install-gelium"},
			},
		},
		Claims: []string{
			"HTML-first",
			"No-JS baseline",
			"Server-first state",
			"Themes without forks",
		},
		FeaturesHeading: sectionHeadingView{
			Eyebrow:  "Why Gelium",
			Title:    "Built for server-rendered apps",
			Centered: true,
		},
		Features: []featureCardView{
			{
				Title: "Native HTML first",
				Body:  "Real buttons, dialogs, selects, and tables. ARIA only when the platform has no equivalent.",
				CTA:   &buttonView{Label: "Read principles", Variant: "outline", Href: "/docs/principles"},
			},
			{
				Title: "Tokens, not forks",
				Body:  "Themes map aesthetics onto --ui-* variables. Markup stays stable when you switch Material or Basecoat.",
				CTA:   &buttonView{Label: "Themes", Variant: "outline", Href: "/docs/themes"},
			},
			{
				Title: "Server contracts",
				Body:  "GET with stable query params, POST + 303, 422 validation, and gelium:toast for transient feedback — no parallel APIs.",
				CTA:   &buttonView{Label: "See a recipe", Variant: "outline", Href: "/recipes/admin-resource"},
			},
		},
		Split: &splitView{
			Eyebrow: "How it fits",
			Title:   "Start with the package. Copy the open code.",
			Body:    "Install gelium-ui for CSS, themes, templates, optional JS helpers, agent guidance, and portable references. The server renders the HTML; progressive enhancement stays optional.",
			CTA:     &buttonView{Label: "Open the docs", Variant: "secondary", Href: "/docs"},
			Media: template.HTML(
				`<pre class="ui-landing-code" tabindex="0"><code>npm install gelium-ui

/* CSS */
@import "gelium-ui/dist/gelium.css";

/* Theme on &lt;html&gt; */
&lt;html class="theme-material"&gt;

/* First component */
&lt;button class="ui-button"&gt;Save changes&lt;/button&gt;

/* Rendered output */
<div class="ui-landing-first-component" aria-label="Rendered button example">
  <p>Rendered output</p>
  <button class="ui-button" type="button">Save changes</button>
</div>

/* Optional */
// gelium.js — toast + 422 helper</code></pre>`,
			),
		},
		Agent: &splitView{
			Eyebrow: "LLM context",
			Title:   "Give your LLM the right instructions.",
			Body:    "Gelium ships agent guidance so your LLM can follow the same HTML-first, token, accessibility, and server-contract rules as the library.",
			CTA:     &buttonView{Label: "Open agent workflow", Variant: "secondary", Href: "https://skills.sh/nahuweb/gelium-ui/install-gelium"},
			Media: template.HTML(`<pre class="ui-landing-code" tabindex="0"><code>Use Gelium UI.

Read the project skill before writing UI.
Keep the main flow functional without JavaScript.
Use semantic HTML and --ui-* tokens.
Validate with the Gelium UX checklist.

Install skill:
https://skills.sh/nahuweb/gelium-ui/install-gelium</code></pre>`),
		},
		Demo: &landingDemoView{
			Heading: sectionHeadingView{
				Eyebrow:  "See it live",
				Title:    "A real app, running on Gelium",
				Centered: true,
			},
			Body: "The WhatsApp manager demo — chat list, search, send, and templates — is a complete server-rendered app built from the same public components and tokens you install.",
			CTA:  buttonView{Label: "Launch live demo", Variant: "secondary", Href: "/demo/whatsapp"},
			Media: template.HTML(
				`<div class="ui-landing-demo-phone" aria-hidden="true">
  <div class="ui-landing-demo-phone-bar"></div>
  <div class="ui-landing-demo-chat">
    <div class="ui-landing-demo-bubble ui-landing-demo-bubble--in"></div>
    <div class="ui-landing-demo-bubble ui-landing-demo-bubble--out"></div>
    <div class="ui-landing-demo-bubble ui-landing-demo-bubble--in ui-landing-demo-bubble--short"></div>
  </div>
</div>`,
			),
		},
		Recipes: &landingRecipesView{
			Heading: sectionHeadingView{
				Eyebrow:  "Screen recipes",
				Title:    "Full flows composed from the catalog",
				Centered: true,
			},
			Cards: []featureCardView{
				{
					Title: "Admin Resource",
					Body:  "Table, filters, create/edit forms, and delete confirm — server-driven end to end.",
					CTA:   &buttonView{Label: "Open recipe", Variant: "outline", Href: "/recipes/admin-resource"},
				},
				{
					Title: "Ops Queue",
					Body:  "FIFO work queue with advance/dequeue actions and refresh without a SPA.",
					CTA:   &buttonView{Label: "Open recipe", Variant: "outline", Href: "/recipes/ops-queue"},
				},
				{
					Title: "Public Feed",
					Body:  "Social-style feed composition with reactions and progressive enhancement.",
					CTA:   &buttonView{Label: "Open recipe", Variant: "outline", Href: "/recipes/public-feed"},
				},
			},
		},
		FAQ: &landingFAQView{
			Heading: sectionHeadingView{
				Eyebrow:  "FAQ",
				Title:    "Frequently asked questions",
				Centered: true,
			},
			Items: []faqItemView{
				{
					Question: "Is Gelium UI a framework?",
					Answer:   "No — the UI package is npm gelium-ui (HTML/CSS/tokens). This documentation site is a Go dogfood app; there is no SPA runtime to install and no service to run for components.",
				},
				{
					Question: "Does it work with my existing stack?",
					Answer:   "Yes — components are HTML partials and CSS, so any server-rendered stack can use them. HTMX stays optional: the core contract is plain forms, links, and progressive enhancement.",
				},
				{
					Question: "How does it differ from a React component library?",
					Answer:   "No component JavaScript and no framework runtime for consumers — Gelium ships server contracts instead of props, and the server renders the HTML.",
				},
				{
					Question: "Can I switch themes?",
					Answer:   "Yes — Material is the default and Basecoat ships in the same bundle. Switch with ?theme=basecoat on any URL or a class on the html element; no rebuild required.",
				},
			},
		},
		CTABand: &landingCTABandView{
			Heading: sectionHeadingView{
				Eyebrow:  "Ready when you are",
				Title:    "Ship UI that works without JavaScript",
				Centered: true,
			},
			// Secondary only — hero already owns the single primary (Persuade / Screens).
			CTA: buttonView{Label: "Read the docs", Variant: "secondary", Href: "/docs"},
			SecondaryCTA: &buttonView{
				Label:   "View source",
				Variant: "outline",
				Href:    "https://github.com/cryptonahue/gelium-ui",
			},
		},
	}
}

// homeLandingNav is the compact primary nav for the marketing site chrome.
func homeLandingNav() []navLink {
	return []navLink{
		{Path: "/docs", Label: "Docs"},
		{Path: "/docs/themes/gallery", Label: "Theme Gallery"},
		{Path: "/components/button", Label: "Components"},
		{Path: "/recipes/admin-resource", Label: "Recipes"},
		{Path: "/docs/agent-workflow", Label: "Agents"},
		{Path: "/demo/whatsapp", Label: "Demo"},
	}
}

func marketingFooter() *footerView {
	sections := []footerSection{
		{Title: "Package", Links: []navLink{
			{Path: "/components/button", Label: "Component registry"},
			{Path: "/docs/themes/gallery", Label: "Themes"},
			{Path: "/docs/templates/product", Label: "Templates"},
		}},
		{Title: "Learn", Links: []navLink{
			{Path: "/docs", Label: "Documentation"},
			{Path: "/docs/server-contracts", Label: "Server contracts"},
			{Path: "/docs/agent-workflow", Label: "Agent guidance"},
		}},
		{Title: "Resources", Links: []navLink{
			{Path: "https://www.npmjs.com/package/gelium-ui", Label: "npm"},
			{Path: "https://github.com/cryptonahue/gelium-ui", Label: "GitHub"},
			{Path: "/docs/changelog", Label: "Changelog"},
		}},
		{Title: "Community", Links: []navLink{
			{Path: "https://github.com/cryptonahue/gelium-ui/issues", Label: "Issues"},
			{Path: "https://github.com/cryptonahue/gelium-ui/discussions", Label: "Discussions"},
		}},
	}
	return &footerView{
		Brand: "Gelium UI",
		Sections: sections,
		Legal: "© 2026 Gelium UI · MIT",
	}
}

func (s *server) home(w http.ResponseWriter, r *http.Request) {
	landing := marketingLanding()
	data := pageView{
		Title:   "Themeable open-code UI components for server-rendered apps",
		Landing: &landing,
		Nav:     homeLandingNav(),
	}
	s.renderLanding(w, r, data)
}

// renderLanding renders the marketing home (or any pageView with Landing set)
// through the shared layout without Markdown content.
func (s *server) renderLanding(w http.ResponseWriter, r *http.Request, data pageView) {
	routePath := "/"
	data.Meta = resolveMeta(data, routePath)
	data.AssetsVersion = lib.AssetsVersion
	if data.Footer == nil {
		data.Footer = marketingFooter()
	}

	selection := applyDocumentSelection(&data, r)
	execution := accordionExecutionFromRequest(r)
	navigation := navigationSelectionFor(r, routePath)

	// Site chrome keeps the native Recipe form; ordinary landing links retain
	// only legacy chrome so recipe fields are never globally published.
	data.RecipeSwitcher = recipeSwitcherFor(r, selection, execution, true, "site-recipe")
	data.SchemeSwitcher = schemeSwitcherForNavigation(selection, execution, navigation)

	for i := range data.Nav {
		data.Nav[i].Path = navigation.href(data.Nav[i].Path)
	}
	if data.Landing != nil {
		applyLandingNavigation(data.Landing, navigation)
	}

	var page bytes.Buffer
	if err := s.templates.ExecuteTemplate(&page, "layout", data); err != nil {
		http.Error(w, "page unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(page.Bytes())
}

// applyLandingNavigation uses the same source-aware link policy as the shell.
func applyLandingNavigation(l *landingView, navigation navigationSelection) {
	withNavigation := func(href string) string { return navigation.href(href) }
	if l.Hero != nil {
		for i := range l.Hero.CTAs {
			if l.Hero.CTAs[i].Href != "" {
				l.Hero.CTAs[i].Href = withNavigation(l.Hero.CTAs[i].Href)
			}
		}
	}
	for i := range l.Features {
		if l.Features[i].CTA != nil && l.Features[i].CTA.Href != "" {
			l.Features[i].CTA.Href = withNavigation(l.Features[i].CTA.Href)
		}
	}
	if l.Split != nil && l.Split.CTA != nil && l.Split.CTA.Href != "" {
		l.Split.CTA.Href = withNavigation(l.Split.CTA.Href)
	}
	if l.Demo != nil && l.Demo.CTA.Href != "" {
		l.Demo.CTA.Href = withNavigation(l.Demo.CTA.Href)
	}
	if l.Recipes != nil {
		for i := range l.Recipes.Cards {
			if l.Recipes.Cards[i].CTA != nil && l.Recipes.Cards[i].CTA.Href != "" {
				l.Recipes.Cards[i].CTA.Href = withNavigation(l.Recipes.Cards[i].CTA.Href)
			}
		}
	}
	if l.CTABand != nil && l.CTABand.CTA.Href != "" {
		l.CTABand.CTA.Href = withNavigation(l.CTABand.CTA.Href)
	}
}
