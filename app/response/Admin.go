package response

type AdminLoginResponse struct {
	Token       string        `json:"token"`
	User        AdminUserDTO  `json:"user"`
	Permissions []string      `json:"permissions"`
	Menus       []AdminMenuDTO `json:"menus"`
}

type AdminUserDTO struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Name     string   `json:"name"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

type AdminMenuDTO struct {
	ID             string         `json:"id"`
	ParentID       string         `json:"parent_id"`
	Title          string         `json:"title"`
	Path           string         `json:"path"`
	Icon           string         `json:"icon"`
	PermissionCode string         `json:"permission_code"`
	Children       []AdminMenuDTO `json:"children,omitempty"`
}

type AdminPageResponse struct {
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int64       `json:"total"`
	Items    interface{} `json:"items"`
}

type DashboardSummary struct {
	TodayScanCount        int64   `json:"today_scan_count"`
	PendingReviewCount    int64   `json:"pending_review_count"`
	AcceptedTodayCount    int64   `json:"accepted_today_count"`
	SuggestedTotalPrice   float64 `json:"suggested_total_price"`
	LowConfidenceCount    int64   `json:"low_confidence_count"`
	RecentScanItems       []any   `json:"recent_scan_items"`
	PendingReviewItems    []any   `json:"pending_review_items"`
	RecentImportTaskItems []any   `json:"recent_import_task_items"`
}
