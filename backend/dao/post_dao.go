package dao

import (
	"go-admin/model"
)

// PostQuery 岗位查询条件
type PostQuery struct {
	PageNum  int
	PageSize int
	PostName string
	PostCode string
	Status   string
}

// ListPosts 分页查询岗位
func ListPosts(q PostQuery) ([]model.Post, int64, error) {
	var posts []model.Post
	var total int64
	db := DB.Model(&model.Post{})
	if q.PostName != "" {
		db = db.Where(likeExprEsc("post_name"), "%"+escapeLike(q.PostName)+"%")
	}
	if q.PostCode != "" {
		db = db.Where(likeExprEsc("post_code"), "%"+escapeLike(q.PostCode)+"%")
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Order("post_sort asc, id asc").Offset((q.PageNum - 1) * q.PageSize).Limit(q.PageSize).Find(&posts).Error
	return posts, total, err
}

// ListAllPosts 查询全部岗位（下拉用，仅正常）
func ListAllPosts() ([]model.Post, error) {
	var posts []model.Post
	err := DB.Where("status = '0'").Order("post_sort asc, id asc").Find(&posts).Error
	return posts, err
}

// GetPostByID 按 ID 查询岗位
func GetPostByID(id uint) (*model.Post, error) {
	var p model.Post
	err := DB.First(&p, id).Error
	return &p, err
}

// CountPostByCode 统计相同岗位编码数量
func CountPostByCode(code string) (int64, error) {
	var count int64
	err := DB.Model(&model.Post{}).Where("post_code = ?", code).Count(&count).Error
	return count, err
}

// CreatePost 新增岗位
func CreatePost(p *model.Post) error {
	return DB.Create(p).Error
}

// UpdatePost 更新岗位
func UpdatePost(p *model.Post) error {
	return DB.Save(p).Error
}

// DeletePost 删除岗位
func DeletePost(id uint) error {
	return DB.Delete(&model.Post{}, id).Error
}

// DeletePosts 批量删除岗位
func DeletePosts(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return DB.Where("id IN ?", ids).Delete(&model.Post{}).Error
}

// CountUserByPost 统计绑定该岗位的用户数量
func CountUserByPost(postID uint) (int64, error) {
	var count int64
	err := DB.Table("sys_user_post").Where("post_id = ?", postID).Count(&count).Error
	return count, err
}
