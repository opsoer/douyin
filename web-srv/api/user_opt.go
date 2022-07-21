package api

import (
	"context"
	"dy_web_srv/global"
	"dy_web_srv/middlewares"
	"dy_web_srv/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

//PublishAction 用户发布视频
func PublishAction(c *gin.Context) {
	videoUrl := c.Query("data")

	token := c.Query("token")
	j := middlewares.NewJWT()
	claim, err := j.ParseToken(token)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "token err: " + err.Error(),
		})
		return
	}
	userId := claim.UserId
	title := c.Query("title")

	//往video表里面新增一条记录
	_, err = global.FeedSrvCli.CreateVideo(context.Background(), &proto.CreateVideoInfo{
		UserId: userId,
		//TODO 这里应该先将data解析为string类型的url(data里面获取) 再进行存储

		PlayUrl:  videoUrl,
		CoverUrl: "www.cover.com",
		Title:    title,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":"发布成功",
	})
}

//FavoriteAction 用户点赞操作
func FavoriteAction(c *gin.Context) {
	token := c.Query("token")
	j := middlewares.NewJWT()
	claim, err := j.ParseToken(token)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "token err: " + err.Error(),
		})
		return
	}

	userId := claim.UserId
	videoIdStr := c.Query("video_id")
	videoId, err := strconv.Atoi(videoIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message:": "请输入正确的video_id参数"+err.Error(),
		})
	}

	actionTypeStr := c.Query("action_type")
	actionType, _ := strconv.Atoi(actionTypeStr)
	_, err = global.UserSrvCli.FavoriteAction(context.Background(), &proto.FavInfo{
		UserId:  userId,
		VideoId: int32(videoId),
		ActionType: int32(actionType),
	})
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message:": "调用 userSrv FavoriteAction err: "+err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "点赞成功",
	})
}

//RelationAction 关注操作
func RelationAction(c *gin.Context) {
	//TODO 开启事务 修改user表的follow_count follower_count follow_list follower_list  提交事务
	//通过jwt获取userId
	token := c.Query("token")
	j := middlewares.NewJWT()
	claim, err := j.ParseToken(token)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "token err: " + err.Error(),
		})
		return
	}
	userId := claim.UserId

	toUserIdStr := c.Query("to_user_id")
	toUserId, err := strconv.Atoi(toUserIdStr)
	if err != nil {
		zap.S().Error("请输入正确的to_user_id")
		c.JSON(http.StatusBadRequest, gin.H{
			"message ": "请输入正确的to_user_id  "+err.Error(),
		})
		return
	}
	actionTypeStr := c.Query("action_type")
	actionType, err := strconv.Atoi(actionTypeStr)
	if err != nil {
		zap.S().Error("action_type")
		c.JSON(http.StatusBadRequest, gin.H{
			"message ": "请输入正确的action_type  "+err.Error(),
		})
		return
	}

	_, err = global.UserSrvCli.RelationAction(context.Background(), &proto.RelationActionInfo{
		UserId:     userId,
		ToUserId:   int32(toUserId),
		ActionType: int32(actionType),
	})
	if err!= nil {
		zap.S().Error("UserSrv RelationAction err:"+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"message ": "请输入正确的action_type  "+err.Error(),
		})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{
		"message ": "关注成功",
	})
	return
}
