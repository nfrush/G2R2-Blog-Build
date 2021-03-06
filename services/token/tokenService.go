package servicesToken

import (
	"errors"
	"fmt"
	"time"

	r "github.com/dancannon/gorethink"
	"github.com/dchest/uniuri"
	"github.com/dgrijalva/jwt-go"
	"github.com/nfrush/G2R2-Blog-Build/database/rethink"
	"github.com/nfrush/G2R2-Blog-Build/models/token"
	"github.com/nfrush/G2R2-Blog-Build/models/user"
)

var session = rethink.GetSession()

//signingKey - Signing Key For Cookies
var signingKey = InitSigningKey()

//InitSigningKey - Initalize Our Key To Sign With
func InitSigningKey() string {
	return uniuri.NewLen(32)
}

//GetSigningKey - get the current signing key
func GetSigningKey() string {
	return signingKey
}

//IssueToken - Issue New JWT Token
func IssueToken(u *modelUser.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "Frush Development LTD",
		"aud": u.Username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"jti": "http://example.com",
	})

	tokenString, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return "", err
	}

	issuedToken := modelToken.JWT{Token: tokenString, Issuer: "Frush Development LTD", Audience: u.Username, IssuedAt: time.Now().Unix(), Expires: time.Now().Add(time.Hour * 72).Unix(), JTI: "http://example.com"}

	if err := r.Table("tokens").Insert(issuedToken).Exec(session); err != nil {
		return "", err
	}
	fmt.Println("Issued Token Successfully")
	return tokenString, nil
}

//RevokeToken - Revoke the JWT Token
func RevokeToken(u *modelUser.User) error {
	result, err := r.Table("tokens").Filter(map[string]interface{}{"Audience": u.Username}).Run(session)
	if err != nil {
		return err
	}
	var transformToken modelToken.JWT
	result.One(&transformToken)
	result.Close()

	if err := r.Table("tokens").Filter("Audience: u.Username").Delete().Exec(session); err != nil {
		return err
	}

	if err := r.Table("blacklist").Insert(&transformToken).Exec(session); err != nil {
		return err
	}

	return nil
}

//RefreshToken - Reissue a new token
func RefreshToken(u *modelUser.User) (string, error) {
	if err := RevokeToken(u); err != nil {
		return "error", err
	}
	token, errB := IssueToken(u)
	if errB != nil {
		return "error", errB
	}
	return token, nil
}

//TokenExists - Check if Token Exists
func TokenExists(token string) (bool, error) {
	if err := r.Table("tokens").Filter(map[string]interface{}{"Token": token}).Exec(session); err != nil {
		return false, err
	}
	return true, nil
}

//TokenExistsUser - Checks if a user has an assigned token
func TokenExistsUser(u *modelUser.User) (bool, error) {
	if err := r.Table("tokens").Filter(map[string]interface{}{"Audience": u.Username}).Exec(session); err != nil {
		return false, err
	}
	return true, nil
}

//RequiresAuth - Authenicates user on service
func RequiresAuth(token string) (bool, error) {
	exists, err := TokenExists(token)
	if err != nil {
		return false, err
	}
	if exists {
		res, err := r.Table("tokens").Filter(map[string]interface{}{"Token": token}).Run(session)
		if err != nil {
			return false, err
		}
		var transformToken modelToken.JWT
		res.One(&transformToken)
		res.Close()

		resu, err := r.Table("users").Filter(map[string]interface{}{"Name": transformToken.Audience}).Run(session)
		if err != nil {
			return false, err
		}
		var user modelUser.User
		resu.One(&user)
		resu.Close()

		if transformToken.Expires <= time.Now().Unix() {
			return true, nil
		}
		if transformToken.Expires > time.Now().Unix() {
			RevokeToken(&user)
			return false, errors.New("Token has expired and been revoked.")
		}
	}
	return false, errors.New("The Token Does Not Exist")
}
