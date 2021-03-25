package apps

import (
	"cloudiac/consts/e"
	"cloudiac/libs/ctx"
	"cloudiac/models"
	"cloudiac/models/forms"
	"cloudiac/services"
	"cloudiac/utils"
	"fmt"
	"net/http"
)

func CreateOrganization(c *ctx.ServiceCtx, form *forms.CreateOrganizationForm) (*models.Organization, e.Error) {
	c.AddLogField("action", fmt.Sprintf("create org %s", form.Name))

	// todo: 验证vcs_provider信息是否有效，一期固定使用内部gitlab，暂不实现

	guid := utils.GenGuid("org")
	org, err := services.CreateOrganization(c.DB(), models.Organization{
		Name:        form.Name,
		Guid:        guid,
		Description: form.Description,
		UserId:      c.UserId,
		VcsType:     form.VcsType,
		VcsVersion:  form.VcsVersion,
		VcsAuthInfo: form.VcsAuthInfo,
	})
	if err != nil {
		return nil, e.AutoNew(err, e.DBError)
	}
	return org, nil
}

type searchOrganizationResp struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	Guid        string `json:"guid"`
	Description string `json:"description"`
	UserId      uint   `json:"userId"`
	Status      string `json:"status"`
	Creator     string `json:"creator"`
}

func (m *searchOrganizationResp) TableName() string {
	return models.Organization{}.TableName()
}

func SearchOrganization(c *ctx.ServiceCtx, form *forms.SearchOrganizationForm) (interface{}, e.Error) {
	query := services.QueryOrganization(c.DB())
	if c.IsAdmin == true {
		if form.Status != "" {
			query = query.Where("status = ?", form.Status)
		}
	} else {
		query = query.Where("status = 'enable'")
		orgIds, er := services.GetOrgIdsByUser(c.DB(), c.UserId)
		if er != nil {
			return nil, e.New(e.DBError, er)
		}
		query = query.Where("id in (?)", orgIds)
	}

	if form.Q != "" {
		qs := "%" + form.Q + "%"
		query = query.Where("name LIKE ?", qs)
	}

	query = query.Order("created_at DESC")
	rs, _ := getPage(query, form, &searchOrganizationResp{})
	return rs, nil
}

func UpdateOrganization(c *ctx.ServiceCtx, form *forms.UpdateOrganizationForm) (org *models.Organization, err e.Error) {
	c.AddLogField("action", fmt.Sprintf("update org %d", form.Id))
	if form.Id == 0 {
		return nil, e.New(e.BadRequest, fmt.Errorf("missing 'id'"))
	}

	attrs := models.Attrs{}
	if form.HasKey("name") {
		attrs["name"] = form.Name
	}

	if form.HasKey("description") {
		attrs["description"] = form.Description
	}

	if form.HasKey("vcsAuthInfo") {
		attrs["vcs_auth_info"] = form.VcsAuthInfo
	}

	org, err = services.UpdateOrganization(c.DB(), form.Id, attrs)
	return
}

func ChangeOrgStatus(c *ctx.ServiceCtx, form *forms.DisableOrganizationForm) (interface{}, e.Error) {
	org, er := services.GetOrganizationById(c.DB(), form.Id)
	if er != nil {
		return nil, er
	}

	if org.Status == form.Status {
		return org, nil
	} else if form.Status != models.OrgEnable && form.Status != models.OrgDisable {
		return nil, e.New(e.OrganizationInvalidStatus)
	}

	org, err := services.UpdateOrganization(c.DB(), form.Id, models.Attrs{"status": form.Status})
	if err != nil {
		return nil, err
	}

	return org, nil
}

func ListOrganizationRepos(c *ctx.ServiceCtx, orgId int, searchKey string) (interface{}, e.Error) {
	repos, err := services.ListOrganizationReposById(c.DB(), orgId, searchKey)
	if err != nil {
		return nil, nil
	}
	return repos, nil
}

func ListRepositoryBranches(c *ctx.ServiceCtx, orgId int, repoId int) (interface{}, e.Error) {
	branches, err := services.ListRepositoryBranches(c.DB(), orgId, repoId)
	if err != nil {
		return nil, nil
	}
	return branches, nil
}

func GetReadmeContent(c *ctx.ServiceCtx, orgId int, repoId int, branch string) (interface{}, e.Error) {
	content, err := services.GetReadmeContent(c.DB(), orgId, repoId, branch)
	if err != nil {
		return nil, nil
	}
	return content, nil
}

type organizationDetailResp struct {
	models.Organization
	Creator string
}

func OrganizationDetail(c *ctx.ServiceCtx, form *forms.DetailOrganizationForm) (resp interface{}, er e.Error) {
	orgIds, err := services.GetOrgIdsByUser(c.DB(), c.UserId)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}

	if utils.UintIsContain(orgIds, form.Id) == false && c.IsAdmin == false {
		return nil, nil
	}
	org, err := services.GetOrganizationById(c.DB(), form.Id)
	if err != nil {
		return nil, e.New(e.DBError, http.StatusInternalServerError, err)
	}
	user, err := services.GetUserById(c.DB(), org.UserId)
	if err != nil {
		return nil, e.New(e.DBError, err)
	}
	var o = organizationDetailResp{
		Organization: *org,
		Creator: user.Name,
	}

	return o, nil
}
