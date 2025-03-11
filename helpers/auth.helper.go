package helpers

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func CheckUserType(c *gin.Context, role string) (err error) {
	err = nil
	user_type := c.GetString("user_type")
	if user_type != role {
		err = errors.New("Unauthorized access to this resource")
		return err
	}
	return err
}

func MatchUserTypeToUid(c *gin.Context, user_id string) (err error) {
	user_type := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil

	if user_type == "USER" && user_id != uid {
		err = errors.New("Unauthorized access to this resource")
		return err
	}

	err = CheckUserType(c, user_type)
	return err
}
