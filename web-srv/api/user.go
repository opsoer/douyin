package api

import (
	"context"
	"dy_web_srv/middlewares"
	"dy_web_srv/models"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"strconv"
	"time"

	"dy_web_srv/global"
	"dy_web_srv/proto"
)

func HandleGrpcErrorToHttp(err error, c *gin.Context) {
	//将grpc的code转换成http的状态码
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{
					"msg": e.Message(),
				})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg:": "内部错误",
				})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{
					"msg": "参数错误",
				})
			case codes.Unavailable:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": "服务不可用",
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"msg": e.Code(),
				})
			}
			return
		}
	}
}

//Register 用户注册
func Register(c *gin.Context) {
	//用户注册
	name := c.Query("username")
	password := c.Query("password")
	rsp, err := global.UserSrvCli.CreateUser(context.Background(), &proto.CreateUserInfo{
		Username: name,
		Password: password,
	})
	if err != nil {
		status.FromError(err)
		zap.S().Info("创建用户失败：err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	zap.S().Info("创建用户成功")
	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("创建用户id为%d", rsp.UserId),
	})
}

//Login 用户登录
func Login(c *gin.Context) {
	//用户注册
	name := c.Query("username")
	password := c.Query("password")

	rsp, err := global.UserSrvCli.CheckPassWord(context.Background(), &proto.PasswordCheckInfo{
		Username: name,
		Password: password,
	})
	if err != nil {
		status.FromError(err)
		zap.S().Info("登录失败：err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("登录失败: %v", err),
		})
		return
	}
	if !rsp.Success {
		zap.S().Info("登录失败：err: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("登录失败: %v", err),
		})
		return
	}
	j := middlewares.NewJWT()
	claims := models.CustomClaims{
		UserId: rsp.Id,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),               //签名的生效时间
			ExpiresAt: time.Now().Unix() + 60*60*24*30, //30天过期
		},
	}
	token, err := j.CreateToken(claims)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成token失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":         rsp.Id,
		"token":      token,
		"expired_at": (time.Now().Unix() + 60*60*24*30) * 1000, //把jwt的过期时间放浏览器
		"message":    "登陆成功",
	})
}

//UserInfo 用户信息
func UserInfo(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请传入正确的ID参数",
		})
		return
	}
	rsp, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{
		UserId: []int32{int32(userId)},
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":             rsp.UserDetailInfoList[0].Id,
		"name":           rsp.UserDetailInfoList[0].Name,
		"follow_count":   rsp.UserDetailInfoList[0].FollowCount,
		"follower_count": rsp.UserDetailInfoList[0].FollowerCount,
	})
}

func PublishList(c *gin.Context) {
	//登录用户的视频发布列表，直接列出用户所有投稿过的视频
	userIdStr := c.Query("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请传入正确的ID参数",
		})
		return
	}
	//TODO videoIdList
	videList, err := global.UserSrvCli.GetUserVideoList(context.Background(), &proto.UserBasicInfo{
		UserId: []int32{int32(userId)},
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	zap.S().Info(videList.VideoId)
	//TODO 需要请求feed-srv获取video的详细信息
	videoInfoList, err := global.FeedSrvCli.GetVideoList(context.Background(), &proto.VideoBasicInfo{VideoId: videList.VideoId})
	if err != nil {
		zap.S().Error(err)
		HandleGrpcErrorToHttp(err, c)
		return
	}
	userList, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{int32(userId)}})
	if err != nil {
		zap.S().Error(err)
		HandleGrpcErrorToHttp(err, c)
		return
	}
	userInfo := userList.UserDetailInfoList[0]
	author := map[string]interface{}{
		"id":userInfo.Id,
		"name":userInfo.Name,
		"follow_count":userInfo.FollowCount,
		"follower_count":userInfo.FollowerCount,
		//TODO  自己关注自己？？
		"is_follow":false,
	}

	//组装返回
	ansList := make(map[string][]map[string]interface{})
	mpList := make([]map[string]interface{}, len(videList.VideoId))
	for i, videoInfo := range videoInfoList.VideoInfos {
		mp := make(map[string]interface{})
		mp["id"]  = videoInfo.Id
		mp["author"] = author
		mp["play_url"] = videoInfo.PlayUrl
		mp["cover_url"] = videoInfo.CoverUrl
		mp["favorite_count"] = videoInfo.FavoriteCount
		mp["comment_count"] = videoInfo.CommentCount
		//自己的视频默认不能点赞
		mp["is_favorite"] = false
		mp["title"] = videoInfo.Title
		mpList[i] = mp
		zap.S().Info(mp)
	}
	ansList["video_list"] = mpList
	zap.S().Info(ansList)
	c.JSON(http.StatusOK, ansList)
}

func FollowList(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请传入正确的ID参数",
		})
		return
	}
	rsp, err := global.UserSrvCli.GetUserFollows(context.Background(), &proto.UserBasicInfo{
		UserId: []int32{int32(userId)},
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	allFollowList, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: rsp.FollowId})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	ansMap := make([]map[string]interface{}, 0)
	for _, userInfo := range allFollowList.UserDetailInfoList {
		mp := map[string]interface{}{}
		mp["id"] = userInfo.Id
		mp["name"] = userInfo.Name
		mp["follow_count"] = userInfo.FollowCount
		mp["follower_count"] = userInfo.FollowerCount
		//查询出来的都是自己的关注
		mp["is_follow"] = true
		ansMap = append(ansMap, mp)
	}
	c.JSON(http.StatusOK, gin.H{
		"user_list": ansMap,
	})
}

func FollowerList(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请传入正确的ID参数",
		})
		return
	}
	rsp, err := global.UserSrvCli.GetUserFollowers(context.Background(), &proto.UserBasicInfo{
		UserId: []int32{int32(userId)},
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	allFollowerList, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: rsp.FollowerId})
	if err != nil {
		HandleGrpcErrorToHttp(err, c)
		return
	}
	ansMap := make([]map[string]interface{}, 0)
	for _, userInfo := range allFollowerList.UserDetailInfoList {
		mp := map[string]interface{}{}
		mp["id"] = userInfo.Id
		mp["name"] = userInfo.Name
		mp["follow_count"] = userInfo.FollowCount
		mp["follower_count"] = userInfo.FollowerCount
		//查一下粉丝follower里面是否有我的id，有就为true
		rsp, err := global.UserSrvCli.GetUserFollowers(context.Background(), &proto.UserBasicInfo{
			UserId: []int32{userInfo.Id},
		})
		if err != nil {
			HandleGrpcErrorToHttp(err, c)
			return
		}
		is_follow := false
		for _, id := range rsp.FollowerId {
			if id == int32(userId) {
				is_follow = true
			}
		}

		mp["is_follow"] = is_follow
		ansMap = append(ansMap, mp)
	}
	c.JSON(http.StatusOK, gin.H{
		"user_list": ansMap,
	})
}

func FavoriteList(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请传入正确的user_id参数 err: " + err.Error(),
		})
		return
	}
	//TODO  先查询出所有喜欢的video的id  再通过videoId 去video表里面查视频
	//查询出所有喜欢的video的id
	userInfoList, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{int32(userId)}})
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "调用userSrv BatchGetUserDetail err: " + err.Error(),
		})
		return
	}
	userFavList := userInfoList.UserDetailInfoList[0].FavList
	zap.S().Info(userFavList)
	//通过videoId 去video表里面查视频
	videoInfoList, err := global.FeedSrvCli.GetVideoList(context.Background(), &proto.VideoBasicInfo{VideoId: userFavList})
	zap.S().Info(videoInfoList.VideoInfos)
	ansMap := make(map[string]interface{})
	videoList := make([]interface{}, 0)
	if err != nil {
		zap.S().Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "调用FeedSrv GetVideoInfo err: " + err.Error(),
		})
		return
	}
	for _, videoInfo := range videoInfoList.VideoInfos {
		//查询 video作者信息
		userInfoList, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{videoInfo.UserId}})
		if err != nil {
			zap.S().Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "调用UserSrv BatchGetUserDetail err: " + err.Error(),
			})
			return
		}
		zap.S().Info(userInfoList.UserDetailInfoList)
		//判断该商品作者是不是我的关注  即看我的follows里面有没有该作者的id
		isFollow := false
		myInfo, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{int32(userId)}})
		if err != nil {
			zap.S().Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "调用UserSrv BatchGetUserDetail err: " + err.Error(),
			})
			return
		}
		zap.S().Info(myInfo.UserDetailInfoList)
		for _, id := range myInfo.UserDetailInfoList[0].Follows {
			if id == videoInfo.UserId {
				isFollow = true
			}
		}

		videoList = append(videoList, map[string]interface{}{
			"id": videoInfo.Id,
			"author": map[string]interface{}{
				"id":             userInfoList.UserDetailInfoList[0].Id,
				"name":           userInfoList.UserDetailInfoList[0].Name,
				"follow_count":   userInfoList.UserDetailInfoList[0].FollowCount,
				"follower_count": userInfoList.UserDetailInfoList[0].FollowerCount,
				"is_follow": isFollow,
			},
			"play_url":       videoInfo.PlayUrl,
			"cover_url":      videoInfo.CoverUrl,
			"favorite_count": videoInfo.FavoriteCount,
			"comment_count":  videoInfo.CommentCount,
			"is_favorite":    videoInfo.FavoriteCount,
			"title":          videoInfo.Title,
		})
	}
	ansMap["video_list"] = videoList
	c.JSON(http.StatusOK, ansMap)
}