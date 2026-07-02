package cmd

import (
	"github.com/hurricanehrndz/respec/internal/artifact"
	"github.com/hurricanehrndz/respec/internal/store"
)

// artifactTarget pairs an artifact kind with its filename and path in a change dir.
type artifactTarget struct {
	name string
	kind artifact.Kind
	path string
}

// artifactTargets is the canonical research/spec/plan triple for a change dir,
// shared by lint and status so the set of artifacts is defined once.
func artifactTargets(changeDir string) []artifactTarget {
	return []artifactTarget{
		{"research.md", artifact.KindResearch, store.ResearchPath(changeDir)},
		{"spec.md", artifact.KindSpec, store.SpecPath(changeDir)},
		{"plan.md", artifact.KindPlan, store.PlanPath(changeDir)},
	}
}
