package handle

import (
	"context"
	"crypto/sha512"
	"dy_uer_srv/global"
	model "dy_uer_srv/modle"
	"dy_uer_srv/proto"
	"fmt"
	"github.com/anaskhan96/go-password-encoder"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"strings"
)

type UserServer struct{}

//CreateUser(context.Context, *CreateUserInfo) (*UserInfoResponse, error)
//CheckPassWord(context.Context, *PasswordCheckInfo) (*CheckResponse, error)
//GetUserDetail(context.Context, *UserBasicInfo) (*UserDetailInfo, error)
//GetUserVideoList(context.Context, *UserBasicInfo) (*VideoIdList, error)
//GetUserFollows(context.Context, *UserBasicInfo) (*UserFollowsInfo, error)
//GetUserFollowers(context.Context, *UserBasicInfo) (*UserFollowersInfo, error)
//UpdateUserInfo(context.Context, *UserDetailInfo) (*Empty, error)
//FavoriteAction(context.Context, *FavInfo) (*Empty, error)
//RelationAction(context.Context, *RelationActionInfo) (*emptypb.Empty, error)
//CommentAction(context.Context, *CommentInfo) (*emptypb.Empty, error)

func (u *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	//新建用户
	var user model.User
	result := global.DB.Where(&model.User{Name: req.Username, Password: req.Password}).First(&user)
	if result.RowsAffected == 1 {
		zap.S().Infof("用户[%s]已经存在", req.Username)
		return nil, status.Errorf(codes.AlreadyExists, "用户名已存在")
	}
	user.Name = req.Username
	//TODO 这里密码可以用摘要算法加密一下
	//密码加密
	options := &password.Options{16, 100, 32, sha512.New}
	salt, encodedPwd := password.Encode(req.Password, options)
	user.Password = fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)

	result = global.DB.Create(&user)
	if result.Error != nil {
		zap.S().Infof("创建用户失败")
		return nil, status.Errorf(codes.Internal, "创建用户失败")
	}

	return &proto.UserInfoResponse{UserId: int32(user.ID)}, nil
}

func (u *UserServer) CheckPassWord(ctx context.Context, req *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	options := &password.Options{16, 100, 32, sha512.New}
	passwordInfo := strings.Split(req.Password, "$")
	check := password.Verify(req.Password, passwordInfo[2], passwordInfo[3], options)
	return &proto.CheckResponse{Success: check}, nil
}

func (u *UserServer) BatchGetUserDetail(ctx context.Context, req *proto.UserBasicInfo) (*proto.UserDetailInfoList, error) {
	userList := make([]model.User, 0)
	result := global.DB.Where("id IN ?", req.UserId).Find(&userList)
	if result.Error != nil {
		zap.S().Info("BatchGetUserDetail err: ", result.Error.Error())
		return nil, result.Error
	}
	userDetailInfoList := make([]*proto.UserDetailInfo, 0)
	for _, user := range userList {
		userDetailInfo := &proto.UserDetailInfo{
			Id:            int32(user.ID),
			Name:          user.Name,
			Passward:      user.Password,
			FollowCount:   user.Follow_count,
			FollowerCount: user.Follower_count,
			Follows:       user.FollowList,
			Followers:     user.FollowerList,
			FavList:       user.FavList,
		}
		userDetailInfoList = append(userDetailInfoList, userDetailInfo)
	}
	return &proto.UserDetailInfoList{UserDetailInfoList: userDetailInfoList}, nil
}

func (u *UserServer) GetUserVideoList(ctx context.Context, req *proto.UserBasicInfo) (*proto.VideoIdList, error) {
	user := model.User{}
	if result := global.DB.First(&user, req.UserId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	return &proto.VideoIdList{VideoId: user.VideoList}, nil
}

func (u *UserServer) GetUserFollows(ctx context.Context, req *proto.UserBasicInfo) (*proto.UserFollowsInfo, error) {
	user := model.User{}
	if result := global.DB.First(&user, req.UserId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	return &proto.UserFollowsInfo{FollowId: user.FollowList}, nil
}

func (u *UserServer) GetUserFollowers(ctx context.Context, req *proto.UserBasicInfo) (*proto.UserFollowersInfo, error) {
	user := model.User{}
	if result := global.DB.First(&user, req.UserId); result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	return &proto.UserFollowersInfo{FollowerId: user.FollowerList}, nil
}

func (*UserServer) UpdateUserInfo(ctx context.Context, req *proto.UserDetailInfo) (*emptypb.Empty, error) {
	userInfo := &model.User{}
	result := global.DB.First(userInfo, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "notFund err: "+result.Error.Error())
	}
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	userInfo = &model.User{
		//TODO 这里密码可以用摘要算法加密一下  因为无要求写这个接口  所以没实现修改密码的操作
		Name:           req.Name,
		Password:       req.Passward,
		Follow_count:   req.FollowCount,
		Follower_count: req.FollowerCount,
		FollowerList:   req.Follows,
		FollowList:     req.Followers,
		VideoList:      req.Videos,
		FavList:        req.Follows,
	}
	userInfo.ID = uint(req.Id)
	global.DB.Updates(userInfo)
	return &emptypb.Empty{}, nil
}

func (*UserServer) FavoriteAction(ctx context.Context, req *proto.FavInfo) (*emptypb.Empty, error) {
	//TODO 开启事务  修改user里面的fav_list  video表里面的favorite_count加一  提交事务
	user := model.User{}

	if req.ActionType == 1 {
		//点赞操作
		tx := global.DB.Begin()
		//修改user里面的fav_list
		if result := tx.First(&user, req.UserId); result.Error != nil {
			zap.S().Error(result.Error.Error())
			tx.Rollback()
			return nil, status.Error(codes.NotFound, result.Error.Error())
		}
		user.FavList = append(user.FavList, req.VideoId)
		if result := tx.Model(&model.User{}).Select("fav_list").Where("id = ?", req.UserId).Updates(&user); result.Error != nil {
			zap.S().Error(result.Error.Error())
			tx.Rollback()
			return nil, status.Error(codes.Internal, result.Error.Error())
		}

		//video表里面的favorite_count加一
		videoInfoList, err := global.FeedSrvCli.GetVideoList(context.Background(), &proto.VideoBasicInfo{VideoId: []int32{req.VideoId}})
		if err != nil {
			zap.S().Error(err)
			tx.Rollback()
			return nil, status.Error(codes.Internal, "调用feed GetVideoInfo err: "+err.Error())
		}
		videoInfo := videoInfoList.VideoInfos[0]
		videoInfo.FavoriteCount++
		_, err = global.FeedSrvCli.UpdateVideoInfo(context.Background(), &proto.VideoInfo{
			Id:            videoInfo.Id,
			UserId:        videoInfo.UserId,
			PlayUrl:       videoInfo.PlayUrl,
			CoverUrl:      videoInfo.CoverUrl,
			Title:         videoInfo.Title,
			FavoriteCount: videoInfo.FavoriteCount,
			CommentCount:  videoInfo.CommentCount,
			CommentList:   videoInfo.CommentList,
		})
		if err != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, "调用feed UpdateVideoInfo err: "+err.Error())
		}
		tx.Commit()
	} else {
		//取消点赞操作
		tx := global.DB.Begin()
		//修改user里面的fav_list
		if result := tx.First(&user, req.UserId); result.Error != nil {
			tx.Rollback()
			return nil, status.Error(codes.NotFound, result.Error.Error())
		}
		user.FavList = append(user.FavList, req.VideoId)
		for i, id := range user.FavList {
			if id == req.VideoId {
				user.FavList = append(user.FavList[:i], user.FavList[i+1:]...)
			}
		}
		if result := tx.Model(&model.User{}).Select("fav_list").Where("id = ?", req.UserId).Updates(&user); result.Error != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, result.Error.Error())
		}

		//video表里面的favorite_count减1
		videoList, err := global.FeedSrvCli.GetVideoList(context.Background(), &proto.VideoBasicInfo{VideoId: []int32{req.VideoId}})
		if err != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, "调用feed GetVideoInfo err: "+err.Error())
		}
		videoInfo := videoList.VideoInfos[0]
		videoInfo.FavoriteCount--
		_, err = global.FeedSrvCli.UpdateVideoInfo(context.Background(), &proto.VideoInfo{
			Id:            videoInfo.Id,
			UserId:        videoInfo.UserId,
			PlayUrl:       videoInfo.PlayUrl,
			CoverUrl:      videoInfo.CoverUrl,
			Title:         videoInfo.Title,
			FavoriteCount: videoInfo.FavoriteCount,
			CommentCount:  videoInfo.CommentCount,
			CommentList:   videoInfo.CommentList,
		})
		if err != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, "调用feed UpdateVideoInfo err: "+err.Error())
		}
		tx.Commit()
	}

	return &emptypb.Empty{}, nil
}

func (*UserServer) RelationAction(ctx context.Context, req *proto.RelationActionInfo) (*emptypb.Empty, error) {
	//TODO 开启事务 修改user表的follow_count follower_count follow_list follower_list  提交事务
	//user 为发起关注操作的用户  toUser为背user关注的用户
	user, toUser := model.User{}, model.User{}
	if req.ActionType == 1 {
		//关注
		result := global.DB.First(&user, req.UserId)
		if result.Error != nil {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return nil, status.Error(codes.NotFound, "没找到这个user")
		}

		result = global.DB.First(&toUser, req.ToUserId)
		if result.Error != nil {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return nil, status.Error(codes.NotFound, "没找到这个user")
		}
		user.FollowList = append(user.FollowList, int32(toUser.ID))
		user.Follow_count++
		toUser.FollowerList = append(toUser.FollowerList, int32(user.ID))
		toUser.Follower_count++
		tx := global.DB.Begin()
		if result := tx.Save(&user); result.Error != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		if result := tx.Save(&toUser); result.Error != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		tx.Commit()
	} else {
		//取消关注
		result := global.DB.First(&user, req.UserId)
		if result.Error != nil {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return nil, status.Error(codes.NotFound, "没找到这个user")
		}

		result = global.DB.First(&toUser, req.ToUserId)
		if result.Error != nil {
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return nil, status.Error(codes.NotFound, "没找到这个user")
		}
		for i, id := range user.FollowList {
			if id == int32(toUser.ID) {
				user.FollowList = append(user.FollowList[:i], user.FollowList[i+1:]...)
				break
			}
		}
		user.Follow_count--
		for i, id := range toUser.FollowerList {
			if id == int32(user.ID) {
				toUser.FollowerList = append(toUser.FollowerList[:i], toUser.FollowerList[i+1:]...)
			}
		}
		toUser.Follower_count--
		tx := global.DB.Begin()
		if result := tx.Save(&user); result.Error != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		if result := tx.Save(&toUser); result.Error != nil {
			tx.Rollback()
			return nil, status.Error(codes.Internal, result.Error.Error())
		}
		tx.Commit()
	}
	return &emptypb.Empty{}, nil
}
