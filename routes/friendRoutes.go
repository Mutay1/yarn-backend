package routes

import (
	controller "github.com/Mutay1/chat-backend/controllers"
	"github.com/gin-gonic/gin"
)

//FriendRoutes Function
func FriendRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/friends", controller.GetFriends())
	incomingRoutes.POST("/friends/archive", controller.Archive())
	incomingRoutes.POST("/friends/favorite", controller.Favorite())
	incomingRoutes.POST("/friends/block", controller.Block())
}
