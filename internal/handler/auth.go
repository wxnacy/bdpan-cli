package handler

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/wxnacy/bdpan-cli/internal/config"
	"github.com/wxnacy/bdpan-cli/internal/dto"
	"github.com/wxnacy/bdpan-cli/internal/logger"
	"github.com/wxnacy/bdpan-cli/internal/model"
	"github.com/wxnacy/bdpan-cli/internal/qrcode"
	"github.com/wxnacy/go-bdpan"
	"github.com/wxnacy/go-tools"
)

var authHandler *AuthHandler

// ErrUserCanceled 用户取消操作的哨兵错误
//
// 用途：
// - 当用户在二维码登录界面按 q、Ctrl+C 或 Esc 键主动取消登录时返回此错误
// - 使用哨兵错误模式，可以通过 errors.Is() 精确识别错误类型
//
// 处理流程：
// 1. 用户按键 → UI 显示红色"登录已取消"提示
// 2. qrcode.ShowQRCodeWithCallback 向 cancelChan 发送通知
// 3. pollDeviceTokenWithQRCode 接收通知并返回此错误
// 4. CmdLogin 检测到此错误后返回 nil，静默退出（避免重复显示错误）
//
// 设计原因：
// - UI 已经显示了友好的取消提示，无需在外层再次打印 "Error: user_canceled"
// - 使用哨兵错误可以区分用户主动取消和其他真正的错误
var ErrUserCanceled = errors.New("user_canceled")

// ErrLoginTimeout 登录超时的哨兵错误
//
// 用途：
// - 当登录流程超时时返回此错误，有两种超时情况：
//   1. UI 倒计时结束（50秒无人扫码）
//   2. 轮询次数耗尽（10次轮询×5秒=50秒都失败）
//
// 处理流程：
// 1. 超时发生 → UI 显示橙色"登录超时，请重试"提示
// 2. qrcode.ShowQRCodeWithCallback 向 timeoutChan 发送通知（情况1）
//    或 pollDeviceTokenWithQRCode 的 errChan 接收到此错误（情况2）
// 3. pollDeviceTokenWithQRCode 返回此错误
// 4. CmdLogin 检测到此错误后返回 nil，静默退出（避免重复显示错误）
//
// 设计原因：
// - UI 已经显示了友好的超时提示，无需在外层再次打印 "Error: 登录超时，请重试"
// - 两种超时情况统一使用同一个错误，简化上层处理逻辑
// - 超时时间（50秒）是轮询间隔（5秒）× 最大次数（10）的结果，与 UI 显示时间保持一致
var ErrLoginTimeout = errors.New("login_timeout")

func GetAuthHandler() *AuthHandler {
	if authHandler == nil {
		authHandler = &AuthHandler{
			accessToken: config.GetAccessToken(),
		}
		c, err := config.GetCredential()
		if err != nil {
			logger.Errorf("获取用户信息失败: %v", err)
		} else {
			authHandler.appID = c.AppID
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
	c, err := config.GetCredential()
	if err != nil || c.IsNil() {
		c, err = getCredentialByInut()
		if c.IsNil() {
			return fmt.Errorf("请正确输入 Credential 信息")
		}
		// 保存配置
		err = config.SaveCredential(*c)
		if err != nil {
			return err
		}
		// config.ReInitConfig()
		// if err != nil {
		// return err
		// }
	}

	access, err := config.GetAccess()
	if err != nil || access.IsExpired() {
		err := h.loginByCredential()
		if err != nil {
			// 用户取消登录或超时，静默退出
			if errors.Is(err, ErrUserCanceled) || errors.Is(err, ErrLoginTimeout) {
				return nil
			}
			return err
		}
	} else {
		h.accessToken = access.AccessToken
	}

	// 显示登录信息
	userInfo, err := h.GetUser()
	if err != nil {
		return err
	}
	fmt.Printf("你好, %s(%s)\n", userInfo.BaiduName, userInfo.GetVipName())

	// 显示配额
	quota, err := h.GetPan()
	if err != nil {
		return err
	}
	fmt.Printf("会员是否到期 %v\n", quota.Expire)
	fmt.Printf("网盘容量 %s/%s\n", tools.FormatSize(quota.Used), tools.FormatSize(quota.Total))
	return nil
}

func (h *AuthHandler) loginByCredential() error {
	c, err := config.GetCredential()
	if err != nil {
		return errors.New("Credential 不存在")
	}
	appKey := c.AppKey
	secretKey := c.SecretKey
	scope := config.Get().App.Scope
	deviceCode, err := bdpan.GetDeviceCode(appKey, scope)
	if err != nil {
		return errors.New("AppKey 不正确")
	}

	// 轮询验证并保存token
	return h.pollDeviceTokenWithQRCode(appKey, secretKey, deviceCode)
}

// pollDeviceTokenWithQRCode 轮询设备 token 并展示二维码，处理登录流程
//
// 参数：
// - appKey: 应用 Key，用于轮询 token
// - secretKey: 应用 Secret Key，用于轮询 token
// - deviceCode: 设备代码信息，包含二维码 URL 和 device code
//
// 返回：
// - error: 登录失败、用户取消或超时时返回相应的哨兵错误，nil 表示登录成功
//   - ErrUserCanceled: 用户按 q/Ctrl+C 主动取消
//   - ErrLoginTimeout: 登录超时（UI 倒计时结束或轮询次数耗尽）
//   - 其他错误: 如下载二维码失败等
//
// 实现逻辑：
//
// 1. 配置参数
//    - pollInterval = 5秒: 每次轮询的间隔时间
//    - pollMaxRetries = 10: 最多轮询10次，总计50秒
//    - qrDisplayTime = 50秒: 二维码显示时间，与轮询总时间一致
//
// 2. 创建通知 channel
//    - successChan: 登录成功后通知 UI 立即退出
//    - cancelChan: 接收 UI 的用户取消通知
//    - timeoutChan: 接收 UI 的倒计时超时通知
//    - tokenChan: 接收轮询 goroutine 的 token 结果
//    - errChan: 接收轮询 goroutine 的超时错误
//
// 3. 后台轮询 token（goroutine）
//    - 每隔 5 秒调用 bdpan.GetDeviceToken
//    - 成功：向 tokenChan 发送 token，向 successChan 通知 UI，结束 goroutine
//    - 失败：继续轮询，直到10次全部失败后向 errChan 发送超时错误
//
// 4. 显示二维码 UI（goroutine）
//    - 调用 qrcode.ShowQRCodeWithCallback 展示二维码
//    - 传入 successChan、cancelChan、timeoutChan 实现双向通信
//    - 如果显示失败（如下载错误），向 uiErrChan 发送错误
//
// 5. 等待并处理结果（select 语句）
//    - tokenChan: 登录成功，调用 saveDeviceToken 保存 token
//    - cancelChan: 用户取消，返回 ErrUserCanceled（UI 已显示"登录已取消"）
//    - timeoutChan: UI 倒计时结束，返回 ErrLoginTimeout（UI 已显示"登录超时"）
//    - uiErrChan: UI 显示错误（如下载图片失败），直接返回错误
//    - errChan: 轮询超时（10次全部失败），返回 ErrLoginTimeout
//
// 设计要点：
//
// 1. 超时机制统一
//    - 轮询总时间 = UI 显示时间 = 50 秒
//    - 两种超时都返回 ErrLoginTimeout，由上层统一处理
//
// 2. 错误提示去重
//    - 用户取消和超时的提示已在 UI 中显示（红色/橙色）
//    - 返回哨兵错误后，CmdLogin 会静默处理，不会二次显示错误
//
// 3. 并发控制
//    - 轮询 goroutine 和 UI goroutine 并行运行
//    - 通过 channel 实现通信和同步
//    - select 语句等待任一 channel 有数据即立即返回
//
// 4. 用户体验优化
//    - 登录成功后 UI 立即显示成功提示并退出
//    - 用户可以随时按 q/Ctrl+C 取消，响应迅速
//    - 倒计时显示让用户知道剩余时间
func (h *AuthHandler) pollDeviceTokenWithQRCode(appKey, secretKey string, deviceCode *bdpan.GetDeviceCodeRes) error {
	// 配置参数
	const (
		pollInterval   = 5 * time.Second  // 轮询间隔
		pollMaxRetries = 10               // 最大轮询次数
		qrDisplayTime  = 50 * time.Second // 二维码显示时间（同时也是总超时时间）
	)

	// 创建通知channel
	successChan := make(chan bool, 1)
	cancelChan := make(chan bool, 1)
	timeoutChan := make(chan bool, 1)
	tokenChan := make(chan *bdpan.GetDeviceTokenRes, 1)
	errChan := make(chan error, 1)

	// 后台轮询token
	go func() {
		for range pollMaxRetries {
			deviceToken, err := bdpan.GetDeviceToken(appKey, secretKey, deviceCode.DeviceCode)
			if err == nil {
				tokenChan <- deviceToken
				successChan <- true // 通知UI登录成功
				return
			}
			time.Sleep(pollInterval)
		}
		errChan <- ErrLoginTimeout
	}()

	// 显示二维码（带成功、取消和超时通知）
	uiErrChan := make(chan error, 1)
	go func() {
		if err := qrcode.ShowQRCodeWithCallback(deviceCode.QrcodeURL, qrDisplayTime, successChan, cancelChan, timeoutChan); err != nil {
			uiErrChan <- err
		}
	}()

	// 等待token、用户取消、UI超时或错误
	select {
	case deviceToken := <-tokenChan:
		// 保存token
		return h.saveDeviceToken(deviceToken)
	case <-cancelChan:
		// 用户取消登录（已在UI显示"登录已取消"）
		return ErrUserCanceled
	case <-timeoutChan:
		// UI倒计时结束（已在UI显示"登录超时，请重试"）
		return ErrLoginTimeout
	case err := <-uiErrChan:
		// UI显示错误（如下载二维码失败）
		return err
	case err := <-errChan:
		// 轮询超时（10次轮询都失败）
		return err
	}
}

// saveDeviceToken 保存设备 token 到本地配置文件
//
// 参数：
// - deviceToken: 从百度服务器获取的 token 信息
//
// 返回：
// - error: 总是返回 nil（config.SaveAccess 不返回错误）
//
// 实现步骤：
//
// 1. 构建 config.Access 结构
//    - AccessToken: 访问令牌，用于 API 调用
//    - ExpiresIn: 过期时间（秒）
//    - RefreshToken: 刷新令牌，用于延长会话
//    - RefreshTimestamp: 计算的刷新时间戳（当前时间 + 过期时间）
//
// 2. 更新 AuthHandler 内存状态
//    - 设置 h.accessToken，供后续 API 调用使用
//
// 3. 持久化到配置文件
//    - 调用 config.SaveAccess 保存到 ~/.bdpan/access.json
//    - 下次启动时可以直接加载，无需重新登录
//
// 设计说明：
// - 同时更新内存和磁盘，保证数据一致性
// - RefreshTimestamp 用于判断 token 是否过期，下次登录时检查
func (h *AuthHandler) saveDeviceToken(deviceToken *bdpan.GetDeviceTokenRes) error {
	var access config.Access
	access.AccessToken = deviceToken.AccessToken
	access.ExpiresIn = int(deviceToken.ExpiresIn)
	access.RefreshToken = deviceToken.RefreshToken
	access.RefreshTimestamp = int(time.Now().Unix()) + access.ExpiresIn
	// auth 赋值
	h.accessToken = deviceToken.AccessToken
	// 保存配置
	config.SaveAccess(access)
	return nil
}

// 从输入中创建用户信息
// 功能需求：
// - 使用 huh.Form 完成 Credential 中 AppID/AppKey/SecretKey 信息的收集
func getCredentialByInut() (*config.Credential, error) {
	item := &config.Credential{}

	appId := ""

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("AppId").
				Value(&appId).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("AppID 不能为空")
					}
					if _, err := strconv.Atoi(s); err != nil {
						return errors.New("AppID 必须为数字")
					}
					return nil
				}),
			huh.NewInput().
				Title("AppKey").
				Value(&item.AppKey).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("AppKey 不能为空")
					}
					return nil
				}),
			huh.NewInput().
				Title("SecretKey").
				Value(&item.SecretKey).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("SecretKey 不能为空")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	id, err := strconv.Atoi(appId)
	if err != nil {
		return nil, fmt.Errorf("AppID: %s 输入错误", appId)
	}
	item.AppID = id
	return item, nil
}
