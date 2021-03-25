package services

import (
	"encoding/json"
	"fmt"

	//"errors"
	"cloudiac/consts/e"
	"cloudiac/libs/db"
	"cloudiac/models"

	"github.com/xanzy/go-gitlab"
)

func CreateOrganization(tx *db.Session, org models.Organization) (*models.Organization, e.Error) {
	if err := models.Create(tx, &org); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.OrganizationAlreadyExists, err)
		}
		return nil, e.New(e.DBError, err)
	}

	return &org, nil
}

func UpdateOrganization(tx *db.Session, id uint, attrs models.Attrs) (org *models.Organization, re e.Error) {
	org = &models.Organization{}
	if _, err := models.UpdateAttr(tx.Where("id = ?", id), &models.Organization{}, attrs); err != nil {
		if e.IsDuplicate(err) {
			return nil, e.New(e.OrganizationAliasDuplicate)
		}
		return nil, e.New(e.DBError, fmt.Errorf("update org error: %v", err))
	}
	if err := tx.Where("id = ?", id).First(org); err != nil {
		return nil, e.New(e.DBError, fmt.Errorf("query org error: %v", err))
	}
	return
}

func DeleteOrganization(tx *db.Session, id uint) e.Error {
	if _, err := tx.Where("id = ?", id).Delete(&models.Organization{}); err != nil {
		return e.New(e.DBError, fmt.Errorf("delete org error: %v", err))
	}
	return nil
}

func GetOrganizationById(tx *db.Session, id uint) (*models.Organization, e.Error) {
	o := models.Organization{}
	if err := tx.Where("id = ?", id).First(&o); err != nil {
		if e.IsRecordNotFound(err) {
			return nil, e.New(e.OrganizationNotExists, err)
		}
		return nil, e.New(e.DBError, err)
	}
	return &o, nil
}

func GetOrganizationNotExistsByName(tx *db.Session, name string) (*models.Organization, error) {
	o := models.Organization{}
	if err := tx.Where("name = ?", name).First(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func GetUserByAlias(tx *db.Session, alias string) (*models.Organization, error) {
	o := models.Organization{}
	if err := tx.Where("alias = ?", alias).First(&o); err != nil {
		return nil, err
	}
	return &o, nil
}

func FindOrganization(query *db.Session) (orgs []*models.Organization, err error) {
	err = query.Find(&orgs)
	return
}

func QueryOrganization(query *db.Session) *db.Session {
	return query.Model(&models.Organization{})
}

func ListOrganizationReposById(tx *db.Session, orgId int, searchKey string) (repos []*gitlab.Project, err error) {
	org := models.Organization{}
	if err := tx.Where("id = ?", orgId).First(&org); err != nil {
		return nil, err
	}
	var dat map[string]string
	vcsAuthInfo := org.VcsAuthInfo
	if err := json.Unmarshal([]byte(vcsAuthInfo), &dat); err != nil {
		return nil, err
	}
	git, err := gitlab.NewClient(dat["token"], gitlab.WithBaseURL(dat["url"]+"/api/v4"))
	opt := &gitlab.ListProjectsOptions{}
	if searchKey != "" {
		opt = &gitlab.ListProjectsOptions{Search: gitlab.String(searchKey)}
	}
	projects, _, err := git.Projects.ListProjects(opt)
	return projects, err
}

func ListRepositoryBranches(tx *db.Session, orgId int, repoId int) (branches []*gitlab.Branch, err error) {
	org := models.Organization{}
	if err := tx.Where("id = ?", orgId).First(&org); err != nil {
		return nil, err
	}
	var dat map[string]string
	vcsAuthInfo := org.VcsAuthInfo
	if err := json.Unmarshal([]byte(vcsAuthInfo), &dat); err != nil {
		return nil, err
	}
	git, err := gitlab.NewClient(dat["token"], gitlab.WithBaseURL(dat["url"]+"/api/v4"))
	if err != nil {
		return nil, err
	}
	opt := &gitlab.ListBranchesOptions{}
	branches, _, er := git.Branches.ListBranches(repoId, opt)
	return branches, er
}

func GetReadmeContent(tx *db.Session, orgId int, repoId int, branch string) (content models.FileContent, err error) {
	org := models.Organization{}
	content = models.FileContent{
		Content: "",
	}
	if err := tx.Where("id = ?", orgId).First(&org); err != nil {
		return content, err
	}
	var dat map[string]string
	vcsAuthInfo := org.VcsAuthInfo
	if err := json.Unmarshal([]byte(vcsAuthInfo), &dat); err != nil {
		return content, err
	}
	git, err := gitlab.NewClient(dat["token"], gitlab.WithBaseURL(dat["url"]+"/api/v4"))
	if err != nil {
		return content, err
	}
	opt := &gitlab.GetRawFileOptions{Ref: gitlab.String(branch)}
	row, _, err := git.RepositoryFiles.GetRawFile(repoId, "README.md", opt)
	if err != nil {
		return content, err
	}

	fmt.Println(content)
	res := models.FileContent{
		Content: string(row[:]),
	}
	return res, nil
}
