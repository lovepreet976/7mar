package controllers

import (
	"fmt"
	"library-management/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterOwnerNew allows an existing owner to create a new owner
func RegisterOwnerNew(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.User

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if input.Role != "owner" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role, must be 'owner'"})
			return
		}

		if err := db.Create(&input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create owner"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "New owner registered successfully", "owner": input})
	}
}

// RegisterAdmin allows an owner to create an admin
func RegisterAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name       string `json:"name" binding:"required"`
			Email      string `json:"email" binding:"required,email"`
			Password   string `json:"password" binding:"required"`
			Contact    string `json:"contact"`
			LibraryIDs []uint `json:"library_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		creatorID := c.GetUint("userID")
		var creator models.User
		if err := db.First(&creator, creatorID).Error; err != nil || creator.Role != "owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only an owner can create an admin"})
			return
		}

		admin := models.User{
			Name:     input.Name,
			Email:    input.Email,
			Password: input.Password,
			Contact:  input.Contact,
			Role:     "admin",
		}

		if err := db.Create(&admin).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create admin"})
			return
		}

		for _, libID := range input.LibraryIDs {
			var library models.Library
			if err := db.First(&library, libID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Library ID %d not found", libID)})
				return
			}
			adminLibrary := models.UserLibrary{UserID: admin.ID, LibraryID: libID}
			if err := db.Create(&adminLibrary).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate admin with library"})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Admin registered successfully", "admin": admin})
	}
}

// RegisterUser allows an admin to create a user
func RegisterUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Name       string `json:"name" binding:"required"`
			Email      string `json:"email" binding:"required,email"`
			Password   string `json:"password" binding:"required"`
			Contact    string `json:"contact"`
			LibraryIDs []uint `json:"library_ids" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		adminID := c.GetUint("userID")
		var admin models.User
		if err := db.First(&admin, adminID).Error; err != nil || admin.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create users"})
			return
		}

		var adminLibraries []uint
		if err := db.Table("user_libraries").Where("user_id = ?", adminID).Pluck("library_id", &adminLibraries).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not verify admin libraries"})
			return
		}

		for _, libID := range input.LibraryIDs {
			found := false
			for _, adminLibID := range adminLibraries {
				if libID == adminLibID {
					found = true
					break
				}
			}
			if !found {
				c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("You can only add users to libraries you manage (Library ID: %d)", libID)})
				return
			}
		}

		user := models.User{
			Name:     input.Name,
			Email:    input.Email,
			Password: input.Password,
			Contact:  input.Contact,
			Role:     "user",
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not register user"})
			return
		}

		for _, libID := range input.LibraryIDs {
			userLibrary := models.UserLibrary{UserID: user.ID, LibraryID: libID}
			if err := db.Create(&userLibrary).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate user with library"})
				return
			}
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "user": user})
	}
}
