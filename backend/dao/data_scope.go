package dao

import (
	"go-admin/model"
)

// DataScopeDeptIDs 计算当前用户的数据权限可见部门 ID 集合
// 返回 (ids, limited)：limited=false 表示不限（全部数据权限）；limited=true 时 ids 为可见部门集合（含祖先链）
// 规则对齐 RuoYi：1 全部 / 2 自定义 / 3 本部门 / 4 本部门及以下 / 5 仅本人
func DataScopeDeptIDs(userID uint) ([]uint, bool) {
	if IsAdmin(userID) {
		return nil, false
	}
	scope, roleID, deptID, _ := dataScopeOf(userID)
	switch scope {
	case "1", "5": // 全部；仅本人对部门无限制意义
		return nil, false
	case "2":
		var deptIDs []uint
		if err := DB.Table("sys_role_dept").Where("role_id = ?", roleID).Pluck("dept_id", &deptIDs).Error; err != nil || len(deptIDs) == 0 {
			return []uint{0}, true
		}
		return withSubAndAncestors(deptIDs), true
	case "3":
		return withAncestors([]uint{deptID}), true
	case "4":
		return withSubAndAncestors([]uint{deptID}), true
	}
	return nil, false
}

// DataScopeUserCond 计算用户列表的数据权限 SQL 条件
// 对齐 RuoYi：2 自定义=精确部门集合；3 本部门；4 本部门及以下（不含祖先）；5 仅本人
func DataScopeUserCond(userID uint) (string, []interface{}, bool) {
	if IsAdmin(userID) {
		return "", nil, false
	}
	scope, roleID, deptID, username := dataScopeOf(userID)
	switch scope {
	case "1":
		return "", nil, false
	case "2":
		var deptIDs []uint
		if err := DB.Table("sys_role_dept").Where("role_id = ?", roleID).Pluck("dept_id", &deptIDs).Error; err != nil || len(deptIDs) == 0 {
			return "1 = 0", nil, true
		}
		return "dept_id IN ?", []interface{}{deptIDs}, true
	case "3":
		return "dept_id = ?", []interface{}{deptID}, true
	case "4":
		ids := collectSubDeptIDs(deptID)
		return "dept_id IN ?", []interface{}{ids}, true
	case "5":
		return "username = ?", []interface{}{username}, true
	}
	return "", nil, false
}

// UserInDataScope 判断目标用户是否落在当前操作者的数据权限范围内
// 超管或未做数据限制时恒为 true；查询异常时按无权处理
func UserInDataScope(operatorID, targetID uint) bool {
	if IsAdmin(operatorID) {
		return true
	}
	scopeSQL, scopeArgs, limited := DataScopeUserCond(operatorID)
	if !limited {
		return true
	}
	query := DB.Model(&model.User{}).Where("id = ?", targetID)
	if scopeSQL != "" {
		query = query.Where(scopeSQL, scopeArgs...)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

// dataScopeOf 取用户的数据权限配置（多角色取最宽松：任一全部则全部；否则取最小数字）
// 返回的 roleID 始终是「产生该 data_scope 的角色」，自定义范围据此查 sys_role_dept
func dataScopeOf(userID uint) (scope string, roleID, deptID uint, username string) {
	var user model.User
	if err := DB.Select("dept_id, username").First(&user, userID).Error; err != nil {
		// 查询失败时按最严格范围处理，避免误放开全部数据
		return "5", 0, 0, ""
	}
	var roles []struct {
		ID        uint   `gorm:"column:id"`
		DataScope string `gorm:"column:data_scope"`
	}
	if err := DB.Raw(`SELECT r.id, r.data_scope FROM sys_role r
		JOIN sys_user_role ur ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.status = '0' ORDER BY r.sort, r.id`, userID).Scan(&roles).Error; err != nil || len(roles) == 0 {
		// 无有效角色（未分配或角色已停用）时按最严格的「仅本人」处理，避免误放开全部数据
		return "5", 0, user.DeptID, user.Username
	}
	// 任一角色为全部 → 全部
	for _, r := range roles {
		if r.DataScope == "1" {
			return "1", 0, user.DeptID, user.Username
		}
	}
	// 取最小数字（最宽松），并记录它所属的角色
	best := roles[0]
	for _, r := range roles[1:] {
		if r.DataScope < best.DataScope {
			best = r
		}
	}
	return best.DataScope, best.ID, user.DeptID, user.Username
}

// collectSubDeptIDs 递归收集部门及其全部子部门 ID
func collectSubDeptIDs(root uint) []uint {
	var depts []model.Dept
	_ = DB.Select("id, parent_id").Find(&depts).Error
	ids := []uint{root}
	queue := []uint{root}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, d := range depts {
			if d.ParentID == cur {
				ids = append(ids, d.ID)
				queue = append(queue, d.ID)
			}
		}
	}
	return ids
}

// collectAncestorIDs 递归收集部门祖先链 ID
func collectAncestorIDs(deptID uint, depts []model.Dept, seen map[uint]bool) []uint {
	var ids []uint
	for _, d := range depts {
		if d.ID == deptID {
			parent := d.ParentID
			for parent != 0 && !seen[parent] {
				seen[parent] = true
				ids = append(ids, parent)
				for _, p := range depts {
					if p.ID == parent {
						parent = p.ParentID
						break
					}
				}
			}
			break
		}
	}
	return ids
}

// withAncestors 部门 + 祖先链
func withAncestors(deptIDs []uint) []uint {
	var depts []model.Dept
	_ = DB.Select("id, parent_id").Find(&depts).Error
	seen := map[uint]bool{}
	for _, id := range deptIDs {
		seen[id] = true
	}
	for _, id := range deptIDs {
		for _, a := range collectAncestorIDs(id, depts, seen) {
			seen[a] = true
		}
	}
	return keysOf(seen)
}

// withSubAndAncestors 部门 + 子部门 + 祖先链
func withSubAndAncestors(deptIDs []uint) []uint {
	var depts []model.Dept
	_ = DB.Select("id, parent_id").Find(&depts).Error
	seen := map[uint]bool{}
	queue := deptIDs
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if seen[cur] {
			continue
		}
		seen[cur] = true
		for _, d := range depts {
			if d.ParentID == cur {
				queue = append(queue, d.ID)
			}
		}
	}
	for _, id := range deptIDs {
		for _, a := range collectAncestorIDs(id, depts, seen) {
			seen[a] = true
		}
	}
	return keysOf(seen)
}

func keysOf(m map[uint]bool) []uint {
	keys := make([]uint, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// FilterDeptTree 按可见部门 ID 集合剪枝部门树（保留可见节点与祖先链）
func FilterDeptTree(tree []model.Dept, visible map[uint]bool) []model.Dept {
	var walk func(nodes []model.Dept) []model.Dept
	walk = func(nodes []model.Dept) []model.Dept {
		res := make([]model.Dept, 0, len(nodes))
		for i := range nodes {
			if !visible[nodes[i].ID] {
				continue
			}
			nodes[i].Children = walk(nodes[i].Children)
			res = append(res, nodes[i])
		}
		return res
	}
	return walk(tree)
}

// VisibleDeptMap 将可见 ID 列表转为 map
func VisibleDeptMap(ids []uint) map[uint]bool {
	m := map[uint]bool{}
	for _, id := range ids {
		m[id] = true
	}
	return m
}
