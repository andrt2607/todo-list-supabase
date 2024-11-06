package middlewares

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

	"todo-list/db"
	"todo-list/models"
	"todo-list/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func IsAuth(c *gin.Context) {

	var token string
	var user models.UserModel

	token = c.GetHeader("Authorization")
	if token == "" {
		utils.ThrowErr(c, http.StatusUnauthorized, "Need Header Auth")
		c.Abort()
		return
	}

	header := strings.Split(token, " ")
	tokenHeader := header[0]
	if tokenHeader != "Bearer" {
		utils.ThrowErr(c, http.StatusUnauthorized, "Invalid Header Auth")
		c.Abort()
		return
	}

	tokenHeader = header[1]
	if tokenHeader == "" {
		utils.ThrowErr(c, http.StatusUnauthorized, "Invalid Header Auth")
		c.Abort()
		return
	}

	decode_token, err := jwt.Parse(tokenHeader, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		utils.ThrowErr(c, http.StatusUnauthorized, "Invalid Token")
		c.Abort()
		return
	}

	userId := decode_token.Claims.(jwt.MapClaims)["userId"].(string)

	if err = db.DB.QueryRow("SELECT id, username, password, name, role, is_active, created_at, updated_at FROM users WHERE id = $1", userId).Scan(&user.Id, &user.Username, &user.Password, &user.Name, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			utils.ThrowErr(c, http.StatusUnauthorized, "User not found")
		} else {
			utils.ThrowErr(c, http.StatusInternalServerError, "Failed to get user data")
		}
		c.Abort()
		return
	}

	if !user.IsActive {
		utils.ThrowErr(c, http.StatusUnauthorized, "User is not active")
		c.Abort()
		return
	}

	c.Set("user", user)
	c.Set("userId", userId)
	c.Set("role", user.Role)
	c.Next()
}
