package protocol

import "time"

// Session 表示本机登录会话。
type Session struct {
	UserID    int64     `json:"userId"`
	FamilyID  int64     `json:"familyId"`
	ChildID   int64     `json:"childId"`
	Role      string    `json:"role"`
	Name      string    `json:"name"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// User 表示本机用户账号。
type User struct {
	ID          int64  `json:"id"`
	FamilyID    int64  `json:"familyId"`
	ChildID     int64  `json:"childId"`
	Role        string `json:"role"`
	LoginName   string `json:"loginName"`
	Name        string `json:"name"`
	ParentTitle string `json:"parentTitle"`
	ParentGroup string `json:"parentGroup"`
	Enabled     bool   `json:"enabled"`
	CreatedAt   string `json:"createdAt"`
}

// Child 表示孩子档案。
type Child struct {
	ID           int64  `json:"id"`
	FamilyID     int64  `json:"familyId"`
	ParentUserID int64  `json:"parentUserId"`
	ParentGroup  string `json:"parentGroup"`
	Name         string `json:"name"`
	Age          int    `json:"age"`
	Gender       string `json:"gender"`
}

// Account 表示孩子当前月份的积分账户。
type Account struct {
	ID                  int64  `json:"id"`
	ChildID             int64  `json:"childId"`
	BaseScore           int    `json:"baseScore"`
	BonusScore          int    `json:"bonusScore"`
	StarCount           int    `json:"starCount"`
	TeamScore           int    `json:"teamScore"`
	StatusLevel         string `json:"statusLevel"`
	CurrentMonth        string `json:"currentMonth"`
	LastExchangeDate    string `json:"lastExchangeDate"`
	AppealCountThisWeek int    `json:"appealCountThisWeek"`
	Version             int    `json:"version"`
}

// ScoreRecord 表示一条积分账户变更流水。
type ScoreRecord struct {
	ID            int64  `json:"id"`
	ChildID       int64  `json:"childId"`
	RecordType    string `json:"recordType"`
	TargetAccount string `json:"targetAccount"`
	ItemName      string `json:"itemName"`
	ScoreChange   int    `json:"scoreChange"`
	BeforeValue   int    `json:"beforeValue"`
	AfterValue    int    `json:"afterValue"`
	OperatorRole  string `json:"operatorRole"`
	Reason        string `json:"reason"`
	Evidence      string `json:"evidence"`
	ConfirmStatus string `json:"confirmStatus"`
	OccurredAt    string `json:"occurredAt"`
}

// TaskQuestion 表示任务中的一道学科题目。
type TaskQuestion struct {
	Subject      string `json:"subject"`
	QuestionType string `json:"questionType"`
	Content      string `json:"content"`
	Answer       string `json:"answer"`
}

// TaskInstance 表示某一天生成给孩子的任务实例。
type TaskInstance struct {
	ID            int64          `json:"id"`
	ChildID       int64          `json:"childId"`
	TemplateID    int64          `json:"templateId"`
	TaskName      string         `json:"taskName"`
	TaskType      string         `json:"taskType"`
	Category      string         `json:"category"`
	Subject       string         `json:"subject"`
	Content       string         `json:"content"`
	QuestionType  string         `json:"questionType"`
	Answer        string         `json:"answer"`
	Questions     []TaskQuestion `json:"questions"`
	ScoreValue    int            `json:"scoreValue"`
	TargetAccount string         `json:"targetAccount"`
	Status        string         `json:"status"`
	SubmitNote    string         `json:"submitNote"`
	AuditNote     string         `json:"auditNote"`
	TaskDate      string         `json:"taskDate"`
	DueAt         string         `json:"dueAt"`
}

// TaskTemplate 表示可配置任务模板。
type TaskTemplate struct {
	ID                int64          `json:"id"`
	TaskName          string         `json:"taskName"`
	TaskType          string         `json:"taskType"`
	Category          string         `json:"category"`
	Subject           string         `json:"subject"`
	Content           string         `json:"content"`
	QuestionType      string         `json:"questionType"`
	Answer            string         `json:"answer"`
	Questions         []TaskQuestion `json:"questions"`
	ScoreValue        int            `json:"scoreValue"`
	TargetAccount     string         `json:"targetAccount"`
	NeedParentConfirm bool           `json:"needParentConfirm"`
	DailyLimit        int            `json:"dailyLimit"`
	WeeklyLimit       int            `json:"weeklyLimit"`
	Enabled           bool           `json:"enabled"`
	Description       string         `json:"description"`
	DueTime           string         `json:"dueTime"`
}

// Reward 表示可兑换奖励配置。
type Reward struct {
	ID                int64  `json:"id"`
	RewardName        string `json:"rewardName"`
	RewardType        string `json:"rewardType"`
	CostScore         int    `json:"costScore"`
	CostStar          int    `json:"costStar"`
	WeeklyLimit       int    `json:"weeklyLimit"`
	MonthlyLimit      int    `json:"monthlyLimit"`
	HealthRisk        string `json:"healthRisk"`
	NeedParentConfirm bool   `json:"needParentConfirm"`
	Enabled           bool   `json:"enabled"`
	Description       string `json:"description"`
}

// ExchangeOrder 表示奖励兑换申请。
type ExchangeOrder struct {
	ID         int64  `json:"id"`
	ChildID    int64  `json:"childId"`
	RewardID   int64  `json:"rewardId"`
	RewardName string `json:"rewardName"`
	CostScore  int    `json:"costScore"`
	CostStar   int    `json:"costStar"`
	Status     string `json:"status"`
	ApplyNote  string `json:"applyNote"`
	AuditNote  string `json:"auditNote"`
	AppliedAt  string `json:"appliedAt"`
}

// Dashboard 表示孩子首页聚合数据。
type Dashboard struct {
	Account Account        `json:"account"`
	Records []ScoreRecord  `json:"records"`
	Tasks   []TaskInstance `json:"tasks"`
}

// Wish 表示孩子提交的奖励愿望。
type Wish struct {
	ID            int64  `json:"id"`
	ChildID       int64  `json:"childId"`
	WishName      string `json:"wishName"`
	WishType      string `json:"wishType"`
	ExpectedScore int    `json:"expectedScore"`
	ExpectedStar  int    `json:"expectedStar"`
	Reason        string `json:"reason"`
	Status        string `json:"status"`
	AuditNote     string `json:"auditNote"`
	CreatedAt     string `json:"createdAt"`
	AuditedAt     string `json:"auditedAt"`
}

// Appeal 表示孩子对任务或积分记录的申诉。
type Appeal struct {
	ID               int64  `json:"id"`
	ChildID          int64  `json:"childId"`
	TargetType       string `json:"targetType"`
	TargetID         int64  `json:"targetId"`
	RecordID         int64  `json:"recordId"`
	AppealReason     string `json:"appealReason"`
	ExpectedSolution string `json:"expectedSolution"`
	Status           string `json:"status"`
	HandleResult     string `json:"handleResult"`
	HandleNote       string `json:"handleNote"`
	CreatedAt        string `json:"createdAt"`
	HandledAt        string `json:"handledAt"`
}

// GuardianGroup 表示监护组。
type GuardianGroup struct {
	ID          int64  `json:"id"`
	FamilyID    int64  `json:"familyId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentCount int    `json:"parentCount"`
	ChildCount  int    `json:"childCount"`
	CreatedAt   string `json:"createdAt"`
}

// BackupRecord 表示本地备份记录。
type BackupRecord struct {
	ID            int64  `json:"id"`
	OperationType string `json:"operationType"`
	FilePath      string `json:"filePath"`
	FileSize      int64  `json:"fileSize"`
	Status        string `json:"status"`
	Remark        string `json:"remark"`
	CreatedAt     string `json:"createdAt"`
}

// SystemStatusParam 表示系统状态。
type SystemStatusParam struct {
	SetupCompleted bool   `json:"setupCompleted"`
	DataDir        string `json:"dataDir"`
	DBPath         string `json:"dbPath"`
	Addr           string `json:"addr"`
	Now            string `json:"now"`
}

// Profile 表示当前登录用户的个人主页基础信息。
type Profile struct {
	User             User   `json:"user"`
	BoundChildName   string `json:"boundChildName"`
	AccessibleChilds int    `json:"accessibleChilds"`
	SessionExpiresAt string `json:"sessionExpiresAt"`
}

// SetupInitParam 表示首次初始化参数。
type SetupInitParam struct {
	FamilyName     string             `json:"familyName"`
	AdminPassword  string             `json:"adminPassword"`
	GroupName      string             `json:"groupName"`
	ImportDefaults bool               `json:"importDefaults"`
	Parents        []SetupParentParam `json:"parents"`
	Children       []SetupChildParam  `json:"children"`
}

// SetupParentParam 表示初始化家长账号参数。
type SetupParentParam struct {
	DisplayName string `json:"displayName"`
	LoginName   string `json:"loginName"`
	Password    string `json:"password"`
	ParentTitle string `json:"parentTitle"`
}

// SetupChildParam 表示初始化孩子档案和账号参数。
type SetupChildParam struct {
	Name           string `json:"name"`
	Age            int    `json:"age"`
	Gender         string `json:"gender"`
	ChildLoginName string `json:"childLoginName"`
	ChildPassword  string `json:"childPassword"`
}

// CreateGuardianGroupParam 表示新增监护组参数。
type CreateGuardianGroupParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateGuardianGroupParam 表示修改监护组参数。
type UpdateGuardianGroupParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateUserParam 表示管理员新增用户参数。
type CreateUserParam struct {
	DisplayName string `json:"displayName"`
	LoginName   string `json:"loginName"`
	Password    string `json:"password"`
	Role        string `json:"role"`
	ChildID     int64  `json:"childId"`
	ParentTitle string `json:"parentTitle"`
	ParentGroup string `json:"parentGroup"`
}

// UpdateUserParam 表示管理员修改用户参数。
type UpdateUserParam struct {
	DisplayName string `json:"displayName"`
	Password    string `json:"password"`
	ChildID     int64  `json:"childId"`
	ParentTitle string `json:"parentTitle"`
	ParentGroup string `json:"parentGroup"`
}

// CreateChildParam 表示新增孩子档案参数。
type CreateChildParam struct {
	Name           string `json:"name"`
	Age            int    `json:"age"`
	Gender         string `json:"gender"`
	ParentUserID   int64  `json:"parentUserId"`
	ParentGroup    string `json:"parentGroup"`
	ChildLoginName string `json:"childLoginName"`
	ChildPassword  string `json:"childPassword"`
}

// UpdateChildParam 表示修改孩子档案参数。
type UpdateChildParam struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	Gender       string `json:"gender"`
	ParentUserID int64  `json:"parentUserId"`
	ParentGroup  string `json:"parentGroup"`
}

// PasswordCaptchaChoice 表示重置密码图片验证码选项。
type PasswordCaptchaChoice struct {
	Key   string `json:"key"`
	Image string `json:"image"`
}

// PasswordCaptcha 表示重置密码图片验证码。
type PasswordCaptcha struct {
	Token   string                  `json:"token"`
	Prompt  string                  `json:"prompt"`
	Choices []PasswordCaptchaChoice `json:"choices"`
}

// ResetPasswordParam 表示忘记密码重置参数。
type ResetPasswordParam struct {
	LoginName     string `json:"loginName"`
	Password      string `json:"password"`
	CaptchaToken  string `json:"captchaToken"`
	CaptchaAnswer string `json:"captchaAnswer"`
}

// ChangePasswordParam 表示登录用户修改自己密码的参数。
type ChangePasswordParam struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// SelfRegisterParam 表示自主注册普通用户参数。
type SelfRegisterParam struct {
	DisplayName string `json:"displayName"`
	LoginName   string `json:"loginName"`
	Password    string `json:"password"`
}

// CreateScoreRecordParam 表示手动记分参数。
type CreateScoreRecordParam struct {
	ChildID       int64  `json:"childId"`
	RecordType    string `json:"recordType"`
	TargetAccount string `json:"targetAccount"`
	ItemName      string `json:"itemName"`
	ScoreChange   int    `json:"scoreChange"`
	Reason        string `json:"reason"`
	Evidence      string `json:"evidence"`
}

// SubmitTaskParam 表示任务提交参数。
type SubmitTaskParam struct {
	SubmitNote string `json:"submitNote"`
}

// AuditParam 表示审核参数。
type AuditParam struct {
	Result    string `json:"result"`
	AuditNote string `json:"auditNote"`
}

// CreateTaskTemplateParam 表示新增任务模板参数。
type CreateTaskTemplateParam struct {
	TaskName      string         `json:"taskName"`
	TaskType      string         `json:"taskType"`
	Category      string         `json:"category"`
	Subject       string         `json:"subject"`
	Content       string         `json:"content"`
	QuestionType  string         `json:"questionType"`
	Answer        string         `json:"answer"`
	Questions     []TaskQuestion `json:"questions"`
	ScoreValue    int            `json:"scoreValue"`
	TargetAccount string         `json:"targetAccount"`
	Description   string         `json:"description"`
	DueTime       string         `json:"dueTime"`
}

// PublishTaskParam 表示发布一次性任务参数。
type PublishTaskParam struct {
	ChildID       int64          `json:"childId"`
	TaskName      string         `json:"taskName"`
	TaskType      string         `json:"taskType"`
	Category      string         `json:"category"`
	Subject       string         `json:"subject"`
	Content       string         `json:"content"`
	QuestionType  string         `json:"questionType"`
	Answer        string         `json:"answer"`
	Questions     []TaskQuestion `json:"questions"`
	ScoreValue    int            `json:"scoreValue"`
	TargetAccount string         `json:"targetAccount"`
	TaskDate      string         `json:"taskDate"`
	DueAt         string         `json:"dueAt"`
	DueTime       string         `json:"dueTime"`
}

// CreateRewardParam 表示新增奖励参数。
type CreateRewardParam struct {
	RewardName   string `json:"rewardName"`
	RewardType   string `json:"rewardType"`
	CostScore    int    `json:"costScore"`
	CostStar     int    `json:"costStar"`
	WeeklyLimit  int    `json:"weeklyLimit"`
	MonthlyLimit int    `json:"monthlyLimit"`
	HealthRisk   string `json:"healthRisk"`
	Description  string `json:"description"`
}

// CreateExchangeOrderParam 表示兑换申请参数。
type CreateExchangeOrderParam struct {
	ChildID  int64  `json:"childId"`
	RewardID int64  `json:"rewardId"`
	Note     string `json:"note"`
}

// CreateWishParam 表示新增愿望参数。
type CreateWishParam struct {
	ChildID       int64  `json:"childId"`
	WishName      string `json:"wishName"`
	WishType      string `json:"wishType"`
	ExpectedScore int    `json:"expectedScore"`
	ExpectedStar  int    `json:"expectedStar"`
	Reason        string `json:"reason"`
}

// AuditWishParam 表示愿望审批参数。
type AuditWishParam struct {
	Result    string `json:"result"`
	AuditNote string `json:"auditNote"`
}

// CreateAppealParam 表示新增申诉参数。
type CreateAppealParam struct {
	ChildID          int64  `json:"childId"`
	TargetType       string `json:"targetType"`
	TargetID         int64  `json:"targetId"`
	RecordID         int64  `json:"recordId"`
	AppealReason     string `json:"appealReason"`
	ExpectedSolution string `json:"expectedSolution"`
}

// HandleAppealParam 表示处理申诉参数。
type HandleAppealParam struct {
	Result     string `json:"result"`
	HandleNote string `json:"handleNote"`
}

// ApplyScoreChangeParam 表示一次账户分值变更所需参数。
type ApplyScoreChangeParam struct {
	// ChildID 孩子 ID。
	ChildID int64
	// RecordType 明细类型：ADD / DEDUCT / REPAIR / EXCHANGE / STAR / TEAM。
	RecordType string
	// TargetAccount 目标账户：AUTO / BASE / BONUS / STAR / TEAM。
	TargetAccount string
	// ItemName 项目名称。
	ItemName string
	// ScoreChange 分值变化，外部可传正数或负数，内部按类型归一化。
	ScoreChange int
	// Reason 变更原因。
	Reason string
	// Evidence 本地证据路径或说明。
	Evidence string
	// Operator 操作人。
	Operator Session
}
