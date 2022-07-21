package initialize

import (
	"dy_web_srv/api"
	"dy_web_srv/global"
	"dy_web_srv/middlewares"
)

func InitUserRouter() {
	apiRouter := global.Router.Group("/douyin").Use(middlewares.Cors())
	{
		//用户服务
		apiRouter.POST("/user/register", api.Register)//ok
		apiRouter.POST("/user/login", api.Login)//ok
		apiRouter.GET("/user", middlewares.JWTAuth(), api.UserInfo)//ok
		apiRouter.GET("/publish/list", middlewares.JWTAuth(), api.PublishList)//ok
		apiRouter.GET("/relation/follow/list/", middlewares.JWTAuth(), api.FollowList)//ok
		apiRouter.GET("/relation/follower/list/", middlewares.JWTAuth(), api.FollowerList)//ok
		//TODO  FavoriteList
		apiRouter.GET("/favorite/list/", middlewares.JWTAuth(), api.FavoriteList)//ok

		//feed服务
		apiRouter.GET("/feed/", api.FeedVideo)//ok
		//TODO  应该不需要登录把
		apiRouter.GET("/comment/list/", middlewares.JWTAuth(), api.CommentList)//ok
		//TODO  CommentAction
		apiRouter.POST("/comment/action/", middlewares.JWTAuth(), api.CommentAction)//ok

		//用户操作服务
		//TODO  PublishAction
		apiRouter.POST("/publish/action/", middlewares.JWTAuth(), api.PublishAction)//ok
		//TODO  FavoriteAction
		apiRouter.POST("/favorite/action/", middlewares.JWTAuth(), api.FavoriteAction)//ok
		//TODO  RelationAction
		apiRouter.POST("/relation/action/", middlewares.JWTAuth(), api.RelationAction)//ok

	}
}
