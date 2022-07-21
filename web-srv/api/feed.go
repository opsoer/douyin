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
	"time"
)

func FeedVideo(c *gin.Context) {
	latestTimeStr := c.Query("latest_time")
	latestTime, err := strconv.ParseInt(latestTimeStr, 10, 64)
	if err != nil {
		latestTime = 1653379096
	}
	rsp, err := global.FeedSrvCli.GetVideos(context.Background(), &proto.LatestTime{Time: latestTime})
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusInternalServerError, feedVideoRsp{
				baseRsp:   baseRsp{
					statusCode: 1,
					statusMsg:  "请求视频失败",
				},
				nextTime:  0,
				videoList: nil,
			},
		)
		return
	}
	feedVideoRsp := feedVideoRsp{}
	ansMap := make([]map[string]interface{}, 0)
	for _, video := range rsp.VideoInfos {
		feedVideoRsp.videoList = append(feedVideoRsp.videoList, videoRsp{
			id:            video.Id,
			playUrl:       video.PlayUrl,
			coverUrl:     video.CoverUrl,
			favoriteCount: video.FavoriteCount,
			commentCount:  video.CommentCount,
			isFavorite:    false,
			title:        video.Title,
		})
		mp := map[string]interface{}{}
		mp["id"] = video.Id
		mp["play_url"] = video.PlayUrl
		mp["cover_url"] = video.CoverUrl
		mp["favorite_count"] = video.FavoriteCount
		mp["comment_count"] = video.CommentCount
		mp["title"] = video.Title
		//TODO 先默认为false
		mp["is_favorite"] = false
		userId := video.UserId
		rsp, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{userId}})
		if err != nil {
			zap.S().Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}
		user := rsp.UserDetailInfoList[0]
		mp["author"] = user
		ansMap = append(ansMap, mp)
	}
	c.JSON(http.StatusOK, gin.H{

		"video_list": ansMap,
		"next_time":  time.Now().Unix(),
	})
}

func CommentList(c *gin.Context) {
	videoIdStr := c.Query("video_id")
	videoId, err := strconv.Atoi(videoIdStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err,
		})
		return
	}
	rsp, err := global.FeedSrvCli.GetVideoComIdList(context.Background(), &proto.VideoBasicInfo{VideoId: []int32{int32(videoId)}})
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err,
		})
		return
	}
	zap.S().Info(rsp.ComIdList)
	comRsp, err := global.FeedSrvCli.BatchGetComment(context.Background(), &proto.CommentList{ComIdList: rsp.ComIdList})
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message:": err,
		})
		return
	}
	ansMap := make(map[string]interface{})
	//TODO 评论的id 创建时间 感觉返回没什么用，就没返回了
	commentList := make([]map[string]interface{}, 0)
	for _, content := range comRsp.ContentList {
		comment := map[string]interface{}{}
		comment["id"] = content.UserId
		comment["content"] = content.Str
		rsp, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{content.UserId}})
		if err != nil {
			zap.S().Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err,
			})
			return
		}
		comment["user"] = rsp.UserDetailInfoList[0]
		commentList = append(commentList, comment)
	}
	ansMap["comment_list"] = commentList
	c.JSON(http.StatusOK, ansMap)
}

func CommentAction(c *gin.Context) {
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
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message:": "video_id 参数错误 err: " + err.Error(),
		})
		return
	}
	actionTypeStr := c.Query("action_type")
	actionType, err := strconv.Atoi(actionTypeStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message:": "action_type 参数错误 err: " + err.Error(),
		})
		return
	}
	commentText := c.Query("comment_text")
	commentIdStr := c.Query("comment_id")
	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message:": "comment_id 参数错误 err: " + err.Error(),
		})
		return
	}
	_, err = global.FeedSrvCli.CommentAction(context.Background(), &proto.CommentInfo{
		UserId:      userId,
		VideoId:     int32(videoId),
		CommentId:   int32(commentId),
		ActionType:  int32(actionType),
		CommentText: commentText,
	})
	if err != nil {
		zap.S().Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message:": "UserSrv CommentAction err: " + err.Error(),
		})
		return
	}

	//判断是不是我的关注
	userList, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{int32(userId)}})
	userInfo := userList.UserDetailInfoList[0]
	isFollow := false
	videoList, err := global.FeedSrvCli.GetVideoList(context.Background(), &proto.VideoBasicInfo{VideoId: []int32{int32(videoId)}})
	if err != nil {
		zap.S().Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message:": "FeedSrvCli GetVideoList err: " + err.Error(),
		})
		return
	}
	if len(videoList.VideoInfos) == 0 {
		zap.S().Error(err.Error())
		c.JSON(http.StatusNotFound, gin.H{
			"message:": "UserSrv GetVideoList notFound: " + err.Error(),
		})
		return
	}
	toUserId := videoList.VideoInfos[0].UserId
	for _, id := range userInfo.Follows {
		if id == toUserId {
			isFollow = true
			break
		}
	}

	ansMap := make(map[string]interface{})
	ansMap["0"] = map[string]interface{}{
		"id":3,
		"user":  map[string]interface{}{
			"id": userInfo.Id,
			"name": userInfo.Name,
			"follow_count": userInfo.FollowCount,
			"follower_count": userInfo.FollowerCount,
			"is_follow": isFollow,
		},
		"content": commentText,
		"create_date":time.Now(),
	}
	c.JSON(http.StatusBadRequest, ansMap)
}
