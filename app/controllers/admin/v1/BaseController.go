package v1

type BaseController struct{}

type Group struct {
	BaseController
	AuthController
	DashboardController
	ManualReviewController
	ScanLogController
	BookController
	PriceSnapshotController
	ImportController
	RecycleRuleController
	SystemController
}
