package api

type baseRsp struct {
	statusCode int32  `json:"status_code"`
	statusMsg  string `json:"status_msg"`
}
type userRsp struct {
	id            int32  `json:"id"`
	name          string `json:"name"`
	followCount   int32  `json:"follow_count"`
	followerCount int32  `json:"follower_count"`
	isFollow      bool   `json:"is_follow"`
}

type videoRsp struct {
	id            int32  `json:"id"`
	playUrl       string `json:"play_url"`
	coverUrl      string `json:"cover_url"`
	favoriteCount int32  `json:"favorite_count"`
	commentCount  int32  `json:"comment_count"`
	isFavorite    bool   `json:"is_favorite"`
	title         string `json:"title"`
}

type feedVideoRsp struct {
	baseRsp
	nextTime  int64 `json:"next_time"`
	videoList []videoRsp `json:"video_list"`
}
