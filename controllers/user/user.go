package controllerUser

import (
	"fmt"
	"regexp"

	"github.com/labstack/echo"
	"github.com/graphql-go/handler"
	"github.com/nfrush/G2R2-Blog-Build/models/user"
	"github.com/nfrush/G2R2-Blog-Build/services/token"
	"github.com/nfrush/G2R2-Blog-Build/services/user"
)

//CreateUser creates a new user
func CreateUser(c echo.Context) error {
	u := &modelUser.User{}

	if err := c.Bind(u); err != nil {
		return c.JSON(409, err)
	}

	if err := servicesUser.CreateUser(u); err != nil {
		return c.JSON(409, err)
	}
	fmt.Println("Create User Successful")

	return c.JSON(200, u)
}

//UpdateUser updates a single user
func UpdateUser(c echo.Context) error {
	if c.Request().Header().Get("Authorization") != "" {
		r, _ := regexp.Compile("((Bearer )*)")
		var token = r.ReplaceAllString(c.Request().Header().Get("Authorization"), "")
		auth, err := servicesToken.RequiresAuth(token)
		if err != nil {
			return c.JSON(409, err)
		}
		if auth {
			u := &modelUser.User{}

			if err := c.Bind(u); err != nil {
				return c.JSON(409, err)
			}
			fmt.Println("Bind Successful")

			if err := servicesUser.UpdateUser(u); err != nil {
				return c.JSON(409, err)
			}
			fmt.Println("User Update Successful")

			return c.JSON(200, nil)
		}
		return c.JSON(409, "Error")
	}
	return c.JSON(409, "Error")
}

//DeleteUser deletes the specified user
func DeleteUser(c echo.Context) error {
	if c.Request().Header().Get("Authorization") != "" {
		r, _ := regexp.Compile("((Bearer )*)")
		var token = r.ReplaceAllString(c.Request().Header().Get("Authorization"), "")
		auth, err := servicesToken.RequiresAuth(token)
		if err != nil {
			return c.JSON(409, err)
		}
		if auth {
			user := servicesUser.DeleteUser(c.Param("username"))
			return c.JSON(200, user)
		}
		return c.JSON(409, "Error")
	}
	return c.JSON(409, "Error")
}

//GraphQLUser handles specified GraphQL queries
func GraphQLUser := handler.New(&handler.Config{
		Schema: &starwars.Schema,
		Pretty: true,
	})
