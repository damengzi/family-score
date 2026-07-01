package consts

const (
	// RoleAdmin 管理员角色。
	RoleAdmin = "ADMIN"
	// RoleParent 家长角色。
	RoleParent = "PARENT"
	// RoleChild 孩子角色。
	RoleChild = "CHILD"
)

const (
	// RecordTypeAdd 加分记录。
	RecordTypeAdd = "ADD"
	// RecordTypeDeduct 扣分记录。
	RecordTypeDeduct = "DEDUCT"
	// RecordTypeRepair 修复记录。
	RecordTypeRepair = "REPAIR"
	// RecordTypeTeam 家庭小队分记录。
	RecordTypeTeam = "TEAM"
	// RecordTypeStar 星星记录。
	RecordTypeStar = "STAR"
	// RecordTypeExchange 兑换记录。
	RecordTypeExchange = "EXCHANGE"
)

const (
	// AccountAuto 自动账户。
	AccountAuto = "AUTO"
	// AccountBase 基准分账户。
	AccountBase = "BASE"
	// AccountBonus 兑换分账户。
	AccountBonus = "BONUS"
	// AccountTeam 家庭小队分账户。
	AccountTeam = "TEAM"
	// AccountStar 星星账户。
	AccountStar = "STAR"
)

const (
	// TaskTypeDaily 每日任务。
	TaskTypeDaily = "DAILY"
	// TaskTypeRepair 修复任务。
	TaskTypeRepair = "REPAIR"
	// TaskTypeTeam 家庭小队任务。
	TaskTypeTeam = "TEAM"
)

const (
	// TaskStatusTodo 待完成。
	TaskStatusTodo = "TODO"
	// TaskStatusSubmitted 待审核。
	TaskStatusSubmitted = "SUBMITTED"
	// TaskStatusApproved 已通过。
	TaskStatusApproved = "APPROVED"
	// TaskStatusRejected 已驳回。
	TaskStatusRejected = "REJECTED"
)

const (
	// DailyTaskScoreLimit 每个孩子每日任务审核加分上限。
	DailyTaskScoreLimit = 15
	// StarLimit 星星数量上限。
	StarLimit = 20
)

const (
	// DefaultAdminLoginName 固定管理员用户名。
	DefaultAdminLoginName = "admin"
	// DefaultAdminPassword 固定管理员密码。
	DefaultAdminPassword = "654321"
)
