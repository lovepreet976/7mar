package controllers

import (
	"bytes"
	"errors"
	"fmt"

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

func TestCreateLibrary(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/libraries", CreateLibrary(gormDB))

	t.Run("Successful Library Creation", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "libraries" ("name","location") VALUES ($1,$2)`)).
			WithArgs("Test Library", "City Center").
			WillReturnResult(sqlmock.NewResult(1, 1)) // Ensure one row is affected

		req := httptest.NewRequest(http.MethodPost, "/libraries", bytes.NewBufferString(`{"name":"Test Library","location":"City Center"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		//assert.Equal(t, http.StatusCreated, w.Code)
		//assert.Contains(t, w.Body.String(), "Library created successfully")
		//assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Invalid Input", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/libraries", bytes.NewBufferString(`{"name":""}`)) // Missing required field
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		//assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Database Error", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "libraries" ("name","location") VALUES ($1,$2)`)).
			WithArgs("Library X", "Downtown").
			WillReturnError(fmt.Errorf("database error"))

		req := httptest.NewRequest(http.MethodPost, "/libraries", bytes.NewBufferString(`{"name":"Library X","location":"Downtown"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Could not create library")
		//assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestListLibraries(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	assert.NoError(t, err)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/libraries", ListLibraries(gormDB))

	t.Run("Successful Fetch", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "libraries"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Central Library"))

		req := httptest.NewRequest(http.MethodGet, "/libraries", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Central Library")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Database Error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "libraries"`)).
			WillReturnError(errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/libraries", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Could not fetch libraries")
	})
}
