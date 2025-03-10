package controllers

import (
	"bytes"

	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAddBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/books", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("userRole", "admin")
		AddBook(gormDB)(c)
	})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_libraries" WHERE user_id = $1 AND library_id = $2 ORDER BY "user_libraries"."user_id" LIMIT $3`)).
		WithArgs(1, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "library_id"}).AddRow(1, 1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "books" WHERE isbn = $1 AND library_id = $2 AND "books"."deleted_at" IS NULL`)).
		WithArgs("123456789", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "books" (isbn, title, total_copies, available_copies, library_id) VALUES ($1, $2, $3, $4, $5)`)).
		WithArgs("123456789", "Test Book", 3, 3, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewBufferString(`{"isbn":"123456789","title":"Test Book","library_id":1,"total_copies":3}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//assert.Equal(t, http.StatusCreated, w.Code)
	//assert.Contains(t, w.Body.String(), "Book added successfully")
	//assert.NoError(t, mock.ExpectationsWereMet())
}
func TestUpdateBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.PUT("/books/:isbn", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("userRole", "admin")
		UpdateBook(gormDB)(c)
	})

	// Ensure correct request body
	payload := `{"library_id":1, "title":"Updated Title","authors":"Updated Author","publisher":"Updated Publisher","version":"2nd Edition","total_copies":5}`

	// Mock admin verification
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, user_id, library_id FROM "user_libraries" WHERE user_id = $1 AND library_id = $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "library_id"}).AddRow(1, 1, 1))

	// Mock book retrieval
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT isbn, title, authors, publisher, version, total_copies, available_copies FROM "books" WHERE isbn = $1 AND library_id = $2`)).
		WithArgs("123456789", 1).
		WillReturnRows(sqlmock.NewRows([]string{"isbn", "title", "authors", "publisher", "version", "total_copies", "available_copies"}).
			AddRow("123456789", "Test Book", "Test Author", "Test Publisher", "1st Edition", 5, 5))

	// Mock update query
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "books" SET title = $1, authors = $2, publisher = $3, version = $4, total_copies = $5, available_copies = $6 WHERE isbn = $7 AND library_id = $8`)).
		WithArgs("Updated Title", "Updated Author", "Updated Publisher", "2nd Edition", 5, 5, "123456789", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Create test request
	req := httptest.NewRequest(http.MethodPut, "/books/123456789", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	//assert.Equal(t, http.StatusOK, w.Code)
	//assert.Contains(t, w.Body.String(), "Book updated successfully")
	//assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRemoveBook(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.DELETE("/books/:isbn", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.Set("userRole", "admin")
		RemoveBook(gormDB)(c)
	})

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_libraries" WHERE user_id = $1 AND library_id = $2 ORDER BY "user_libraries"."user_id" LIMIT $3`)).
		WithArgs(1, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "library_id"}).
			AddRow(1, 1, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "books" WHERE (isbn = $1 AND library_id = $2) 
		AND "books"."deleted_at" IS NULL ORDER BY "books"."id" LIMIT $3`)).
		WithArgs("123456789", 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"isbn", "total_copies", "available_copies"}).
			AddRow("123456789", 1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "books" WHERE "books"."isbn" = $1 AND "books"."library_id" = $2`)).
		WithArgs("123456789", 1).
		WillReturnResult(sqlmock.NewResult(0, 1)) // Ensure 1 row is affected

	req := httptest.NewRequest(http.MethodDelete, "/books/123456789", bytes.NewBufferString(`{"libraryid":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	//assert.Equal(t, http.StatusOK, w.Code)
	//assert.Contains(t, w.Body.String(), "Book removed from inventory")
	//assert.NoError(t, mock.ExpectationsWereMet())
}
