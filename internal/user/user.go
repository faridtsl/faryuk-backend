package user

import (
	"fmt"
	"time"

	"FaRyuk/internal/types"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// NewUser : returns new user from username and plain password
func NewUser(username, plainPwd string) *types.User {
	id := uuid.New().String()

	hashedpwd, err := bcrypt.GenerateFromPassword([]byte(plainPwd), bcrypt.DefaultCost)
	if err != nil {
		return nil
	}

	return &types.User{ID: id, Username: username, Password: string(hashedpwd)}
}

func GetHashedPassword(plainPwd string) (string, error) {
	hashedpwd, err := bcrypt.GenerateFromPassword([]byte(plainPwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedpwd), nil
}

// Login : checks if creds are valid
func Login(u *types.User, plainPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPwd))
	return err == nil
}

// GenerateJWT : generates a JWT token for a user
func GenerateJWT(usr *types.User, signingKey string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["username"] = usr.Username
	claims["id"] = usr.ID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(signingKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyJWT : verifies if JWT token is signed properly
func VerifyJWT(tokenStr, signingKey string) bool {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		return false
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return false
	}

	return true
}

// GetUsername : returns username from JWT token
func GetUsername(tokenStr, signingKey string) (string, string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		return "", "", err
	}

	claims := token.Claims.(jwt.MapClaims)
	return fmt.Sprintf("%s", claims["username"]), fmt.Sprintf("%s", claims["id"]), nil
}
