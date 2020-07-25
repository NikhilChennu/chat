package jwt

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
)

//Token - strict for maintaining token information
type Token struct {
	UserId string `json:"userId"`
	Mobile string `json:"mobile"`
	jwt.StandardClaims
}

//GetToken - Handler for getting token
func GetToken(userId string, mobile string, tokenSecret string) string {
	fmt.Println("Creating Jwt token ", userId, mobile)
	tk := &Token{UserId: userId, Mobile: mobile}
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tk)
	tokenString, _ := token.SignedString([]byte(tokenSecret))
	return tokenString
}

//ValidateToken - function for validating token
func ValidateToken(token string, tokenSecret string) (*string, *string, error) {
	tk := &Token{}

	tokenResp, err := jwt.ParseWithClaims(token, tk, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("Malformed authentication token")
	}

	if !tokenResp.Valid {
		return nil, nil, fmt.Errorf("Token is not valid")
	}
	return &tk.UserId, &tk.Mobile, nil
}
