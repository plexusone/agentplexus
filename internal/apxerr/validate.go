package apxerr

import (
	"fmt"

	multiagent "github.com/agentplexus/multi-agent-spec/sdk/go"
)

// ValidateTeam validates a team spec and returns any warnings.
// It also normalizes the team in-place (e.g., setting nil slices to empty).
func ValidateTeam(team *multiagent.Team, teamFile string) []Warning {
	var warnings []Warning

	if team.Workflow == nil {
		return warnings
	}

	for i := range team.Workflow.Steps {
		step := &team.Workflow.Steps[i]

		// Check for missing DependsOn field
		if step.DependsOn == nil {
			warnings = append(warnings, Warning{
				Code:    MissingField,
				Message: fmt.Sprintf("Step %q missing depends_on field, defaulting to []", step.Name),
				Team:    team.Name,
				Step:    step.Name,
				Field:   "depends_on",
				File:    teamFile,
			})
			// Normalize: set to empty slice
			step.DependsOn = []string{}
		}
	}

	return warnings
}
