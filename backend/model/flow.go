package model

import "time"

// 工作流/审批状态机（自研轻量引擎）
// 状态约定：
//   - sys_flow_instance.status：1 运行中 / 2 已通过 / 3 已驳回 / 4 已撤销 / 5 已终止
//   - sys_flow_task.status：0 待办 / 1 已通过 / 2 已驳回 / 3 已转出 / 4 已加签 / 5 已撤销 / 6 已读（抄送）

// FlowDef 流程定义表
type FlowDef struct {
	ID           uint       `gorm:"column:flow_def_id;primaryKey" json:"flowDefId"`
	DefName      string     `gorm:"column:def_name;size:64" json:"defName"`
	DefKey       string     `gorm:"column:def_key;size:64;uniqueIndex" json:"defKey"`
	Category     string     `gorm:"column:category;size:32" json:"category"`            // 流程分类（如 审批/请假）
	Status       string     `gorm:"column:status;size:8;default:0" json:"status"`       // 0 启用 1 停用
	TimeoutHours int        `gorm:"column:timeout_hours;default:0" json:"timeoutHours"` // 流程级超时（小时，0=不限，节点未单独配置时继承）
	Remark       string     `gorm:"column:remark;size:255" json:"remark"`
	CreateBy     string     `gorm:"column:create_by;size:64" json:"createBy"`
	CreatedAt    time.Time  `json:"createTime"`
	UpdatedAt    time.Time  `json:"updateTime"`
	Nodes        []FlowNode `gorm:"foreignKey:FlowDefID" json:"nodes"` // 设计时节点列表
}

// TableName 指定表名
func (FlowDef) TableName() string { return "sys_flow_def" }

// FlowNode 流程节点表（有序步骤，node_order 从 1 递增）
type FlowNode struct {
	ID           uint      `gorm:"column:flow_node_id;primaryKey" json:"flowNodeId"`
	FlowDefID    uint      `gorm:"column:flow_def_id;index" json:"flowDefId"`
	NodeOrder    int       `gorm:"column:node_order;default:1" json:"nodeOrder"` // 步骤序号（1 起）
	NodeName     string    `gorm:"column:node_name;size:64" json:"nodeName"`
	ApproverType string    `gorm:"column:approver_type;size:16" json:"approverType"`       // role 按角色 / user 指定人 / dept_leader 发起人部门负责人
	ApproverIDs  string    `gorm:"column:approver_ids;size:1024" json:"approverIds"`       // 逗号分隔：role=角色ID，user=用户ID，dept_leader=部门ID（单值）
	NodeMode     string    `gorm:"column:node_mode;size:8;default:one" json:"nodeMode"`    // one 或签（任一通过） / any 会签（全部通过）
	CanAddSign   string    `gorm:"column:can_add_sign;size:1;default:0" json:"canAddSign"` // 0 否 1 是（审批中可加签）
	TimeoutHours int       `gorm:"column:timeout_hours;default:0" json:"timeoutHours"`     // 节点级超时（小时，0=继承流程定义）
	RejectTo     string    `gorm:"column:reject_to;size:8;default:0" json:"rejectTo"`      // 0 驳回回发起人 / 1 退回上一节点
	Remark       string    `gorm:"column:remark;size:255" json:"remark"`
	CreatedAt    time.Time `json:"createTime"`
	UpdatedAt    time.Time `json:"updateTime"`
}

// TableName 指定表名
func (FlowNode) TableName() string { return "sys_flow_node" }

// FlowInstance 流程实例表
type FlowInstance struct {
	ID               uint       `gorm:"column:flow_instance_id;primaryKey" json:"flowInstanceId"`
	FlowDefID        uint       `gorm:"column:flow_def_id;index" json:"flowDefId"`
	DefName          string     `gorm:"column:def_name;size:64" json:"defName"` // 冗余流程名（快照）
	Title            string     `gorm:"column:title;size:128" json:"title"`
	Content          string     `gorm:"column:content;type:text" json:"content"`
	ApplicantID      uint       `gorm:"column:applicant_id;index" json:"applicantId"`
	ApplicantName    string     `gorm:"column:applicant_name;size:64" json:"applicantName"`
	ApplicantDeptID  uint       `gorm:"column:applicant_dept_id;default:0" json:"applicantDeptId"` // 发起时部门快照（dept_leader 路由用）
	CCUserIDs        string     `gorm:"column:cc_user_ids;size:1024" json:"ccUserIds"`             // 发起时抄送人（逗号分隔用户 ID）
	Status           string     `gorm:"column:status;size:8;default:1;index" json:"status"`        // 1 运行中 2 已通过 3 已驳回 4 已撤销 5 已终止
	CurrentNodeID    uint       `gorm:"column:current_node_id;default:0" json:"currentNodeId"`
	CurrentOrder     int        `gorm:"column:current_order;default:0" json:"currentOrder"`           // 当前步骤序号
	RejectFromNodeID uint       `gorm:"column:reject_from_node_id;default:0" json:"rejectFromNodeId"` // 驳回来源节点（重新提交时从此恢复）
	CreatedAt        time.Time  `json:"createTime"`
	FinishedAt       *time.Time `json:"finishedAt"`
	UpdatedAt        time.Time  `json:"updateTime"`
	Nodes            []FlowNode `gorm:"-" json:"-"` // 运行时节点快照（详情组装，不落库）
}

// TableName 指定表名
func (FlowInstance) TableName() string { return "sys_flow_instance" }

// FlowTask 流程任务表（一个待办=一条任务；抄送/加签同表区分类型）
type FlowTask struct {
	ID           uint       `gorm:"column:flow_task_id;primaryKey" json:"flowTaskId"`
	InstanceID   uint       `gorm:"column:instance_id;index" json:"instanceId"`
	NodeID       uint       `gorm:"column:node_id;default:0" json:"nodeId"`
	NodeName     string     `gorm:"column:node_name;size:64" json:"nodeName"`
	NodeOrder    int        `gorm:"column:node_order;default:0" json:"nodeOrder"`
	TaskType     string     `gorm:"column:task_type;size:8;default:1" json:"taskType"` // 1 审批 2 抄送 3 加签
	AssigneeID   uint       `gorm:"column:assignee_id;index" json:"assigneeId"`
	AssigneeName string     `gorm:"column:assignee_name;size:64" json:"assigneeName"`
	Status       string     `gorm:"column:status;size:8;default:0;index" json:"status"` // 0 待办 1 已通过 2 已驳回 3 已转出 4 已加签 5 已撤销 6 已读
	Comment      string     `gorm:"column:comment;size:512" json:"comment"`
	OverdueAt    *time.Time `gorm:"column:overdue_at" json:"overdueAt"` // 超时时间（节点/流程配置小时换算，可为空）
	PassedAt     *time.Time `gorm:"column:passed_at" json:"passedAt"`
	CreatedAt    time.Time  `json:"createTime"`
	UpdatedAt    time.Time  `json:"updateTime"`
}

// TableName 指定表名
func (FlowTask) TableName() string { return "sys_flow_task" }

// FlowRecord 审批记录表（时间线）
type FlowRecord struct {
	ID           uint      `gorm:"column:flow_record_id;primaryKey" json:"flowRecordId"`
	InstanceID   uint      `gorm:"column:instance_id;index" json:"instanceId"`
	NodeID       uint      `gorm:"column:node_id;default:0" json:"nodeId"`
	NodeName     string    `gorm:"column:node_name;size:64" json:"nodeName"`
	OperatorID   uint      `gorm:"column:operator_id" json:"operatorId"`
	OperatorName string    `gorm:"column:operator_name;size:64" json:"operatorName"`
	Action       string    `gorm:"column:action;size:8" json:"action"`           // 1 发起 2 通过 3 驳回 4 转办 5 加签 6 撤回 7 撤销 8 终止 9 催办 10 重新提交 11 抄送
	TargetName   string    `gorm:"column:target_name;size:64" json:"targetName"` // 操作对象（被转办人/被加签人/抄送人，可为空）
	Comment      string    `gorm:"column:comment;size:512" json:"comment"`
	CreatedAt    time.Time `json:"createTime"`
}

// TableName 指定表名
func (FlowRecord) TableName() string { return "sys_flow_record" }
