package handler

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/internal/qrcode"
	"github.com/wxnacy/bdpan/auth"
	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
)

var authHandler *AuthHandler

func GetAuthHandler() *AuthHandler {
	if authHandler == nil {
		authHandler = &AuthHandler{
			accessToken: config.GetAccessToken(),
			appID:       config.Get().Credential.AppID,
		}
	}
	return authHandler
}

type AuthHandler struct {
	accessToken string
	appID       int
}

func (h *AuthHandler) GetUser() (*model.User, error) {

	info, err := bdpan.GetUserInfo(h.accessToken)
	if err != nil {
		return nil, err
	}
	return model.NewUser(info), nil
}

// func (h *AuthHandler) RefreshUserInfo() (*model.User, error) {
// return bdpan.GetUserInfo(h.accessToken)
// }

func (h *AuthHandler) GetPan() (*model.Pan, error) {
	info, err := bdpan.GetPanInfo(h.accessToken)
	if err != nil {
		return nil, err
	}
	return model.NewPan(h.appID, info), nil
}

func (h *AuthHandler) NewPan(panInfo *bdpan.GetPanInfoRes) *model.Pan {
	return model.NewPan(h.appID, panInfo)
}

func (h *AuthHandler) GetPanFromDB() *model.Pan {
	return model.FindFirstByID[model.Pan](h.appID)
}

func (h *AuthHandler) RefreshPan() (*model.Pan, error) {
	pan, err := h.GetPan()
	if err != nil {
		return nil, err
	}
	model.Save(pan)

	return pan, nil
}

func (h *AuthHandler) CmdLogin(req *dto.LoginReq) error {
	conf := config.Get()
	c := conf.Credential
	if c.IsNil() {
		c = *getCredentialByInut()
		if c.IsNil() {
			return errors.New(fmt.Sprintf("请正确输入 Credential 信息"))
		}
		// 保存配置
		config.SaveCredential(c)
	}

	access := conf.Access
	if access.IsExpired() {
		err := h.loginByCredential()
		if err != nil {
			return err
		}
	}

	// 显示登录信息
	userInfo, err := auth.GetUserInfo(h.accessToken)
	if err != nil {
		return err
	}
	fmt.Printf("你好, %s(%s)\n", userInfo.BaiduName, userInfo.GetVipName())

	// 显示配额
	quota, err := auth.GetQuota(h.accessToken)
	if err != nil {
		return err
	}
	fmt.Printf("会员是否到期 %v\n", quota.Expire)
	fmt.Printf("网盘容量 %s/%s\n", tools.FormatSize(quota.Used), tools.FormatSize(quota.Total))
	return nil
}

func (h *AuthHandler) loginByCredential() error {
	appKey := config.Get().Credential.AppKey
	secretKey := config.Get().Credential.SecretKey
	scope := config.Get().App.Scope
	deviceCode, err := auth.GetDeviceCode(appKey, scope)
	if err != nil {
		return errors.New("AppKey 不正确")
	}

	for i := 0; i < int(10); i++ {
		err := qrcode.ShowByUrl(deviceCode.QrcodeURL, 5*time.Second)
		if err != nil {
			return err
		}
		deviceToken, err := auth.GetDeviceToken(appKey, secretKey, deviceCode.DeviceCode)
		if err == nil {
			var access config.Access
			access.AccessToken = deviceToken.AccessToken
			access.ExpiresIn = int(deviceToken.ExpiresIn)
			access.RefreshToken = deviceToken.RefreshToken
			access.RefreshTimestamp = int(time.Now().Unix()) + access.ExpiresIn
			// auth 赋值
			// h.auth.SetAccess(config.ToAuthAccess())
			h.accessToken = deviceToken.AccessToken
			// 保存配置
			config.SaveAccess(access)
			return nil
		}
	}
	return errors.New("登录超时")
}

// 从输入中创建用户信息
func getCredentialByInut() *config.Credential {

	item := &config.Credential{}

	appId := ""
	huh.NewInput().
		Title("AppId").
		Value(&appId).
		Run()

	huh.NewInput().
		Title("AppKey").
		Value(&item.AppKey).
		Run()

	huh.NewInput().
		Title("SecretKey").
		Value(&item.SecretKey).
		Run()

	id, err := strconv.Atoi(appId)
	if err != nil {
		panic("AppID 输入错误")
	}
	item.AppID = id
	return item
}
