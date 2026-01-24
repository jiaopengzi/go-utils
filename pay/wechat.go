//
// FilePath    : go-utils\pay\wechat.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 微信支付
//

package pay

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/jiaopengzi/go-utils"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
	wechatUtils "github.com/wechatpay-apiv3/wechatpay-go/utils"

	"go.uber.org/zap"
)

// 微信支付状态常量
const (
	// 文档：https://pay.weixin.qq.com/doc/v3/merchant/4012791882
	TradeStateWechatPaySuccess    = "SUCCESS"    // 支付成功
	TradeStateWechatPayRefund     = "REFUND"     // 转入退款
	TradeStateWechatPayNotPay     = "NOTPAY"     // 未支付
	TradeStateWechatPayClosed     = "CLOSED"     // 已关闭
	TradeStateWechatPayRevoked    = "REVOKED"    // 已撤销（仅付款码支付会返回）
	TradeStateWechatPayUserPaying = "USERPAYING" // 用户支付中（仅付款码支付会返回）
	TradeStateWechatPayPayError   = "PAYERROR"   // 支付失败（仅付款码支付会返回）
)

// WeChatPayConfig 微信支付配置
type WeChatPayConfig struct {
	Enabled                    bool   `mapstructure:"enabled" json:"enabled"`                                                                                                     // 是否启用微信支付
	MchID                      string `mapstructure:"mch_id" json:"mch_id" binding:"required_if=Enabled true" example:"1234567890"`                                               // 商户号
	MchCertificateSerialNumber string `mapstructure:"mch_certificate_serial_number" json:"mch_certificate_serial_number" binding:"required_if=Enabled true" example:"1234567890"` // 商户证书序列号
	MchPrivateKey              string `mapstructure:"mch_private_key" json:"mch_private_key" binding:"required_if=Enabled true" example:"key"`                                    // 商户私钥
	AppID                      string `mapstructure:"app_id" json:"app_id" binding:"required_if=Enabled true" example:"app1234567890"`                                            // 应用ID
	APIv3Key                   string `mapstructure:"api_v3_key" json:"api_v3_key" binding:"required_if=Enabled true" example:"key1234567890"`                                    // APIv3密钥
	NotifyHost                 string `mapstructure:"notify_host" json:"notify_host" binding:"required_if=Enabled true" example:"https://example.com:8080"`                       // 支付结果通知主机地址
	NotifyPath                 string `mapstructure:"notify_path" json:"notify_path" binding:"required_if=Enabled true" example:"/wechat/notify"`                                 // 支付结果通知路由
	RefundPath                 string `mapstructure:"refund_path" json:"refund_path" binding:"required_if=Enabled true" example:"/refund_notify"`                                 // 退款结果通知路由
}

type WeChatPay struct {
	Client      *core.Client // 微信支付客户端
	PrivateKey  *rsa.PrivateKey
	Conf        *WeChatPayConfig // 支付宝配置
	APIPath     string           // API 路径前缀 e.g. /api/v1
	PayBasePath string           // 支付基础路由 e.g. /pay
}

// NewWeChatPay 创建新的微信支付实例
func NewWeChatPay(conf *WeChatPayConfig, apiPath, payBasePath string) (*WeChatPay, error) {
	// 使用 utils 提供的函数从本地文件中加载商户私钥，商户私钥会用来生成请求的签名
	mchPrivateKey, err := wechatUtils.LoadPrivateKey(conf.MchPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("load WeChatPay private key error: %w", err)
	}

	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(
			conf.MchID,
			conf.MchCertificateSerialNumber,
			mchPrivateKey,
			conf.APIv3Key,
		),
	}

	// 创建 WeChatPay 客户端
	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("create WeChatPay client error: %w", err)
	}

	// 创建 WeChatPay 实例
	wechatPay := &WeChatPay{
		Client:      client,
		PrivateKey:  mchPrivateKey,
		Conf:        conf,
		APIPath:     apiPath,
		PayBasePath: payBasePath,
	}

	// 打印日志确认微信支付客户端创建成功
	zap.L().Debug("WeChatPay client created successfully", zap.String("mchID", conf.MchID))

	return wechatPay, nil
}

// Prepay 微信支付实现 二维码的URL, 使用二维码转码工具生成二维码图片, 手机扫码支付
func (w *WeChatPay) Prepay(orderID uint64, amount int64, description, returnURL string, timeExpire time.Time) (string, error) {
	// 文档: https://github.com/wechatpay-apiv3/wechatpay-go/tree/main
	// 支付结果通知地址
	notifyURL := fmt.Sprintf("%s/%s%s%s",
		w.Conf.NotifyHost,
		w.APIPath,
		w.PayBasePath,
		w.Conf.NotifyPath,
	)
	svc := native.NativeApiService{Client: w.Client}
	resp, _, err := svc.Prepay(context.Background(),
		native.PrepayRequest{
			Appid:       core.String(w.Conf.AppID),
			Mchid:       core.String(w.Conf.MchID),
			Description: core.String(description),                // 使用商品title作为描述
			OutTradeNo:  core.String(utils.Uint64ToStr(orderID)), // 商户订单号字符串规则最小长度为6
			NotifyUrl:   core.String(notifyURL),
			Amount: &native.Amount{
				Currency: core.String("CNY"), // CNY：人民币, 境内商户号仅支持人民币。
				Total:    core.Int64(amount), // 订单总金额, 单位为分
			},
			TimeExpire: core.Time(timeExpire), // 订单失效时间, 格式为 ISO 8601
		},
	)

	if err != nil {
		return "", fmt.Errorf("WeChatPay prepay error: %w", err)
	}

	return *resp.CodeUrl, nil
}

// GetNotifyPayment 微信支付实现应答支付结果通知接口, 包含验签和获取支付结果
func (w *WeChatPay) GetNotifyPayment(request *http.Request) (bool, *PaymentResult, error) {
	// 验签和解析
	transaction, err := validateParseNotifyRequest[payments.Transaction](w, request)
	if err != nil {
		// 如果验签未通过，或者解密失败
		return false, nil, fmt.Errorf("WeChatPay verify sign error: %w", err)
	}

	// 检查响应字段是否为 nil
	if transaction.OutTradeNo == nil ||
		transaction.TransactionId == nil ||
		transaction.TradeType == nil ||
		transaction.Amount == nil ||
		transaction.Amount.Total == nil ||
		transaction.Appid == nil ||
		transaction.Mchid == nil ||
		transaction.TradeState == nil {
		return false, nil, errors.New("transaction fields are nil")
	}

	// 检查交易状态是否为成功
	if *transaction.TradeState != TradeStateWechatPaySuccess {
		return false, nil, errors.New("trade state is not success : " + *transaction.TradeState)
	}

	result := &PaymentResult{
		PayType:       PayTypeWechat, // 微信支付类型
		OrderID:       utils.StrToUint64(*transaction.OutTradeNo),
		TotalAmount:   *transaction.Amount.Total, // 订单总金额，单位为分
		TransactionID: *transaction.TransactionId,
		TradeState:    TradeStatePaid,
		TradeType:     *transaction.TradeType,
		AppID:         *transaction.Appid,
		MchID:         *transaction.Mchid,
	}

	return true, result, nil
}

// ValidateNotifyPayment 微信支付实现验证支付结果通知接口
// 主要校验商户订单号、金额、商户号、appid 等信息是否匹配
//
//nolint:dupl
func (w *WeChatPay) ValidateNotifyPayment(payment *PaymentResult, orderID uint64, amount int64) (bool, *PaymentResult, error) {
	// 检查支付结果是否为 nil
	if payment == nil {
		return false, nil, errors.New("WeChatPay payment result is nil")
	}

	// 校验订单号
	if payment.OrderID != orderID {
		return false, nil, fmt.Errorf("order ID mismatch: expected %d, got %d", orderID, payment.OrderID)
	}

	// 校验金额
	if payment.TotalAmount != amount {
		return false, nil, fmt.Errorf("amount mismatch: expected %d, got %d", amount, payment.TotalAmount)
	}

	// 校验商户号
	if payment.MchID != w.Conf.MchID {
		return false, nil, fmt.Errorf("MchID mismatch: expected %s, got %s", w.Conf.MchID, payment.MchID)
	}

	// 校验 AppID
	if payment.AppID != w.Conf.AppID {
		return false, nil, fmt.Errorf("AppID mismatch: expected %s, got %s", w.Conf.AppID, payment.AppID)
	}

	// 如果所有校验都通过，返回 true 和支付结果
	return true, payment, nil
}

// QueryPayment 微信支付实现查询支付结果接口
func (w *WeChatPay) QueryPayment(orderID uint64) (*PaymentResult, error) {
	ctx := context.Background()
	svc := native.NativeApiService{Client: w.Client}

	resp, _, err := svc.QueryOrderByOutTradeNo(ctx,
		native.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(utils.Uint64ToStr(orderID)), // 商户订单号字符串规则最小长度为6
			Mchid:      core.String(w.Conf.MchID),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("WeChatPay query payment error: %w", err)
	}

	result := &PaymentResult{
		PayType: PayTypeWechat, // 微信支付类型
		OrderID: orderID,
	}

	state := TradeStateUnpaid // 初始状态未知

	// 状态对齐
	if resp.TradeState != nil {
		switch *resp.TradeState {
		case TradeStateWechatPaySuccess: // 支付成功
			state = TradeStatePaid
		case TradeStateWechatPayRefund: // 转入退款
			state = TradeStateRefunded
		case TradeStateWechatPayNotPay: // 未支付
			state = TradeStateUnpaid
		case TradeStateWechatPayClosed: // 已关闭
			state = TradeStateClosed
		default:
			return nil, fmt.Errorf("WeChatPay unknown trade state: %s", *resp.TradeState)
		}
	}

	// 检查字段是否为 nil
	if resp.Amount != nil && resp.Amount.Total != nil {
		result.TotalAmount = *resp.Amount.Total // 订单总金额，单位为分
	}

	// 检查字段是否为 nil
	if resp.TransactionId != nil {
		result.TransactionID = *resp.TransactionId // 交易号
	}

	if resp.TradeType != nil {
		result.TradeType = *resp.TradeType // 交易类型
	}

	result.TradeState = state // 设置支付状态

	return result, nil
}

// CloseOrder 微信支付实现关闭订单接口
func (w *WeChatPay) CloseOrder(orderID uint64) error {
	// 文档: https://github.com/wechatpay-apiv3/wechatpay-go/tree/main
	svc := native.NativeApiService{Client: w.Client}

	result, err := svc.CloseOrder(context.Background(),
		native.CloseOrderRequest{
			OutTradeNo: core.String(utils.Uint64ToStr(orderID)), // 商户订单号字符串规则最小长度为6
			Mchid:      core.String(w.Conf.MchID),
		},
	)

	if err != nil {
		return fmt.Errorf("WeChatPay cancel order error: %w", err)
	}

	// 检查响应状态码是否为 204 No Content
	// 文档: https://pay.weixin.qq.com/doc/v3/merchant/4012791881
	if result.Response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("WeChatPay cancel order failed: status code %d", result.Response.StatusCode)
	}

	return nil
}

// Refund 微信支付实现退款接口
func (w *WeChatPay) Refund(orderID, refundID uint64, amount, refundAmount int64, reason string) (*RefundResult, error) {
	// 退款结果通知地址
	refundURL := fmt.Sprintf("%s/%s%s%s",
		w.Conf.NotifyHost,
		w.APIPath,
		w.PayBasePath,
		w.Conf.RefundPath,
	)
	svc := refunddomestic.RefundsApiService{Client: w.Client}

	resp, apiResult, err := svc.Create(context.Background(),
		refunddomestic.CreateRequest{
			OutTradeNo:  core.String(utils.Uint64ToStr(orderID)),
			OutRefundNo: core.String(utils.Uint64ToStr(refundID)),
			Reason:      core.String(reason),
			NotifyUrl:   core.String(refundURL),
			Amount: &refunddomestic.AmountReq{
				Currency: core.String("CNY"),
				Refund:   core.Int64(refundAmount), // 退款金额，单位为分
				Total:    core.Int64(amount),       // 订单总金额，单位为分
			},
		},
	)

	if err != nil {
		// 如果是余额不足错误，则返回自定义错误
		if apiResult != nil && apiResult.Response != nil && apiResult.Response.StatusCode == http.StatusForbidden {
			matched, errR := regexp.MatchString(`NOT_ENOUGH`, err.Error())
			if errR != nil {
				return nil, fmt.Errorf("regexp match error: %w", errR)
			}

			// 如果匹配到余额不足错误，则返回自定义错误
			if matched {
				zap.L().Warn("WeChatPay refund error", zap.Error(err))
				return nil, utils.ErrRefundWeChatNotEnough
			}
		}

		return nil, fmt.Errorf("WeChatPay refund error: %w", err)
	}

	// 检查响应字段是否为 nil
	if err = checkRefundFields(resp); err != nil {
		return nil, fmt.Errorf("WeChatPay response fields are nil: %w", err)
	}

	// 对齐状态
	state, err := parseRefundStatus(*resp.Status)
	if err != nil {
		return nil, fmt.Errorf("WeChatPay parse refund status error: %w", err)
	}

	// 成功发起后状态为退款中, 等待退款通知再更新状态
	result := &RefundResult{
		PayType:             PayTypeWechat, // 微信支付类型
		RefundID:            refundID,
		OrderID:             orderID,
		TransactionID:       *resp.TransactionId,
		RefundTransactionID: *resp.RefundId,
		TotalAmount:         *resp.Amount.Total,  // 订单总金额，单位为分
		RefundAmount:        *resp.Amount.Refund, // 退款金额
		Reason:              reason,              // 退款原因
		Status:              state,               // 退款状态
	}

	return result, nil
}

// 根据文档定义退款通知的结构体，**这里很坑查询的结构体和通知结构体居然不一样**
// https://pay.weixin.qq.com/doc/v3/merchant/4012791886

// AmountRefundNotifyWechat 微信支付退款金额结构体
type AmountRefundNotifyWechat struct {
	Total       int64 `json:"total"`
	Refund      int64 `json:"refund"`
	PayerTotal  int64 `json:"payer_total"`
	PayerRefund int64 `json:"payer_refund"`
}

// RefundNotifyWechat 微信支付退款通知结构体
type RefundNotifyWechat struct {
	MchID               string                   `json:"mchid"`
	TransactionID       string                   `json:"transaction_id"`
	OutTradeNo          string                   `json:"out_trade_no"`
	RefundID            string                   `json:"refund_id"`
	OutRefundNo         string                   `json:"out_refund_no"`
	RefundStatus        refunddomestic.Status    `json:"refund_status"`
	SuccessTime         time.Time                `json:"success_time"`
	UserReceivedAccount string                   `json:"user_received_account"`
	Amount              AmountRefundNotifyWechat `json:"amount"`
}

// GetNotifyRefund 微信支付实现应答退款结果通知接口
func (w *WeChatPay) GetNotifyRefund(request *http.Request) (bool, *RefundResult, error) {
	// 验签和解析
	refund, err := validateParseNotifyRequest[RefundNotifyWechat](w, request)
	if err != nil {
		// 如果验签未通过，或者解密失败
		return false, nil, fmt.Errorf("WeChatPay verify sign error: %w", err)
	}

	// 对齐状态
	state, err := parseRefundStatus(refund.RefundStatus)
	if err != nil {
		return false, nil, fmt.Errorf("WeChatPay parse refund status error: %w", err)
	}

	result := &RefundResult{
		PayType:             PayTypeWechat,                         // 微信支付类型
		RefundID:            utils.StrToUint64(refund.OutRefundNo), // 商户退款单号
		OrderID:             utils.StrToUint64(refund.OutTradeNo),  // 商户订单号
		TransactionID:       refund.TransactionID,
		RefundTransactionID: refund.RefundID,
		TotalAmount:         refund.Amount.Total,  // 订单总金额，单位为分
		RefundAmount:        refund.Amount.Refund, // 退款金额
		Status:              state,                // 退款状态
	}

	return true, result, nil
}

// QueryRefund 微信支付实现查询退款结果接口
func (w *WeChatPay) QueryRefund(orderID, refundID uint64) (*RefundResult, error) {
	svc := refunddomestic.RefundsApiService{Client: w.Client}

	resp, _, err := svc.QueryByOutRefundNo(
		context.Background(),
		refunddomestic.QueryByOutRefundNoRequest{
			OutRefundNo: core.String(utils.Uint64ToStr(refundID)), // 商户退款单号
		},
	)
	if err != nil {
		return nil, fmt.Errorf("WeChatPay query refund error: %w", err)
	}

	// 检查响应字段是否为 nil
	if err = checkRefundFields(resp); err != nil {
		return nil, fmt.Errorf("WeChatPay response fields are nil: %w", err)
	}

	// 对齐状态
	state, err := parseRefundStatus(*resp.Status)
	if err != nil {
		return nil, fmt.Errorf("WeChatPay parse refund status error: %w", err)
	}

	result := &RefundResult{
		PayType:             PayTypeWechat,       // 微信支付类型
		RefundID:            refundID,            // 商户退款单号
		OrderID:             orderID,             // 商户订单号
		TransactionID:       *resp.TransactionId, // 微信支付订单
		RefundTransactionID: *resp.RefundId,      // 微信退款单号
		TotalAmount:         *resp.Amount.Total,  // 订单总金额，单位为分
		RefundAmount:        *resp.Amount.Refund, // 退款金额
		Status:              state,               // 退款状态
	}

	return result, nil
}

// validateParseNotifyRequest 验签和解析微信支付通知请求, 包含支付和退款通知
func validateParseNotifyRequest[T any](w *WeChatPay, request *http.Request) (*T, error) {
	// 文档: https://github.com/wechatpay-apiv3/wechatpay-go/tree/main
	ctx := context.Background()

	// 1. 使用 `RegisterDownloaderWithPrivateKey` 注册下载器
	err := downloader.MgrInstance().RegisterDownloaderWithPrivateKey(
		ctx,
		w.PrivateKey,
		w.Conf.MchCertificateSerialNumber,
		w.Conf.MchID,
		w.Conf.APIv3Key,
	)
	if err != nil {
		return nil, fmt.Errorf("WeChatPay register downloader error: %w", err)
	}

	// 2. 获取商户号对应的微信支付平台证书访问器
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(w.Conf.MchID)

	// 3. 使用证书访问器初始化 `notify.Handler`
	handler := notify.NewNotifyHandler(w.Conf.APIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))

	// 4. 验签和解析通知请求
	t := new(T)

	_, err = handler.ParseNotifyRequest(ctx, request, t)
	if err != nil {
		// 如果验签未通过，或者解密失败
		return nil, fmt.Errorf("WeChatPay verify sign error: %w", err)
	}

	return t, nil
}

// parseRefundStatus 根据微信退款返回的 Status 解析为系统内部状态
func parseRefundStatus(status refunddomestic.Status) (RefundStatus, error) {
	switch status {
	case refunddomestic.STATUS_SUCCESS: // 退款成功
		return RefundStatusSuccess, nil
	case refunddomestic.STATUS_CLOSED: // 退款关闭
		return RefundStatusClosed, nil
	case refunddomestic.STATUS_PROCESSING: // 退款处理中
		return RefundStatusProcessing, nil
	case refunddomestic.STATUS_ABNORMAL: // 退款异常
		return RefundStatusFailed, nil
	default:
		return "", fmt.Errorf("unknown refund status: %s", status)
	}
}

// checkRefundFields 检查退款响应字段是否为 nil
func checkRefundFields(refund *refunddomestic.Refund) error {
	if refund.Status == nil ||
		refund.OutTradeNo == nil ||
		refund.TransactionId == nil ||
		refund.RefundId == nil ||
		refund.Amount == nil ||
		refund.Amount.Total == nil ||
		refund.Amount.Refund == nil {
		return errors.New("response fields are nil")
	}

	return nil
}
