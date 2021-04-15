package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/libs/page"
	"cloudiac/models/forms"
	"cloudiac/services"
	"encoding/json"
	"time"
)

type Projects struct {
	ID             int        `json:"id"`
	Description    string     `json:"description"`
	DefaultBranch  string     `json:"default_branch"`
	SSHURLToRepo   string     `json:"ssh_url_to_repo"`
	HTTPURLToRepo  string     `json:"http_url_to_repo"`
	Name           string     `json:"name"`
	LastActivityAt *time.Time `json:"last_activity_at,omitempty"`
}

func ListOrganizationRepos(c *ctx.ServiceCtx, form *forms.GetGitProjectsForm) (interface{}, e.Error) {
	projects, total, err := services.ListOrganizationReposById(c.DB(), c.OrgId, form)
	if err != nil {
		return nil, err
	}

	jsonProjects, er := json.Marshal(projects)
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}
	repos := make([]*Projects, 0)
	er = json.Unmarshal(jsonProjects, &repos)
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}

	return page.PageResp{
		Total:    int64(total),
		PageSize: form.CurrentPage(),
		List:     repos,
	}, nil
}

type Branches struct {
	Name string `json:"name"`
}

func ListRepositoryBranches(c *ctx.ServiceCtx, form *forms.GetGitBranchesForm) (brans []*Branches, err e.Error) {
	branches, err := services.ListRepositoryBranches(c.DB(), c.OrgId, form.RepoId)
	if err != nil {
		return nil, err
	}

	jsonBranches, er := json.Marshal(branches)
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}

	er = json.Unmarshal(jsonBranches, &brans)
	if er != nil {
		return nil, e.New(e.JSONParseError, er)
	}
	return brans, nil
}

func GetReadmeContent(c *ctx.ServiceCtx, form *forms.GetReadmeForm) (interface{}, e.Error) {
	content, err := services.GetReadmeContent(c.DB(), c.OrgId, form)
	if err != nil {
		return nil, nil
	}
	return content, nil
}