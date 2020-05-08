package user

import (
	"encoding/json"
	"fmt"
	"net/url"
	"errors"

	"github.com/silenceper/wechat/context"
	"github.com/silenceper/wechat/util"
)

const (
	userInfoURL     = "https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN"
	updateRemarkURL = "https://api.weixin.qq.com/cgi-bin/user/info/updateremark?access_token=%s"
	userListURL     = "https://api.weixin.qq.com/cgi-bin/user/get"
	batchGetUserInfoURL = "https://api.weixin.qq.com/cgi-bin/user/info/batchget?access_token=%s"
)

//User 用户管理
type User struct {
	*context.Context
}

//NewUser 实例化
func NewUser(context *context.Context) *User {
	user := new(User)
	user.Context = context
	return user
}

//Info 用户基本信息
type Info struct {
	util.CommonError

	Subscribe      int32   `json:"subscribe"`
	OpenID         string  `json:"openid"`
	Nickname       string  `json:"nickname"`
	Sex            int32   `json:"sex"`
	City           string  `json:"city"`
	Country        string  `json:"country"`
	Province       string  `json:"province"`
	Language       string  `json:"language"`
	Headimgurl     string  `json:"headimgurl"`
	SubscribeTime  int32   `json:"subscribe_time"`
	UnionID        string  `json:"unionid"`
	Remark         string  `json:"remark"`
	GroupID        int32   `json:"groupid"`
	TagidList      []int32 `json:"tagid_list"`
	SubscribeScene string  `json:"subscribe_scene"`
	QrScene        int     `json:"qr_scene"`
	QrSceneStr     string  `json:"qr_scene_str"`
}

// OpenidList 用户列表
type OpenidList struct {
	Total int `json:"total"`
	Count int `json:"count"`
	Data  struct {
		OpenIDs []string `json:"openid"`
	} `json:"data"`
	NextOpenID string `json:"next_openid"`
}

//GetUserInfo 获取用户基本信息
func (user *User) GetUserInfo(openID string) (userInfo *Info, err error) {
	var accessToken string
	accessToken, err = user.GetAccessToken()
	if err != nil {
		return
	}

	uri := fmt.Sprintf(userInfoURL, accessToken, openID)
	var response []byte
	response, err = util.HTTPGet(uri)
	if err != nil {
		return
	}
	userInfo = new(Info)
	err = json.Unmarshal(response, userInfo)
	if err != nil {
		return
	}
	if userInfo.ErrCode != 0 {
		err = fmt.Errorf("GetUserInfo Error , errcode=%d , errmsg=%s", userInfo.ErrCode, userInfo.ErrMsg)
		return
	}
	return
}

// UpdateRemark 设置用户备注名
func (user *User) UpdateRemark(openID, remark string) (err error) {
	var accessToken string
	accessToken, err = user.GetAccessToken()
	if err != nil {
		return
	}

	uri := fmt.Sprintf(updateRemarkURL, accessToken)
	var response []byte
	response, err = util.PostJSON(uri, map[string]string{"openid": openID, "remark": remark})
	if err != nil {
		return
	}

	return util.DecodeWithCommonError(response, "UpdateRemark")
}

// ListUserOpenIDs 返回用户列表
func (user *User) ListUserOpenIDs(nextOpenid ...string) (*OpenidList, error) {
	accessToken, err := user.GetAccessToken()
	if err != nil {
		return nil, err
	}

	uri, _ := url.Parse(userListURL)
	q := uri.Query()
	q.Set("access_token", accessToken)
	if len(nextOpenid) > 0 && nextOpenid[0] != "" {
		q.Set("next_openid", nextOpenid[0])
	}
	uri.RawQuery = q.Encode()

	response, err := util.HTTPGet(uri.String())
	if err != nil {
		return nil, err
	}

	userlist := new(OpenidList)
	err = json.Unmarshal(response, userlist)
	if err != nil {
		return nil, err
	}

	return userlist, nil
}

// ListAllUserOpenIDs 返回所有用户OpenID列表
func (user *User) ListAllUserOpenIDs() ([]string, error) {
	nextOpenid := ""
	openids := []string{}
	count := 0
	for {
		ul, err := user.ListUserOpenIDs(nextOpenid)
		if err != nil {
			return nil, err
		}
		openids = append(openids, ul.Data.OpenIDs...)
		count += ul.Count
		if ul.Total > count {
			nextOpenid = ul.NextOpenID
		} else {
			return openids, nil
		}
	}
}

// BatchUserQuery 待查询的用户列表
type BatchUserQuery struct {
	OpenID 		string		`json:"openid"`
	Lang 		string		`json:"lang"`
}

// BatchGetUser 批量获取用户基本信息
func (user *User) BatchGetUser(batchUserQuery ... *BatchUserQuery)([]*Info, error){
	if len(batchUserQuery)>100{
		return nil, errors.New("最多支持一次拉取100条")
	}
	var accessToken string
	accessToken, err := user.GetAccessToken()
	if err != nil {
		return nil, err
	}

	requestMap := make(map[string]interface{})
	requestMap["user_list"] = batchUserQuery
	uri := fmt.Sprintf(batchGetUserInfoURL, accessToken)
	response, err := util.PostJSON(uri, requestMap)
	if err != nil {
		return nil, err
	}
	//	batchUserQueryResponse 批量查询返回值
	type batchUserQueryResponse struct {
		List    []*Info		`json:"user_info_list"`
	}
	userList := &batchUserQueryResponse{}
	err = json.Unmarshal(response, userList)
	if err != nil {
		return nil, err
	}
	return userList.List, nil
}
