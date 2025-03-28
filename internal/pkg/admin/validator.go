package admin

import (
	"errors"
	"strings"
)

// ProjectConfigValidator loads available projects from config
type ProjectConfigValidator struct {
	projects map[string]bool
}

// NewProjectConfigValidator creates project validator, reads available projects from config
func NewProjectConfigValidator(projects string) (*ProjectConfigValidator, error) {
	if projects == "" {
		return nil, errors.New("no projects provided")
	}
	res := ProjectConfigValidator{}
	res.projects = make(map[string]bool)
	for _, p := range strings.Split(projects, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			res.projects[p] = true
		}
	}
	if len(res.projects) == 0 {
		return nil, errors.New("no projects provided")
	}
	return &res, nil
}

// Check tests if project is available
func (pv *ProjectConfigValidator) Check(pr string) bool {
	return pv.projects[pr]
}

// Projects returns all supported projects
func (pv *ProjectConfigValidator) Projects() []string {
	res := make([]string, len(pv.projects))
	i := 0
	for k := range pv.projects {
		res[i] = k
		i++
	}
	return res
}
