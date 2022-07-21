package handle

import (
	"context"
	"dy_feed_srv/global"
	model "dy_feed_srv/modle"
	"dy_feed_srv/proto"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type FeedSrv struct{}

//GetVideos(context.Context, *LatestTime) (*VideoInfoList, error)
//GetVideoCom(context.Context, *VideoBasicInfo) (*CommentList, error)
//BatchGetComment(context.Context, *CommentList) (*ContentList, error)
//CreateVideos(context.Context, *CreateVideoInfo) (*Empty, error)
//GetVideoList(context.Context, *VideoBasicInfo) (*VideoInfoList, error)
//UpdateVideoInfo(context.Context, *VideoBasicInfo) (*Empty, error)

//GetVideos 视频流
func (*FeedSrv) GetVideos(ctx context.Context, req *proto.LatestTime) (*proto.VideoInfoList, error) {
	videoList := make([]model.Video, 0)
	if result := global.DB.Where("update_time >= ?", req.Time).Find(&videoList); result.Error != nil {
		return nil, status.Error(codes.Internal, "获取视频异常")
	}
	rsp := &proto.VideoInfoList{}
	for _, video := range videoList {
		videoInfo := &proto.VideoInfo{
			Id:            video.Id,
			UserId:        video.UserID,
			PlayUrl:       video.PlayUrl,
			CoverUrl:      video.CoverUrl,
			Title:         video.Title,
			FavoriteCount: video.Favorite_count,
			CommentCount:  video.Comment_count,
			CommentList:   video.CommentList,
		}
		rsp.VideoInfos = append(rsp.VideoInfos, videoInfo)
	}
	return rsp, nil
}

//GetVideoComIdList 获取用户评论
func (*FeedSrv) GetVideoComIdList(ctx context.Context, req *proto.VideoBasicInfo) (*proto.CommentList, error) {
	video := model.Video{}
	if result := global.DB.Where(&model.Video{Id: req.VideoId[0]}).Find(&video); result.Error != nil {
		zap.S().Info(result.Error)
		return nil, status.Error(codes.Internal, "获取视频异常")
	}
	return &proto.CommentList{ComIdList: video.CommentList}, nil
}

func (*FeedSrv) BatchGetComment(ctx context.Context, req *proto.CommentList) (*proto.ContentList, error) {
	commentList := make([]model.Comment, 0)

	if result := global.DB.Where("id IN ?", req.ComIdList).Find(&commentList); result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	contentList := make([]*proto.Content, 0)
	for _, comment := range commentList {
		content := &proto.Content{
			UserId: comment.UserId,
			Str:    comment.Content,
		}
		contentList = append(contentList, content)
	}
	return &proto.ContentList{ContentList: contentList}, nil
}

func (*FeedSrv) CreateVideo(ctx context.Context, req *proto.CreateVideoInfo) (*proto.VideoBasicInfo, error) {
	//TODO  开启事务 修改user表里面的video_list  往video表里面新增一条记录  提交事务
	video := &model.Video{
		CreateTime: time.Now().Unix(),
		UpdateTime: time.Now().Unix(),
		UserID:   req.UserId,
		PlayUrl:  req.PlayUrl,
		CoverUrl: req.CoverUrl,
		Title:    req.Title,
	}
	//开启事务
	tx := global.DB.Begin()
	if result := tx.Create(video); result.Error != nil {
		tx.Rollback()
		zap.S().Error(result.Error.Error())
		return nil, status.Error(codes.Internal, result.Error.Error())
	}

	newVideoId := video.Id
	//修改user表里面的video_list
	rspVideoList, err := global.UserSrvCli.GetUserVideoList(context.Background(), &proto.UserBasicInfo{UserId: []int32{req.UserId}})
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal,"GetUserVideoList"+err.Error())
	}

	//把新的videoId加入到user 的videoList里面
	rspVideoList.VideoId = append(rspVideoList.VideoId, newVideoId)
	userInfoList, err := global.UserSrvCli.BatchGetUserDetail(context.Background(), &proto.UserBasicInfo{UserId: []int32{req.UserId}})
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "BatchGetUserDetail err: "+err.Error())
	}

	userInfoList.UserDetailInfoList[0].Videos = rspVideoList.VideoId
	_, err = global.UserSrvCli.UpdateUserInfo(context.Background(), userInfoList.UserDetailInfoList[0])
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "UpdateUserInfo err:"+err.Error())
	}

	tx.Commit()
	return &proto.VideoBasicInfo{VideoId: []int32{video.Id}}, nil
}

func (*FeedSrv) GetVideoList(ctx context.Context,req *proto.VideoBasicInfo) (*proto.VideoInfoList, error) {
	videoList := make([]model.Video, 0)
	result := global.DB.Where("id IN ?", req.VideoId).Find(&videoList)
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "一个都没找到")
	}
	ans := make([]*proto.VideoInfo, len(req.VideoId))
	for i, videoInfo := range videoList {
		ans[i] = &proto.VideoInfo{
			Id:            videoInfo.Id,
			UserId:        videoInfo.UserID,
			PlayUrl:       videoInfo.PlayUrl,
			CoverUrl:      videoInfo.CoverUrl,
			Title:         videoInfo.Title,
			FavoriteCount: videoInfo.Favorite_count,
			CommentCount:  videoInfo.Comment_count,
			CommentList:   videoInfo.CommentList,
		}
	}
	return &proto.VideoInfoList{VideoInfos: ans}, nil
}
func (*FeedSrv) UpdateVideoInfo(ctx context.Context, req *proto.VideoInfo) (*emptypb.Empty, error) {
	video := &model.Video{Id: req.Id}
	result := global.DB.First(&video)
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("没找到id为%d的video", req.Id))
	}
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	video = &model.Video{
		Id:             req.Id,
		UserID:         req.UserId,
		PlayUrl:        req.PlayUrl,
		CoverUrl:       req.CoverUrl,
		Title:          req.Title,
		Favorite_count: req.FavoriteCount,
		Comment_count:  req.CommentCount,
		CommentList:    req.CommentList,
	}
	global.DB.Save(video)
	return &emptypb.Empty{}, nil
}

func (*FeedSrv)	CommentAction(ctx context.Context, req *proto.CommentInfo) (*emptypb.Empty, error)  {

	video := &model.Video{}
	result := global.DB.First(video, req.VideoId)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "没找到")
	}
	if req.ActionType == 1 {
		//TODO 开启事务  先往comment表里面加一条记录  修改video表的comment_list字段
		//往comment表里面加一条记录
		tx := global.DB.Begin()
		comment := &model.Comment{
			UserId:     req.UserId,
			Content:    req.CommentText,
		}
		if result := tx.Create(comment); result.Error != nil {
			return nil, status.Error(codes.Internal, "Create comment err:"+result.Error.Error())
		}
		//修改video表的comment_list 字段
		video.Comment_count++
		video.CommentList = append(video.CommentList, int32(comment.ID))
		if result := tx.Updates(video); result.Error != nil {
			return nil, status.Error(codes.Internal, "Updates video err:"+result.Error.Error())
		}
		tx.Commit()
	} else {
		//TODO 开启事务  删除comment表里一条记录  修改video表的comment_list字段
		tx := global.DB.Begin()
		comment := &model.Comment{}
		comment.ID = uint(req.CommentId)
		if result := tx.Delete(comment); result.Error != nil {
			return nil, status.Error(codes.Internal, "Delete comment err:"+result.Error.Error())
		}
		//修改video表的comment_list 字段
		video.Comment_count--
		for i, id := range video.CommentList {
			if id == req.CommentId {
				video.CommentList = append(video.CommentList[:i], video.CommentList[i+1:]...)
				break
			}
		}
		if result := tx.Save(video); result.Error != nil {
			return nil, status.Error(codes.Internal, "Save video err:"+result.Error.Error())
		}
		tx.Commit()
	}

	return &emptypb.Empty{}, nil
}
