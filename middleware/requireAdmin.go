package middleware

import (
	"carport-backend/initializers"
	"carport-backend/models"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequireAdmin(c *gin.Context) {
	// Get the cookie of req
	// tokenString, err := c.Cookie("Authorization")
	// if err != nil {
	// 	c.AbortWithStatus(http.StatusUnauthorized)
	// }

	// Get the Authorization header from the request
	tokenString := c.GetHeader("Authorization")

	// Check if the Authorization header is present
	if tokenString == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Token is usually sent as "Bearer <token>", so we split to get the actual token
	splitToken := strings.Split(tokenString, "Bearer ")
	if len(splitToken) != 2 {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	tokenString = splitToken[1]

	// Decode/validate it
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check the exp
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Find the user with token sub
		var user models.User
		err := initializers.DB.Preload("Addresses").First(&user, claims["sub"]).Error

		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		if user.Role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "This user is not an admin", "details": err.Error()})
			return
		}

		// Attach to req
		c.Set("user", user)

		// continue
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

}
