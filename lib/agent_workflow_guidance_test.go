package lib

import (
	"strings"
	"testing"
)

// TestAgentWorkflowGuidanceContract keeps the distributed agent pack aligned:
// every installed guidance entrypoint must describe the same proportional route
// and point at the shipped records needed for a design-gated change.
func TestAgentWorkflowGuidanceContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		path []string
		want []string
	}{
		{
			name: "repository entrypoint",
			path: []string{"AGENTS.md"},
			want: []string{
				"lib/SKILL.md",
				"lib/skills/00-agent-routing.md",
				"lib/skills/14-component-implementation.md",
				"go test ./...",
			},
		},
		{
			name: "entrypoint",
			path: []string{"lib", "AGENTS.md"},
			want: []string{
				"ROUTE → ORIENT → PLAN → ARCHITECT → APPROVE → BUILD → AUDIT → RELEASE",
				"direct-exempt",
				"delegated-direct",
				"design-gated",
				"required for design-gated",
				"skills/00-agent-routing.md",
				"skills/templates/gate-ledger.md",
			},
		},
		{
			name: "decision pack",
			path: []string{"lib", "llms-ux.txt"},
			want: []string{
				"ROUTE → ORIENT → PLAN → ARCHITECT → APPROVE → BUILD → AUDIT → RELEASE",
				"skills/00-agent-routing.md",
				"delegated-direct",
				"prebuild criteria plan",
				"rendered audit",
				"media-metadata-unknown",
				"skills/templates/wireframe-approval-packet.md",
				"global shell",
				"WIDE and NARROW",
				"INTENT MODE",
				"WF-DIRECTION",
			},
		},
		{
			name: "served decision pack",
			path: []string{"site", "web", "static", "llms-ux.txt"},
			want: []string{
				"skills/00-agent-routing.md",
				"delegated-direct",
				"design-gated",
			},
		},
		{
			name: "skill index",
			path: []string{"lib", "SKILLS.md"},
			want: []string{
				"skills/00-agent-routing.md",
				"ROUTE → ORIENT → PLAN → ARCHITECT → APPROVE → BUILD → AUDIT → RELEASE",
				"skills/templates/gate-ledger.md",
				"skills/templates/wireframe-approval-packet.md",
			},
		},
		{
			name: "canonical routing",
			path: []string{"lib", "skills", "00-agent-routing.md"},
			want: []string{
				"direct-exempt",
				"delegated-direct",
				"design-gated",
				"full-sdd",
				"Working",
				"Needs your decision",
				"Checking",
				"Ready",
				"per action",
				"authority boundaries",
				"visual-migration",
			},
		},

		{
			name: "criteria skill",
			path: []string{"lib", "skills", "11-design-criteria.md"},
			want: []string{
				"Criteria plan (prebuild)",
				"Rendered audit (postbuild)",
				"pass-with-escalation",
				"Gelium icon allowlist",
				"extract-used-icons",
				"--set material|tabler",
				"Visual + interaction direction",
			},
		},
		{
			name: "foundations skill",
			path: []string{"lib", "skills", "01-foundations.md"},
			want: []string{
				"Do not flatten mixed relationships",
				"--ui-space-1",
			},
		},
		{
			name: "approval skill",
			path: []string{"lib", "skills", "12-wireframe-approval.md"},
			want: []string{
				"Intent wireframe (Plan)",
				"Buildable wireframe (Architect)",
				"approved | changes-requested | declined | exception",
				"Visible packet",
				"continua",
				"Gelium structural ASCII wireframes",
				"Global shell",
				"WIDE",
				"NARROW",
				"States and recovery",
				"Intent mode",
				"theme/recipe/skin",
			},
		},
		{
			name: "component implementation skill",
			path: []string{"lib", "skills", "14-component-implementation.md"},
			want: []string{
				"gelium-ui-component-implementation",
				"NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST",
				"styles_contract_test.go",
				"READY_FOR_INTEGRATION",
			},
		},
		{
			name: "gate ledger template",
			path: []string{"lib", "skills", "templates", "gate-ledger.md"},
			want: []string{
				`"schema_version": 1`,
				"Reading attestations",
				"Prebuild gates",
				"Rendered audit",
			},
		},
		{
			name: "approval packet template",
			path: []string{"lib", "skills", "templates", "wireframe-approval-packet.md"},
			want: []string{
				"Packet version:",
				"Plan — intent wireframe",
				"Architect — buildable wireframe",
				"Shell scope",
				"Header / navigation decision",
				"Footer decision",
				"WIDE buildable wireframe",
				"NARROW buildable wireframe",
				"Gelium behavior and state mapping",
				"Media direction and MEDIA-* references",
				"Decision: approved | changes-requested | declined | exception",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := repositoryFile(t, tc.path...)
			for _, required := range tc.want {
				if !strings.Contains(content, required) {
					t.Errorf("%s is missing %q", strings.Join(tc.path, "/"), required)
				}
			}
		})
	}
}

func TestServedAgentPacksMirrorPackageSources(t *testing.T) {
	for _, name := range []string{"llms.txt", "llms-ux.txt"} {
		packageSource := repositoryFile(t, "lib", name)
		servedProjection := repositoryFile(t, "site", "web", "static", name)
		if servedProjection != packageSource {
			t.Errorf("served %s must exactly mirror lib/%s", name, name)
		}
	}
}

func TestComponentImplementationSkillIsIndexed(t *testing.T) {
	skill := repositoryFile(t, "lib", "skills", "14-component-implementation.md")
	index := repositoryFile(t, "lib", "SKILLS.md")
	for _, required := range []string{
		"gelium-ui-component-implementation",
		"NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST",
		"styles_contract_test.go",
		"READY_FOR_INTEGRATION",
	} {
		if !strings.Contains(skill, required) {
			t.Errorf("component skill is missing %q", required)
		}
	}
	if !strings.Contains(index, "14-component-implementation.md") {
		t.Errorf("component guidance surface does not point to skill 14")
	}
}
