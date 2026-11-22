package api

import (
	"strings"
	"time"

	"go-admin/dao"
	"go-admin/model"
	"go-admin/utils"

	"github.com/gin-gonic/gin"
)

// ListPosts 分页查询岗位
func ListPosts(c *gin.Context) {
	q := dao.PostQuery{
		PageNum:  pageNum(c),
		PageSize: pageSize(c),
		PostName: c.Query("postName"),
		PostCode: c.Query("postCode"),
		Status:   c.Query("status"),
	}
	posts, total, err := dao.ListPosts(q)
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkTable(c, posts, total)
}

// ListAllPosts 查询全部岗位（下拉用）
func ListAllPosts(c *gin.Context) {
	posts, err := dao.ListAllPosts()
	if err != nil {
		utils.Fail(c, 500, "查询失败")
		return
	}
	utils.OkData(c, posts)
}

// GetPost 查询岗位详情
func GetPost(c *gin.Context) {
	post, err := dao.GetPostByID(idParam(c))
	if err != nil {
		utils.Fail(c, 500, "岗位不存在")
		return
	}
	utils.OkData(c, post)
}

// PostReq 岗位请求体
type PostReq struct {
	PostID   uint   `json:"postId"`
	PostCode string `json:"postCode"`
	PostName string `json:"postName"`
	PostSort int    `json:"postSort"`
	Status   string `json:"status"`
	Remark   string `json:"remark"`
}

// CreatePost 新增岗位
func CreatePost(c *gin.Context) {
	var req PostReq
	if err := c.ShouldBindJSON(&req); err != nil || req.PostCode == "" || req.PostName == "" {
		utils.Fail(c, 500, "岗位编码和岗位名称不能为空")
		return
	}
	if count, _ := dao.CountPostByCode(req.PostCode); count > 0 {
		utils.Fail(c, 500, "岗位编码已存在")
		return
	}
	p := &model.Post{PostCode: req.PostCode, PostName: req.PostName, PostSort: req.PostSort, Status: req.Status, Remark: req.Remark}
	if p.Status == "" {
		p.Status = "0"
	}
	if err := dao.CreatePost(p); err != nil {
		utils.Fail(c, 500, "新增岗位失败")
		return
	}
	utils.Ok(c)
}

// UpdatePost 修改岗位
func UpdatePost(c *gin.Context) {
	var req PostReq
	if err := c.ShouldBindJSON(&req); err != nil || req.PostID == 0 {
		utils.Fail(c, 500, "请求参数错误")
		return
	}
	old, err := dao.GetPostByID(req.PostID)
	if err != nil {
		utils.Fail(c, 500, "岗位不存在")
		return
	}
	if req.PostCode != old.PostCode {
		if count, _ := dao.CountPostByCode(req.PostCode); count > 0 {
			utils.Fail(c, 500, "岗位编码已存在")
			return
		}
	}
	p := &model.Post{
		ID: old.ID, PostCode: req.PostCode, PostName: req.PostName,
		PostSort: req.PostSort, Status: req.Status, Remark: req.Remark,
		CreatedAt: old.CreatedAt, UpdatedAt: time.Now(),
	}
	if p.Status == "" {
		p.Status = "0"
	}
	if err := dao.UpdatePost(p); err != nil {
		utils.Fail(c, 500, "修改岗位失败")
		return
	}
	utils.Ok(c)
}

// DeletePost 删除岗位（支持逗号分隔批量，有用户引用时拦截）
func DeletePost(c *gin.Context) {
	ids := strings.Split(c.Param("id"), ",")
	var idList []uint
	for _, s := range ids {
		id := parseUint(s)
		if id == 0 {
			utils.Fail(c, 500, "参数错误")
			return
		}
		if count, _ := dao.CountUserByPost(id); count > 0 {
			utils.Fail(c, 500, "岗位已分配给用户，不允许删除")
			return
		}
		idList = append(idList, id)
	}
	if err := dao.DeletePosts(idList); err != nil {
		utils.Fail(c, 500, "删除岗位失败")
		return
	}
	utils.Ok(c)
}
