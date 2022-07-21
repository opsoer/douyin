### douyin

- 简介：采用分布式架构，分为用户服务、Feed服务和统一对外的Web服务三个服务；用户服务实现了视频投稿，个人信息，点赞列表，关注列表，粉丝列表等接口；Feed服务实现了视频Feed流，获取用户评论等接口。

- 技术栈：Go语言，MySQL，Gorm，Gin，gRPC，Viper，Zap。



#### feed-srv proto接口定义

```protobuf
syntax = "proto3";
option go_package = ".;proto";
import "google/protobuf/empty.proto";
service Feed {
  rpc GetVideos (LatestTime) returns (VideoInfoList);
  rpc GetVideoComIdList (VideoBasicInfo) returns (CommentList);
  rpc BatchGetComment(CommentList) returns (ContentList);
  rpc CreateVideo (CreateVideoInfo) returns (VideoBasicInfo);
  rpc GetVideoList(VideoBasicInfo) returns (VideoInfoList);
  rpc UpdateVideoInfo(VideoInfo) returns (google.protobuf.Empty);
  rpc CommentAction (commentInfo)returns(google.protobuf.Empty);
}
message commentInfo {
  int32 userId = 1;
  int32 videoId = 2;
  int32  comment_id = 3; //删除评论要用
  int32 actionType = 4; // 1为发布评论  2为删除评论
  string commentText = 5;
}
message CreateVideoInfo {
  int32 userId = 1;
  string playUrl = 2;
  string coverUrl = 3;
  string title = 4;
}
message Content {
  int32 userId = 1;
  string str = 2;
}
message ContentList {
  repeated Content contentList = 1;
}
message LatestTime {
  int64 time = 1;
}
message VideoInfo {
  int32 id = 1;
  int32 userId = 2;
  string playUrl = 3;
  string coverUrl = 4;
  string title = 5;
  int32 favorite_count = 6;
  int32 comment_count = 7;
  repeated int32 commentList = 8;
}
message VideoInfoList {
  repeated VideoInfo videoInfos = 1;
}
message NextTime {
  uint64 time = 2;
}
message VideoBasicInfo {
  repeated int32 videoId = 1;
}
message CommentList {
  repeated int32 ComIdList = 1;
}
```



### user-srv proto接口定义

```protobuf
syntax = "proto3";
option go_package = ".;proto";
import "google/protobuf/empty.proto";
//protoc -I . user.proto --go_out=plugins=grpc:.
service User{
  rpc CreateUser(CreateUserInfo) returns (UserInfoResponse);
  rpc CheckPassWord(PasswordCheckInfo) returns (CheckResponse);
  rpc BatchGetUserDetail(UserBasicInfo) returns (UserDetailInfoList);
  rpc GetUserVideoList(UserBasicInfo) returns(VideoIdList);
  rpc GetUserFollows (UserBasicInfo) returns (UserFollowsInfo);
  rpc GetUserFollowers (UserBasicInfo) returns (UserFollowersInfo);
  rpc UpdateUserInfo (UserDetailInfo) returns (google.protobuf.Empty);
  rpc FavoriteAction (favInfo)returns(google.protobuf.Empty);
  rpc RelationAction(RelationActionInfo)returns(google.protobuf.Empty);
}

message RelationActionInfo {
  int32 userId = 1;
  int32 toUserId = 2;
  int32 actionType = 3;
}
message favInfo {
  int32 userId = 1;
  int32 videoId = 2;
  int32 actionType = 3;
}
message CreateUserInfo {
  string  username = 1;
  string password = 2;
}

message UserInfoResponse {
  int32 user_id = 3;
}

message PasswordCheckInfo {
  string  username = 1;
  string password = 2;
}

message CheckResponse {
  int32 id = 1;
  bool success = 2;
}

message UserBasicInfo {
  repeated int32 user_id = 1;
}

message UserDetailInfo {
  int32 id = 1;
  string  name = 2;
  string passward = 3;
  int32  follow_count = 4;
  int32 follower_count = 5;
  repeated int32 follows = 6;
  repeated int32 followers = 7;
  repeated int32 Videos = 8;
  repeated int32 comments = 9;
  repeated int32 FavList = 10;
}
message UserDetailInfoList {
  repeated UserDetailInfo userDetailInfoList = 1;
}
message VideoIdList {
  repeated int32 videoId = 1;
}

message UserFollowsInfo {
  repeated int32 FollowId = 1;
}

message UserFollowersInfo {
  repeated int32 FollowerId = 1;
}

```
