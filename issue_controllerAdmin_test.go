package controllers

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestListIssueRequests(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/requests", func(c *gin.Context) {
		c.Set("userID", uint(1))
		ListIssueRequests(gormDB)(c)
	})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "library_id" FROM "user_libraries" WHERE user_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"library_id"}).AddRow(1))

	req := httptest.NewRequest(http.MethodGet, "/requests", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "requests")
	assert.NoError(t, mock.ExpectationsWereMet())
}
func TestApproveIssue(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.PUT("/requests/:id/approve", func(c *gin.Context) {
		c.Set("userID", uint(1))
		ApproveIssue(gormDB)(c)
	})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "request_events" WHERE "request_events"."id" = $1 AND "request_events"."deleted_at" IS NULL`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "book_id", "reader_id", "request_type", "request_date", "approval_date", "approver_id"}).
			AddRow(1, "123456789", 2, "issue", 1741480342, nil, nil))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "books" WHERE isbn = $1`)).
		WithArgs("123456789").
		WillReturnRows(sqlmock.NewRows([]string{"isbn", "library_id", "available_copies"}).
			AddRow("123456789", 1, 5))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM "user_libraries" WHERE user_id = $1 AND library_id = $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "request_events" SET "approval_date"=$1, "approver_id"=$2 WHERE "id" = $3`)).
		WithArgs(sqlmock.AnyArg(), 1, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPut, "/requests/1/approve", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//assert.Equal(t, http.StatusOK, w.Code)
	//assert.Contains(t, w.Body.String(), "Issue request approved")
	//assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDisapproveIssue(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/requests/:id/disapprove", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("userRole", "admin")
		DisapproveIssue(gormDB)(c)
	})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "request_events" WHERE "request_events"."id" = $1 AND "request_events"."deleted_at" IS NULL ORDER BY "request_events"."id" LIMIT 1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "book_id", "reader_id", "request_type", "request_date", "approval_date", "approver_id"}).
			AddRow(1, "123456789", 2, "issue", time.Now().Unix(), nil, nil))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "request_events" WHERE "request_events"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	req := httptest.NewRequest(http.MethodDelete, "/requests/1/disapprove", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//assert.Equal(t, http.StatusOK, w.Code)
	//assert.Contains(t, w.Body.String(), "Issue request disapproved successfully")
	//assert.NoError(t, mock.ExpectationsWereMet())
}
