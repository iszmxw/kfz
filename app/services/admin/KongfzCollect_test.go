package admin

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"goapi/app/models"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestMapKongfzCollectItemMapsBookShowInfoAndPrice(t *testing.T) {
	item, err := mapKongfzCollectItem(43, 1, 0, []byte(`{
		"id": 123,
		"mid": 456,
		"isbn": "978-7-5063-6543-7",
		"bookName": "活着",
		"imgUrlEntity": {"bigImgUrl": "https://example.com/cover.jpg"},
		"oldBookMinPrice": 0.01,
		"oldBookOnSaleNum": 4708,
		"bookShowInfo": ["余华  著", "作家出版社", "2012-08", "平装", "20.00"]
	}`))
	if err != nil {
		t.Fatalf("map item: %v", err)
	}
	if !item.Valid || item.NormalizedIsbn != "9787506365437" || item.Title != "活着" {
		t.Fatalf("mapped basic fields incorrectly: %+v", item)
	}
	if item.Input.Author != "余华  著" || item.Input.Publisher != "作家出版社" || item.Input.PublishYear != "2012" {
		t.Fatalf("mapped show info incorrectly: %+v", item.Input)
	}
	if item.Input.CoverURL != "https://example.com/cover.jpg" {
		t.Fatalf("mapped cover incorrectly: %q", item.Input.CoverURL)
	}
	if item.Input.OldBookMinPrice == nil || item.Input.OldBookMinPrice.String() != "0.01" || item.Input.OldBookOnSaleNum != 4708 {
		t.Fatalf("mapped price incorrectly: %+v", item.Input)
	}
}

func TestKongfzCollectorRunTaskImportsTwoPages(t *testing.T) {
	db := setupAdminServiceTestDB(t)
	task, err := CreateKongfzCollectTask(KongfzCollectCreateInput{CatID: 43, StartPage: 1, DelayMS: 0, RetryTimes: 0, CreatedBy: 1})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	claimed, ok, err := claimNextKongfzCollectTask()
	if err != nil || !ok || claimed.ID != task.ID {
		t.Fatalf("claim task = %+v ok=%v err=%v", claimed, ok, err)
	}

	collector := NewKongfzCollector(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		page := req.URL.Query().Get("page")
		body := `{"status":1,"data":{"itemResponse":{"pager":{"page":1,"pages":2,"total":2,"size":1},"list":[{"isbn":"9787506365437","bookName":"活着","imgUrlEntity":{"bigImgUrl":"https://example.com/cover1.jpg"},"oldBookMinPrice":0.01,"oldBookOnSaleNum":4708,"bookShowInfo":["余华  著","作家出版社","2012-08","平装","20.00"]}]}}}`
		if page == "2" {
			body = `{"status":1,"data":{"itemResponse":{"pager":{"page":2,"pages":2,"total":2,"size":1},"list":[{"isbn":"9787020002207","bookName":"红楼梦","imgUrlEntity":{"bigImgUrl":"https://example.com/cover2.jpg"},"oldBookMinPrice":0.20,"oldBookOnSaleNum":4823,"bookShowInfo":["曹雪芹","人民文学出版社","2008-07","平装","59.70"]}]}}}`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{}}, nil
	})})
	if err := collector.RunTask(context.Background(), task.ID); err != nil {
		t.Fatalf("run task: %v", err)
	}

	var stored models.KongfzCollectTask
	if err := db.First(&stored, "id = ?", task.ID).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if stored.Status != KongfzCollectStatusCompleted || stored.CollectedCount != 2 || stored.ValidCount != 2 || stored.ImportedCount != 2 {
		t.Fatalf("task status/counts = %+v", stored)
	}
	var rawCount, bookCount, snapshotCount int64
	_ = db.Model(&models.KongfzCollectRawRow{}).Count(&rawCount).Error
	_ = db.Model(&models.Book{}).Count(&bookCount).Error
	_ = db.Model(&models.PriceSnapshot{}).Count(&snapshotCount).Error
	if rawCount != 2 || bookCount != 2 || snapshotCount != 2 {
		t.Fatalf("counts raw/book/snapshot=%d/%d/%d, want 2/2/2", rawCount, bookCount, snapshotCount)
	}
	var rawRow models.KongfzCollectRawRow
	if err := db.First(&rawRow, "normalized_isbn = ?", "9787506365437").Error; err != nil {
		t.Fatalf("load raw row: %v", err)
	}
	if rawRow.CoverURL == "" {
		t.Fatal("raw row cover url should be saved")
	}
}

func TestKongfzCollectorStopsOnBlockedResponse(t *testing.T) {
	db := setupAdminServiceTestDB(t)
	task, err := CreateKongfzCollectTask(KongfzCollectCreateInput{CatID: 43, StartPage: 1, EndPage: 1, DelayMS: 0, RetryTimes: 0, CreatedBy: 1})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	claimed, ok, err := claimNextKongfzCollectTask()
	if err != nil || !ok || claimed.ID != task.ID {
		t.Fatalf("claim task = %+v ok=%v err=%v", claimed, ok, err)
	}
	collector := NewKongfzCollector(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("需要验证码")), Header: http.Header{}}, nil
	})})
	if err := collector.RunTask(context.Background(), task.ID); err == nil {
		t.Fatal("run task should fail on blocked response")
	}
	var stored models.KongfzCollectTask
	if err := db.First(&stored, "id = ?", task.ID).Error; err != nil {
		t.Fatalf("load task: %v", err)
	}
	if stored.Status != KongfzCollectStatusFailed || !strings.Contains(stored.ErrorMessage, "风控") {
		t.Fatalf("task should be failed with blocked message: %+v", stored)
	}
}
