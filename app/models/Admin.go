package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type AdminUser struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Username     string     `gorm:"column:username" json:"username"`
	PasswordHash string     `gorm:"column:password_hash" json:"-"`
	Name         string     `gorm:"column:name" json:"name"`
	Status       string     `gorm:"column:status" json:"status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"lastLoginAt"`
	LastLoginIP  string     `gorm:"column:last_login_ip" json:"lastLoginIp"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (m *AdminUser) TableName() string {
	return tableName("admin_user")
}

type AdminRole struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Code        string    `gorm:"column:code" json:"code"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Status      string    `gorm:"column:status" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (m *AdminRole) TableName() string {
	return tableName("admin_role")
}

type AdminUserRole struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    uint64    `gorm:"column:user_id" json:"userId"`
	RoleID    uint64    `gorm:"column:role_id" json:"roleId"`
	CreatedAt time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *AdminUserRole) TableName() string {
	return tableName("admin_user_role")
}

type AdminMenu struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ParentID       uint64    `gorm:"column:parent_id" json:"parentId"`
	Title          string    `gorm:"column:title" json:"title"`
	Path           string    `gorm:"column:path" json:"path"`
	Icon           string    `gorm:"column:icon" json:"icon"`
	PermissionCode string    `gorm:"column:permission_code" json:"permissionCode"`
	Sort           int       `gorm:"column:sort" json:"sort"`
	Status         string    `gorm:"column:status" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (m *AdminMenu) TableName() string {
	return tableName("admin_menu")
}

type AdminAPIPermission struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Method         string    `gorm:"column:method" json:"method"`
	Path           string    `gorm:"column:path" json:"path"`
	PermissionCode string    `gorm:"column:permission_code" json:"permissionCode"`
	Description    string    `gorm:"column:description" json:"description"`
	Status         string    `gorm:"column:status" json:"status"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (m *AdminAPIPermission) TableName() string {
	return tableName("admin_api_permission")
}

type AdminRolePermission struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	RoleID         uint64    `gorm:"column:role_id" json:"roleId"`
	PermissionType string    `gorm:"column:permission_type" json:"permissionType"`
	PermissionCode string    `gorm:"column:permission_code" json:"permissionCode"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *AdminRolePermission) TableName() string {
	return tableName("admin_role_permission")
}

type AdminOperationLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID       uint64    `gorm:"column:user_id" json:"userId"`
	Username     string    `gorm:"column:username" json:"username"`
	Action       string    `gorm:"column:action" json:"action"`
	Resource     string    `gorm:"column:resource" json:"resource"`
	RequestBrief string    `gorm:"column:request_brief" json:"requestBrief"`
	Result       string    `gorm:"column:result" json:"result"`
	IP           string    `gorm:"column:ip" json:"ip"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *AdminOperationLog) TableName() string {
	return tableName("admin_operation_log")
}

type ManualDecision struct {
	ID                 uint64           `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ScanLogID          uint64           `gorm:"column:scan_log_id" json:"scanLogId"`
	Isbn               string           `gorm:"column:isbn" json:"isbn"`
	AdminUserID        uint64           `gorm:"column:admin_user_id" json:"adminUserId"`
	ManualDecision     string           `gorm:"column:manual_decision" json:"manualDecision"`
	ActualRecyclePrice *decimal.Decimal `gorm:"column:actual_recycle_price;type:decimal(20,2)" json:"actualRecyclePrice"`
	Note               string           `gorm:"column:note" json:"note"`
	CreatedAt          time.Time        `gorm:"column:created_at" json:"createdAt"`
}

func (m *ManualDecision) TableName() string {
	return tableName("manual_decision")
}

type RecycleRule struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Version           string          `gorm:"column:version" json:"version"`
	Name              string          `gorm:"column:name" json:"name"`
	MinAcceptAvgPrice decimal.Decimal `gorm:"column:min_accept_avg_price;type:decimal(20,2)" json:"minAcceptAvgPrice"`
	RecycleRate       decimal.Decimal `gorm:"column:recycle_rate;type:decimal(10,4)" json:"recycleRate"`
	MinSampleCount    int             `gorm:"column:min_sample_count" json:"minSampleCount"`
	PriceValidDays    int             `gorm:"column:price_valid_days" json:"priceValidDays"`
	LowConfidenceMode string          `gorm:"column:low_confidence_mode" json:"lowConfidenceMode"`
	Enabled           bool            `gorm:"column:enabled" json:"enabled"`
	CreatedBy         uint64          `gorm:"column:created_by" json:"createdBy"`
	CreatedAt         time.Time       `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt         time.Time       `gorm:"column:updated_at" json:"updatedAt"`
}

func (m *RecycleRule) TableName() string {
	return tableName("recycle_rule")
}

type ImportTask struct {
	ID           uint64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	FileName     string     `gorm:"column:file_name" json:"fileName"`
	FileType     string     `gorm:"column:file_type" json:"fileType"`
	Status       string     `gorm:"column:status" json:"status"`
	TotalRows    int        `gorm:"column:total_rows" json:"totalRows"`
	SuccessRows  int        `gorm:"column:success_rows" json:"successRows"`
	FailedRows   int        `gorm:"column:failed_rows" json:"failedRows"`
	CreatedBy    uint64     `gorm:"column:created_by" json:"createdBy"`
	ErrorMessage string     `gorm:"column:error_message" json:"errorMessage"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"createdAt"`
	CompletedAt  *time.Time `gorm:"column:completed_at" json:"completedAt"`
}

func (m *ImportTask) TableName() string {
	return tableName("import_task")
}

type ImportTaskRow struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	TaskID       uint64    `gorm:"column:task_id" json:"taskId"`
	RowNumber    int       `gorm:"column:row_no" json:"rowNumber"`
	Isbn         string    `gorm:"column:isbn" json:"isbn"`
	Title        string    `gorm:"column:title" json:"title"`
	Valid        bool      `gorm:"column:valid" json:"valid"`
	ErrorMessage string    `gorm:"column:error_message" json:"errorMessage"`
	RawPayload   string    `gorm:"column:raw_payload" json:"rawPayload"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"createdAt"`
}

func (m *ImportTaskRow) TableName() string {
	return tableName("import_task_row")
}
