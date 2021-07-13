package handlers

import (
	"cloudiac/portal/apps"
	"cloudiac/portal/libs/ctrl"
	"cloudiac/portal/libs/ctx"
	"cloudiac/portal/models/forms"
)

type Vcs struct {
	ctrl.BaseController
}

// 创建vcs仓库
// @Tags Vcs仓库
// @Summary 创建vcs仓库
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param form formData forms.CreateVcsForm true "parameter"
// @Router /vcs [post]
// @Success 200 {object} ctx.JSONResult
func (Vcs) Create(c *ctx.GinRequestCtx) {
	form := &forms.CreateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.CreateVcs(c.ServiceCtx(), form))
}

// 查询vcs仓库
// @Tags Vcs仓库
// @Summary 查询vcs仓库
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param Iac-Org-Id header string true "组织ID"
// @Param form query forms.SearchVcsForm true "parameter"
// @Router /vcs [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]models.Vcs}}
func (Vcs) Search(c *ctx.GinRequestCtx) {
	form := &forms.SearchVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.SearchVcs(c.ServiceCtx(), form))
}


// 更新vcs仓库
// @Tags Vcs仓库
// @Summary 更新vcs仓库
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.UpdateVcsForm true "parameter"
// @Router /vcs/{vcsId} [put]
// @Success 200 {object} ctx.JSONResult{result=models.Vcs}
func (Vcs) Update(c *ctx.GinRequestCtx) {
	form := &forms.UpdateVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.UpdateVcs(c.ServiceCtx(), form))
}


// 删除Vcs 仓库
// @Tags Vcs仓库
// @Summary 删除vcs仓库
// @Accept multipart/form-data
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form formData forms.DeleteVcsForm true "patameter"
// @Param vcsId path string true "vcs的Id"
// @Router /vcs/{vcsId} [delete]
// @Success 200 {object} ctx.JSONResult
func (Vcs) Delete(c *ctx.GinRequestCtx) {
	form := &forms.DeleteVcsForm{}
	if err := c.Bind(form); err != nil {
		return
	}
	c.JSONResult(apps.DeleteVcs(c.ServiceCtx(), form))
}

func ListEnableVcs(c *ctx.GinRequestCtx) {
	c.JSONResult(apps.ListEnableVcs(c.ServiceCtx()))
}

// 列出Vcs地址下所有的代码仓库
// @Tags Vcs仓库
// @Summary 列出vcs地址下所有代码仓库
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form query forms.GetGitProjectsForm true "patameter"
// @Router /vcs/repo/search [get]
// @Success 200 {object} ctx.JSONResult{result=page.PageResp{list=[]vcsrv.Projects}}
func (Vcs) ListRepos(c *ctx.GinRequestCtx) {
	form := forms.GetGitProjectsForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepos(c.ServiceCtx(), &form))
}

// 列出代码仓库下所有分支
// @Tags Vcs仓库
// @Summary 列出代码仓库下所有分支
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form query forms.GetGitRevisionForm true "parameter"
// @Router /vcs/branch/search [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.Revision}
func (Vcs) ListBranches(c *ctx.GinRequestCtx) {
	form := forms.GetGitRevisionForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepoBranches(c.ServiceCtx(), &form))
}

// 列出代码仓库下tag
// @Tags Vcs仓库
// @Summary 列出代码仓库下所有分支
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form query forms.GetGitRevisionForm true "parameter"
// @Router /vcs/tag/search [get]
// @Success 200 {object} ctx.JSONResult{result=[]apps.Revision}
func (Vcs) ListTags(c *ctx.GinRequestCtx) {
	form := forms.GetGitRevisionForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.ListRepoTags(c.ServiceCtx(), &form))
}

// 列出代码仓库下Readme 文件内容
// @Tags Vcs仓库
// @Summary 列出代码仓库下 Readme 文件内容
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form query forms.GetReadmeForm true "parameter"
// @Router /vcs/readme [get]
// @Success 200 {object} ctx.JSONResult{result=string}
func (Vcs) GetReadmeContent(c *ctx.GinRequestCtx) {
	form := forms.GetReadmeForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.GetReadme(c.ServiceCtx(), &form))
}

// 列出代码仓库下包含.tfvars 的所有文件
// @Tags Vcs仓库
// @Summary 列出代码仓库下.tfvars 的所有文件
// @Accept application/x-www-form-urlencoded
// @Accept json
// @Produce json
// @Security AuthToken
// @Param form query forms.TemplateTfvarsSearchForm true "parameter"
// @Router /vcs/readme [get]
// @Success 200 {object} ctx.JSONResult{result=[]vcsrv.VcsIfaceOptions}
func TemplateTfvarsSearch(c *ctx.GinRequestCtx) {
	form := forms.TemplateTfvarsSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsTfVarsSearch(c.ServiceCtx(), &form))
}

// TemplateVariableSearch 查询云模板TF参数
// @Tags 云模板
// @Summary 云模板参数接口
// @Accept application/x-www-form-urlencoded
// @Param form query forms.TemplateVariableSearchForm true "parameter"
// @Router /template/variable/search [get]
// @Success 200 {object} ctx.JSONResult{result=[]services.TemplateVariable}
func TemplateVariableSearch(c *ctx.GinRequestCtx) {
	form := forms.TemplateVariableSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsVariableSearch(c.ServiceCtx(), &form))
}

//TemplatePlaybookSearch
// @Tags playbook列表查询
// @Summary  playbook列表接口
// @Accept application/x-www-form-urlencoded
// @Param form query forms.TemplatePlaybookSearchForm true "parameter"
// @router /template/playbook/search [get]
// @Success 200 {object} ctx.JSONResult{result=[]vcsrv.VcsIfaceOptions}
func TemplatePlaybookSearch(c *ctx.GinRequestCtx) {
	form := forms.TemplatePlaybookSearchForm{}
	if err := c.Bind(&form); err != nil {
		return
	}
	c.JSONResult(apps.VcsPlaybookSearch(c.ServiceCtx(), &form))
}
