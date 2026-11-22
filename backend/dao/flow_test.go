package dao

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-admin/config"
	"go-admin/model"
	"go-admin/utils"
)

// 工作流引擎测试：独立临时 SQLite + 全量种子
// 覆盖：发起/或签/会签/通过推进/驳回重提/转办/加签/抄送/撤回/终止/定义CRUD 约束

var (
	flowTestAdmin model.User // admin 用户（种子：id=1，部门100，昵称管理员，角色1）
	flowTestUser  model.User // 测试审批人（动态创建，绑定角色1）
	flowTestDefID uint       // 通用审批（user=admin 或签）
	flowTestDef2  uint       // 请假审批（dept_leader → role 1）
)

func TestMain(m *testing.M) {
	dbPath := filepath.Join(os.TempDir(), fmt.Sprintf("go-admin-flow-test-%d.db", time.Now().UnixNano()))
	config.Cfg = config.Config{
		Server:  config.ServerConfig{Port: 8080},
		JWT:     config.JWTConfig{Secret: "flow-test-secret", Expire: 30},
		Captcha: config.CaptchaConfig{Enabled: false},
		Database: config.DatabaseConfig{
			Type: "sqlite",
			Path: dbPath,
		},
	}
	if err := InitDB(config.Cfg.Database); err != nil {
		log.Fatalf("初始化测试数据库失败: %v", err)
	}
	// 种子：admin 与两条示例流程
	admin, err := GetUserByUsername("admin")
	if err != nil {
		log.Fatalf("种子用户 admin 缺失: %v", err)
	}
	flowTestAdmin = *admin
	// 创建测试审批人（绑定超级管理员角色，供会签/转办/加签/抄送使用）
	hash, _ := utils.HashPassword("123456")
	flowTestUser = model.User{
		DeptID: 101, Username: "flowtest1", Password: hash, Nickname: "测试审批人",
		Email: "flow@example.com", Phone: "13900000000", Status: "0", Remark: "工作流测试用户",
	}
	if err := DB.Create(&flowTestUser).Error; err != nil {
		log.Fatalf("创建测试用户失败: %v", err)
	}
	role1 := model.Role{}
	if err := DB.First(&role1, 1).Error; err != nil {
		log.Fatalf("角色1缺失: %v", err)
	}
	if err := DB.Model(&flowTestUser).Association("Roles").Append(&role1); err != nil {
		log.Fatalf("绑定测试用户角色失败: %v", err)
	}
	// 示例流程 ID
	var defs []model.FlowDef
	if err := DB.Order("flow_def_id ASC").Find(&defs).Error; err != nil {
		log.Fatalf("读取示例流程失败: %v", err)
	}
	flowTestDefID = defs[0].ID // 通用审批
	flowTestDef2 = defs[1].ID  // 请假审批
	code := m.Run()
	_ = os.Remove(dbPath)
	_ = os.Remove(dbPath + "-wal")
	_ = os.Remove(dbPath + "-shm")
	os.Exit(code)
}

// mustStart 发起并返回实例 ID
func mustStart(t *testing.T, defID uint, title string) uint {
	t.Helper()
	id, err := StartFlowInstance(flowTestAdmin, FlowStartReq{FlowDefID: defID, Title: title, Content: "测试内容"})
	if err != nil {
		t.Fatalf("发起失败: %v", err)
	}
	return id
}

// todoOf 取指定实例中操作人的待办任务
func todoOf(t *testing.T, instID uint, userID uint) []model.FlowTask {
	t.Helper()
	var tasks []model.FlowTask
	if err := DB.Where("instance_id = ? AND assignee_id = ? AND status = '0'", instID, userID).Find(&tasks).Error; err != nil {
		t.Fatalf("查询待办失败: %v", err)
	}
	return tasks
}

func instByID(t *testing.T, id uint) model.FlowInstance {
	t.Helper()
	var inst model.FlowInstance
	if err := DB.First(&inst, id).Error; err != nil {
		t.Fatalf("查询实例失败: %v", err)
	}
	return inst
}

func TestFlowStartOrSignApprove(t *testing.T) {
	// 通用审批（或签，指定 admin）：发起 → admin 待办 → 通过 → 实例通过
	instID := mustStart(t, flowTestDefID, "通用审批-或签")
	todos := todoOf(t, instID, flowTestAdmin.ID)
	if len(todos) != 1 {
		t.Fatalf("期望 1 条待办，实际 %d", len(todos))
	}
	if err := ApproveFlowTask(todos[0].ID, flowTestAdmin.ID, "同意"); err != nil {
		t.Fatalf("审批失败: %v", err)
	}
	inst := instByID(t, instID)
	if inst.Status != "2" {
		t.Fatalf("期望实例通过(2)，实际 %s", inst.Status)
	}
	// 重复审批应被拒绝
	if err := ApproveFlowTask(todos[0].ID, flowTestAdmin.ID, "again"); err == nil {
		t.Fatal("重复审批应报错")
	}
	// 非本人审批应被拒绝
	inst2 := mustStart(t, flowTestDefID, "通用审批-越权")
	t2 := todoOf(t, inst2, flowTestAdmin.ID)[0]
	if err := ApproveFlowTask(t2.ID, flowTestUser.ID, "越权"); err == nil {
		t.Fatal("越权审批应报错")
	}
	// 抄送生成
	inst3, err := StartFlowInstance(flowTestAdmin, FlowStartReq{
		FlowDefID: flowTestDefID, Title: "带抄送", Content: "x", CCUserIDs: []uint{flowTestUser.ID},
	})
	if err != nil {
		t.Fatalf("发起失败: %v", err)
	}
	var cc []model.FlowTask
	if err := DB.Where("instance_id = ? AND task_type = '2' AND assignee_id = ?", inst3, flowTestUser.ID).Find(&cc).Error; err != nil {
		t.Fatalf("查询抄送失败: %v", err)
	}
	if len(cc) != 1 {
		t.Fatalf("期望 1 条抄送，实际 %d", len(cc))
	}
	if err := MarkCcRead(cc[0].ID, flowTestUser.ID); err != nil {
		t.Fatalf("抄送已读失败: %v", err)
	}
}

func TestFlowMultiLevelAndDeptLeader(t *testing.T) {
	// 请假审批（dept_leader → role 1）：admin 部门100 负责人=管理员→admin
	instID := mustStart(t, flowTestDef2, "请假审批-两级")
	// 节点1：部门负责人路由到 admin
	t1 := todoOf(t, instID, flowTestAdmin.ID)
	if len(t1) != 1 {
		t.Fatalf("节点1期望 1 条待办，实际 %d", len(t1))
	}
	if t1[0].NodeOrder != 1 {
		t.Fatalf("期望节点顺序 1，实际 %d", t1[0].NodeOrder)
	}
	if err := ApproveFlowTask(t1[0].ID, flowTestAdmin.ID, "部门负责人同意"); err != nil {
		t.Fatalf("节点1审批失败: %v", err)
	}
	// 节点2：角色1 → admin（待办推进）
	t2 := todoOf(t, instID, flowTestAdmin.ID)
	if len(t2) != 1 {
		t.Fatalf("节点2期望 1 条待办，实际 %d", len(t2))
	}
	if t2[0].NodeOrder != 2 {
		t.Fatalf("期望节点顺序 2，实际 %d", t2[0].NodeOrder)
	}
	if err := ApproveFlowTask(t2[0].ID, flowTestAdmin.ID, "终审同意"); err != nil {
		t.Fatalf("节点2审批失败: %v", err)
	}
	if instByID(t, instID).Status != "2" {
		t.Fatal("期望实例通过")
	}
}

func TestFlowRejectReapplyAndBackPrev(t *testing.T) {
	// 请假审批节点1 驳回（reject_to=1 → 退回上一节点；节点1无上一节点 → 回发起人）
	instID := mustStart(t, flowTestDef2, "请假审批-驳回")
	t1 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := RejectFlowTask(t1.ID, flowTestAdmin.ID, "材料不全"); err != nil {
		t.Fatalf("驳回失败: %v", err)
	}
	inst := instByID(t, instID)
	if inst.Status != "3" || inst.RejectFromNodeID != t1.NodeID {
		t.Fatalf("期望驳回回发起人(3, node=%d)，实际 status=%s from=%d", t1.NodeID, inst.Status, inst.RejectFromNodeID)
	}
	// 重新提交 → 从驳回节点恢复
	if err := ReapplyFlowInstance(instID, flowTestAdmin.ID, "请假审批-重新提交", "补全材料"); err != nil {
		t.Fatalf("重新提交失败: %v", err)
	}
	if instByID(t, instID).Status != "1" {
		t.Fatal("期望重新提交后运行中")
	}
	t1b := todoOf(t, instID, flowTestAdmin.ID)
	if len(t1b) != 1 || t1b[0].NodeOrder != 1 {
		t.Fatalf("期望从节点1恢复，实际 %+v", t1b)
	}
	// 走完全流程
	if err := ApproveFlowTask(t1b[0].ID, flowTestAdmin.ID, "同意"); err != nil {
		t.Fatalf("审批失败: %v", err)
	}
	t2 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := ApproveFlowTask(t2.ID, flowTestAdmin.ID, "同意"); err != nil {
		t.Fatalf("终审失败: %v", err)
	}
	if instByID(t, instID).Status != "2" {
		t.Fatal("期望最终通过")
	}
}

func TestFlowCounterSign(t *testing.T) {
	// 会签流程：节点 role 1（admin + 测试审批人 2 人），需两人都通过才推进到下一节点（user=admin）
	defID, err := SaveFlowDef(FlowDefInput{
		DefName: "会签测试", DefKey: "cntsign", Category: "测试",
		Nodes: []FlowNodeInput{
			{NodeOrder: 1, NodeName: "会签步骤", ApproverType: "role", ApproverIDs: "1", NodeMode: "any"},
			{NodeOrder: 2, NodeName: "汇总审批", ApproverType: "user", ApproverIDs: fmt.Sprintf("%d", flowTestAdmin.ID), NodeMode: "one"},
		},
	}, "admin")
	if err != nil {
		t.Fatalf("创建会签流程失败: %v", err)
	}
	instID := mustStart(t, defID, "会签-测试")
	todos := todoOf(t, instID, flowTestAdmin.ID)
	if len(todos) != 1 {
		t.Fatalf("admin 期望 1 条待办，实际 %d", len(todos))
	}
	// 只有 admin 通过 → 未推进（会签等测试审批人）
	if err := ApproveFlowTask(todos[0].ID, flowTestAdmin.ID, "admin 同意"); err != nil {
		t.Fatalf("审批失败: %v", err)
	}
	if instByID(t, instID).Status != "1" || instByID(t, instID).CurrentOrder != 1 {
		t.Fatal("会签未全过不应推进")
	}
	u2 := todoOf(t, instID, flowTestUser.ID)
	if len(u2) != 1 {
		t.Fatalf("测试审批人期望 1 条待办，实际 %d", len(u2))
	}
	if err := ApproveFlowTask(u2[0].ID, flowTestUser.ID, "会签同意"); err != nil {
		t.Fatalf("会签审批失败: %v", err)
	}
	// 推进到节点2（admin）
	t2 := todoOf(t, instID, flowTestAdmin.ID)
	if len(t2) != 1 || t2[0].NodeOrder != 2 {
		t.Fatalf("期望推进到节点2，实际 %+v", t2)
	}
	if err := ApproveFlowTask(t2[0].ID, flowTestAdmin.ID, "汇总通过"); err != nil {
		t.Fatalf("汇总审批失败: %v", err)
	}
	if instByID(t, instID).Status != "2" {
		t.Fatal("期望会签全过+汇总通过后实例通过")
	}
}

func TestFlowTransferAndAddSign(t *testing.T) {
	// 转办：通用审批 admin 待办转给测试审批人，原任务置转出，被转办人可审批
	instID := mustStart(t, flowTestDefID, "转办-测试")
	t1 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := TransferFlowTask(t1.ID, flowTestAdmin.ID, flowTestUser.ID, "请代为处理"); err != nil {
		t.Fatalf("转办失败: %v", err)
	}
	if st := todoOf(t, instID, flowTestAdmin.ID); len(st) != 0 {
		t.Fatal("转办后原审批人不保留待办")
	}
	u2 := todoOf(t, instID, flowTestUser.ID)
	if len(u2) != 1 {
		t.Fatalf("被转办人期望 1 条待办，实际 %d", len(u2))
	}
	if err := ApproveFlowTask(u2[0].ID, flowTestUser.ID, "代审同意"); err != nil {
		t.Fatalf("被转办人审批失败: %v", err)
	}
	if instByID(t, instID).Status != "2" {
		t.Fatal("转办后通过应完成实例")
	}
	// 加签：admin 待办加签给测试审批人，两人并行；或签任一过即完成
	inst2 := mustStart(t, flowTestDefID, "加签-测试")
	t2 := todoOf(t, inst2, flowTestAdmin.ID)[0]
	if err := AddSignFlowTask(t2.ID, flowTestAdmin.ID, flowTestUser.ID, "请协同审批"); err != nil {
		t.Fatalf("加签失败: %v", err)
	}
	u2b := todoOf(t, inst2, flowTestUser.ID)
	if len(u2b) != 1 || u2b[0].TaskType != "3" {
		t.Fatalf("加签任务异常: %+v", u2b)
	}
	// 或签：加签人通过即完成（原任务撤销）
	if err := ApproveFlowTask(u2b[0].ID, flowTestUser.ID, "加签通过"); err != nil {
		t.Fatalf("加签审批失败: %v", err)
	}
	if instByID(t, inst2).Status != "2" {
		t.Fatal("或签+加签后任一通过应完成实例")
	}
}

func TestFlowWithdrawTerminateAndGuard(t *testing.T) {
	// 撤回
	instID := mustStart(t, flowTestDefID, "撤回-测试")
	if err := WithdrawFlowInstance(instID, flowTestAdmin.ID); err != nil {
		t.Fatalf("撤回失败: %v", err)
	}
	if instByID(t, instID).Status != "4" {
		t.Fatal("期望撤回(4)")
	}
	// 撤回后不可再审批
	t1 := todoOf(t, instID, flowTestAdmin.ID)
	if len(t1) != 0 {
		t.Fatal("撤回后不应有待办")
	}
	// 非本人撤回拒绝
	inst2 := mustStart(t, flowTestDefID, "撤回-越权")
	if err := WithdrawFlowInstance(inst2, flowTestUser.ID); err == nil {
		t.Fatal("非本人撤回应报错")
	}
	// 管理员终止
	if err := TerminateFlowInstance(inst2, flowTestAdmin.ID, "作废"); err != nil {
		t.Fatalf("终止失败: %v", err)
	}
	if instByID(t, inst2).Status != "5" {
		t.Fatal("期望终止(5)")
	}
	// 运行中实例禁止修改流程节点
	defID, err := SaveFlowDef(FlowDefInput{
		DefName: "守卫测试", DefKey: "guard", Category: "测试",
		Nodes: []FlowNodeInput{{NodeOrder: 1, NodeName: "审批", ApproverType: "user", ApproverIDs: fmt.Sprintf("%d", flowTestAdmin.ID), NodeMode: "one"}},
	}, "admin")
	if err != nil {
		t.Fatalf("创建流程失败: %v", err)
	}
	runningID := mustStart(t, defID, "运行中实例")
	_ = runningID
	if _, err := SaveFlowDef(FlowDefInput{
		FlowDefID: defID, DefName: "守卫测试", DefKey: "guard", Category: "测试",
		Nodes: []FlowNodeInput{{NodeOrder: 1, NodeName: "改后", ApproverType: "user", ApproverIDs: fmt.Sprintf("%d", flowTestAdmin.ID), NodeMode: "one"}},
	}, "admin"); err == nil {
		t.Fatal("运行中实例存在时应禁止修改节点")
	}
	// 重复标识拒绝
	if _, err := SaveFlowDef(FlowDefInput{
		DefName: "重复标识", DefKey: "guard", Category: "测试",
		Nodes: []FlowNodeInput{{NodeOrder: 1, NodeName: "审批", ApproverType: "user", ApproverIDs: fmt.Sprintf("%d", flowTestAdmin.ID), NodeMode: "one"}},
	}, "admin"); err == nil {
		t.Fatal("重复 def_key 应报错")
	}
	// 删除：无实例的流程可删；有实例的禁止
	freeID, err := SaveFlowDef(FlowDefInput{
		DefName: "可删流程", DefKey: "free", Category: "测试",
		Nodes: []FlowNodeInput{{NodeOrder: 1, NodeName: "审批", ApproverType: "user", ApproverIDs: fmt.Sprintf("%d", flowTestAdmin.ID), NodeMode: "one"}},
	}, "admin")
	if err != nil {
		t.Fatalf("创建流程失败: %v", err)
	}
	if err := DeleteFlowDef(freeID); err != nil {
		t.Fatalf("删除无实例流程失败: %v", err)
	}
	if err := DeleteFlowDef(defID); err == nil {
		t.Fatal("删除有实例的流程应报错")
	}
}

func TestFlowOverdueCalc(t *testing.T) {
	// 超时时间：节点优先，其次流程定义
	var def model.FlowDef
	if err := DB.First(&def, flowTestDefID).Error; err != nil {
		t.Fatalf("读取流程失败: %v", err)
	}
	node := model.FlowNode{TimeoutHours: 0}
	overdue := flowNodeTimeoutAt(node, def, time.Now())
	if overdue != nil {
		t.Fatalf("无超时配置期望 nil，实际 %v", overdue)
	}
	node2 := model.FlowNode{TimeoutHours: 2}
	overdue2 := flowNodeTimeoutAt(node2, def, time.Now())
	if overdue2 == nil || !overdue2.After(time.Now().Add(1*time.Hour)) {
		t.Fatalf("节点超时计算异常: %v", overdue2)
	}
}

func TestFlowWithdrawAfterApproved(t *testing.T) {
	// 撤回语义：第一节点审批人已通过、实例仍在运行 → 申请人不可再撤回
	instID := mustStart(t, flowTestDef2, "撤回-已审批")
	t1 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := ApproveFlowTask(t1.ID, flowTestAdmin.ID, "节点1同意"); err != nil {
		t.Fatalf("节点1审批失败: %v", err)
	}
	if instByID(t, instID).Status != "1" {
		t.Fatalf("期望实例仍在运行，实际 %s", instByID(t, instID).Status)
	}
	if err := WithdrawFlowInstance(instID, flowTestAdmin.ID); err == nil {
		t.Fatal("已有审批人处理后撤回应报错")
	}
	// 未被处理时仍可撤回（回归）
	inst2 := mustStart(t, flowTestDef2, "撤回-未处理")
	if err := WithdrawFlowInstance(inst2, flowTestAdmin.ID); err != nil {
		t.Fatalf("未处理撤回应成功: %v", err)
	}
}

func TestFlowTransferAddSignGuards(t *testing.T) {
	// 转办给自己拒绝
	instID := mustStart(t, flowTestDefID, "转办-自己")
	t1 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := TransferFlowTask(t1.ID, flowTestAdmin.ID, flowTestAdmin.ID, "x"); err == nil {
		t.Fatal("转办给自己应报错")
	}
	// 加签给自己拒绝
	inst2 := mustStart(t, flowTestDefID, "加签-自己")
	t2 := todoOf(t, inst2, flowTestAdmin.ID)[0]
	if err := AddSignFlowTask(t2.ID, flowTestAdmin.ID, flowTestAdmin.ID, "x"); err == nil {
		t.Fatal("加签给自己应报错")
	}
	// 重复加签：目标人已是当前节点待办审批人 → 拒绝
	inst3 := mustStart(t, flowTestDefID, "加签-重复")
	t3 := todoOf(t, inst3, flowTestAdmin.ID)[0]
	if err := AddSignFlowTask(t3.ID, flowTestAdmin.ID, flowTestUser.ID, "加签1"); err != nil {
		t.Fatalf("加签失败: %v", err)
	}
	if err := AddSignFlowTask(t3.ID, flowTestAdmin.ID, flowTestUser.ID, "加签2"); err == nil {
		t.Fatal("重复加签同一人应报错")
	}
	// 转办到已是当前节点待办人 → 拒绝（先转办给测试审批人后，被转办人已是待办）
	inst4 := mustStart(t, flowTestDefID, "转办-重复")
	t4 := todoOf(t, inst4, flowTestAdmin.ID)[0]
	if err := TransferFlowTask(t4.ID, flowTestAdmin.ID, flowTestUser.ID, "第一次"); err != nil {
		t.Fatalf("首次转办失败: %v", err)
	}
	u2 := todoOf(t, inst4, flowTestUser.ID)
	if len(u2) != 1 {
		t.Fatalf("被转办人期望 1 条待办，实际 %d", len(u2))
	}
	if err := TransferFlowTask(u2[0].ID, flowTestUser.ID, flowTestUser.ID, "再转办"); err == nil {
		t.Fatal("转办给自己应报错")
	}
}

func TestFlowCcExcludeSelfAndListSearch(t *testing.T) {
	// 抄送排除申请人本人 + 去重
	instID, err := StartFlowInstance(flowTestAdmin, FlowStartReq{
		FlowDefID: flowTestDefID, Title: "抄送-排除自己", Content: "x",
		CCUserIDs: []uint{flowTestAdmin.ID, flowTestUser.ID, flowTestUser.ID},
	})
	if err != nil {
		t.Fatalf("发起失败: %v", err)
	}
	var cc []model.FlowTask
	if err := DB.Where("instance_id = ? AND task_type = '2'", instID).Find(&cc).Error; err != nil {
		t.Fatalf("查询抄送失败: %v", err)
	}
	if len(cc) != 1 || cc[0].AssigneeID != flowTestUser.ID {
		t.Fatalf("期望仅 1 条抄送（测试审批人），实际 %+v", cc)
	}
	// 待办列表关键词搜索：命中与不命中（「加签-重复」实例中 admin 有待办）
	if _, total, err := ListFlowTodo(flowTestAdmin.ID, 1, 10, "加签-重复"); err != nil {
		t.Fatalf("待办搜索失败: %v", err)
	} else if total == 0 {
		t.Fatal("按标题搜索待办应命中")
	}
	if _, total2, _ := ListFlowTodo(flowTestAdmin.ID, 1, 10, "不存在的标题XYZ"); total2 != 0 {
		t.Fatalf("无关关键词应命中 0，实际 %d", total2)
	}
	// 抄送列表关键词搜索
	if _, ccTotal, err := ListFlowCc(flowTestUser.ID, 1, 10, "抄送-排除自己"); err != nil {
		t.Fatalf("抄送搜索失败: %v", err)
	} else if ccTotal == 0 {
		t.Fatal("抄送按标题搜索应命中")
	}
	// 已办列表空关键词 = 全量不报错
	if _, doneTotal, err := ListFlowDone(flowTestAdmin.ID, 1, 10, ""); err != nil {
		t.Fatalf("已办搜索失败: %v", err)
	} else if doneTotal == 0 {
		t.Fatal("已办列表应存在记录")
	}
}

func TestFlowRejectBackRecord(t *testing.T) {
	// 动态两级流程（两节点均 reject_to=1）：节点2 驳回 → 退回节点1 运行、时间线写入 action=12「退回」记录
	defID, err := SaveFlowDef(FlowDefInput{
		DefName: "退回测试", DefKey: "rejectback", Category: "测试",
		Nodes: []FlowNodeInput{
			{NodeOrder: 1, NodeName: "初审", ApproverType: "user", ApproverIDs: fmt.Sprintf("%d", flowTestAdmin.ID), NodeMode: "one", RejectTo: "1"},
			{NodeOrder: 2, NodeName: "终审", ApproverType: "user", ApproverIDs: fmt.Sprintf("%d", flowTestAdmin.ID), NodeMode: "one", RejectTo: "1"},
		},
	}, "admin")
	if err != nil {
		t.Fatalf("创建退回流程失败: %v", err)
	}
	instID := mustStart(t, defID, "退回-标记")
	t1 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := ApproveFlowTask(t1.ID, flowTestAdmin.ID, "初审同意"); err != nil {
		t.Fatalf("节点1审批失败: %v", err)
	}
	t2 := todoOf(t, instID, flowTestAdmin.ID)
	if len(t2) != 1 || t2[0].NodeOrder != 2 {
		t.Fatalf("期望推进到节点2，实际 %+v", t2)
	}
	if err := RejectFlowTask(t2[0].ID, flowTestAdmin.ID, "补充说明"); err != nil {
		t.Fatalf("节点2驳回失败: %v", err)
	}
	inst := instByID(t, instID)
	if inst.Status != "1" || inst.CurrentOrder != 1 {
		t.Fatalf("期望退回节点1运行中，实际 status=%s order=%d", inst.Status, inst.CurrentOrder)
	}
	back := todoOf(t, instID, flowTestAdmin.ID)
	if len(back) != 1 || back[0].NodeOrder != 1 {
		t.Fatalf("期望节点1重新生成待办，实际 %+v", back)
	}
	var recs []model.FlowRecord
	if err := DB.Where("instance_id = ?", instID).Order("flow_record_id ASC").Find(&recs).Error; err != nil {
		t.Fatalf("查询记录失败: %v", err)
	}
	hasBack := false
	for _, r := range recs {
		if r.Action == "12" {
			hasBack = true
		}
	}
	if !hasBack {
		t.Fatal("退回上一节点应写入 action=12 记录")
	}
}

func TestFlowUrgeDedupByNode(t *testing.T) {
	// 催办判重口径统一：同实例同节点只记一次；流转到新节点后当前节点变化可再催
	instID := mustStart(t, flowTestDef2, "催办-判重")
	if err := UrgeFlowInstance(instID, flowTestAdmin.ID); err != nil {
		t.Fatalf("首次催办失败: %v", err)
	}
	if err := UrgeFlowInstance(instID, flowTestAdmin.ID); err != nil {
		t.Fatalf("同节点二次催办不应报错: %v", err)
	}
	var cnt int64
	if err := DB.Model(&model.FlowRecord{}).Where("instance_id = ? AND action = '9'", instID).Count(&cnt).Error; err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("同实例同节点期望 1 条催办记录，实际 %d", cnt)
	}
	// 节点1通过 → 节点2，当前节点变化 → 可再催
	t1 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := ApproveFlowTask(t1.ID, flowTestAdmin.ID, "通过"); err != nil {
		t.Fatalf("节点1审批失败: %v", err)
	}
	if err := UrgeFlowInstance(instID, flowTestAdmin.ID); err != nil {
		t.Fatalf("新节点催办失败: %v", err)
	}
	if err := DB.Model(&model.FlowRecord{}).Where("instance_id = ? AND action = '9'", instID).Count(&cnt).Error; err != nil {
		t.Fatalf("统计失败: %v", err)
	}
	if cnt != 2 {
		t.Fatalf("流转后新节点期望 2 条催办记录，实际 %d", cnt)
	}
}

func TestFlowInstanceDetailScope(t *testing.T) {
	// 数据权限：申请人本人可看、无关联用户不可看、抄送人可看、admin 豁免
	instID := mustStart(t, flowTestDefID, "详情-权限") // 通用审批：节点审批人=admin
	if _, err := GetFlowInstanceDetail(instID, flowTestAdmin.ID); err != nil {
		t.Fatalf("申请人应可查看详情: %v", err)
	}
	// flowTestUser 与实例无任何关联（非申请人/非任务 assignee/非 admin）→ 拒绝
	if _, err := GetFlowInstanceDetail(instID, flowTestUser.ID); err == nil || err.Error() != "无权查看该申请" {
		t.Fatalf("无关联用户应被拒绝，实际: %v", err)
	}
	// 抄送人可看：发起带抄送 flowTestUser
	instID2, err := StartFlowInstance(flowTestAdmin, FlowStartReq{
		FlowDefID: flowTestDefID, Title: "详情-抄送", Content: "c",
		CCUserIDs: []uint{flowTestUser.ID},
	})
	if err != nil {
		t.Fatalf("发起失败: %v", err)
	}
	if _, err := GetFlowInstanceDetail(instID2, flowTestUser.ID); err != nil {
		t.Fatalf("抄送人应可查看详情: %v", err)
	}
}

func TestFlowUrgeScope(t *testing.T) {
	// 权限：申请人本人可催办，无关联用户不可催办
	instID := mustStart(t, flowTestDefID, "催办-权限")
	if err := UrgeFlowInstance(instID, flowTestAdmin.ID); err != nil {
		t.Fatalf("申请人催办失败: %v", err)
	}
	if err := UrgeFlowInstance(instID, flowTestUser.ID); err == nil || err.Error() != "无权催办该申请" {
		t.Fatalf("无关联用户应被拒绝催办，实际: %v", err)
	}
}

func TestFlowWithdrawAfterTransfer(t *testing.T) {
	// 撤回语义：审批人已转办（action=4）后申请人不可再撤回
	instID := mustStart(t, flowTestDefID, "撤回-转办")
	t1 := todoOf(t, instID, flowTestAdmin.ID)[0]
	if err := TransferFlowTask(t1.ID, flowTestAdmin.ID, flowTestUser.ID, "转给你处理"); err != nil {
		t.Fatalf("转办失败: %v", err)
	}
	if err := WithdrawFlowInstance(instID, flowTestAdmin.ID); err == nil || err.Error() != "已有审批人处理，无法撤回" {
		t.Fatalf("转办后撤回应被拒绝，实际: %v", err)
	}
}
