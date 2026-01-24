//
// FilePath    : go-utils\pay\alipay.go
// Author      : jiaopengzi
// Blog        : https://jiaopengzi.com
// Copyright   : Copyright (c) 2026 by jiaopengzi, All Rights Reserved.
// Description : 支付宝支付
//

package pay

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jiaopengzi/go-utils"

	"github.com/smartwalle/alipay/v3"
	"go.uber.org/zap"
)

const (
	// AlipayTradeTypePage 支付宝网页支付
	AlipayTradeTypePage = "alipay_trade_page_pay"

	// 文档: https://opendocs.alipay.com/open/357441a2_alipay.trade.fastpay.refund.query?scene=common&pathHash=01981dca
	AlipayTradeTypeRefundSuccess = "REFUND_SUCCESS"
)

// AlipayConfig 支付宝支付配置
type AlipayConfig struct {
	Enabled         bool   `mapstructure:"enabled" json:"enabled"`                                                                               // 是否启用支付宝支付
	IsProduction    bool   `mapstructure:"is_production" json:"is_production" example:"true"`                                                    // 是否为生产环境，默认为 false（沙箱环境）
	AppID           string `mapstructure:"app_id" json:"app_id" binding:"required_if=Enabled true" example:"2021000118631234"`                   // 支付宝应用ID
	SellerID        string `mapstructure:"seller_id" json:"seller_id" binding:"required_if=Enabled true" example:"2088102180131234"`             // 支付宝商户ID
	AppPrivateKey   string `mapstructure:"app_private_key" json:"app_private_key" binding:"required_if=Enabled true" example:"key"`              // 支付商户私钥
	AlipayPublicKey string `mapstructure:"alipay_public_key" json:"alipay_public_key" binding:"required_if=Enabled true" example:"pubkey"`       // 支付宝公钥
	EncryptKey      string `mapstructure:"encrypt_key" json:"encrypt_key" example:"A9A5/JJ4tcRrs/KL0MNg=="`                                      // 可选，接口内容加密密钥
	NotifyHost      string `mapstructure:"notify_host" json:"notify_host" binding:"required_if=Enabled true" example:"https://example.com:8080"` // 支付结果通知主机地址
	NotifyPath      string `mapstructure:"notify_path" json:"notify_path" binding:"required_if=Enabled true" example:"/alipay/notify"`           // 支付结果通知路由
	RefundPath      string `mapstructure:"refund_path" json:"refund_path" binding:"required_if=Enabled true" example:"/alipay/refund_notify"`    // 退款结果通知路由
}

// Alipay 支付宝支付实现
type Alipay struct {
	Client      *alipay.Client // Alipay 客户端
	Conf        *AlipayConfig  // 支付宝配置
	APIPath     string         // API 路径前缀 e.g. /api/v1
	PayBasePath string         // 支付基础路由 e.g. /pay
}

// NewAlipay 创建新的支付宝支付实例
//   - conf: 支付宝支付配置
//   - apiPath: API 路径前缀 e.g. /api/v1
//   - payBasePath: 支付基础路由 e.g. /pay
func NewAlipay(conf *AlipayConfig, apiPath, payBasePath string) (*Alipay, error) {
	client, err := alipay.New(
		conf.AppID,
		conf.AppPrivateKey,
		conf.IsProduction,
	)

	// 创建 Alipay 客户端失败
	if err != nil {
		return nil, fmt.Errorf("create Alipay client error: %w", err)
	}

	// 加载支付宝公钥
	if err = client.LoadAliPayPublicKey(conf.AlipayPublicKey); err != nil {
		return nil, fmt.Errorf("load Alipay public key error: %w", err)
	}

	// 设置接口内容加密密钥
	if conf.EncryptKey != "" {
		if err = client.SetEncryptKey(conf.EncryptKey); err != nil {
			return nil, fmt.Errorf("set Alipay encrypt key error: %w", err)
		}
	}

	// appPath 和 payBasePath 不为空
	if apiPath == "" || payBasePath == "" {
		return nil, fmt.Errorf("apiPath and payBasePath cannot be empty")
	}

	// 返回支付宝支付实例
	return &Alipay{
		Client:      client,
		Conf:        conf,
		APIPath:     apiPath,
		PayBasePath: payBasePath,
	}, nil
}

// Prepay 支付宝支付实现
//   - orderID: 订单ID
//   - amount: 金额，单位为分
//   - description: 商品描述
//   - returnURL: 支付完成后跳转的页面
//
// 返回值为支付链接, 在浏览器中打开即可完成支付
func (a *Alipay) Prepay(orderID uint64, amount int64, description, returnURL string, timeExpire time.Time) (string, error) {
	// 文档: https://github.com/smartwalle/alipay/tree/master
	// 支付结果通知地址
	notifyURL := fmt.Sprintf("%s/%s%s%s",
		a.Conf.NotifyHost,
		a.APIPath,
		a.PayBasePath,
		a.Conf.NotifyPath,
	)

	// 网站端支付使用 TradePagePay
	var p = alipay.TradePagePay{
		Trade: alipay.Trade{
			NotifyURL:   notifyURL,
			ReturnURL:   returnURL, // 支付完成后跳转的页面
			Subject:     description,
			OutTradeNo:  utils.Uint64ToStr(orderID),
			TotalAmount: utils.Int64FenToStrYuan(amount),          // 金额单位为元
			ProductCode: "FAST_INSTANT_TRADE_PAY",                 // 支付宝产品码默认
			TimeExpire:  timeExpire.Format("2006-01-02 15:04:05"), // 订单失效时间, 格式为yyyy-MM-dd HH:mm:ss
		},
		// 是否自定义二维码
		// 文档:https://opendocs.alipay.com/open/59da99d0_alipay.trade.page.pay?scene=22&pathHash=e26b497f
		QRPayMode:   "4",   // 使用二维码支付模式
		QRCodeWidth: "200", // 二维码宽度，单位为像素
	}

	url, err := a.Client.TradePagePay(p)
	if err != nil {
		return "", fmt.Errorf("alipay prepay error: %w", err)
	}

	// 打印日志确认支付宝支付链接生成成功
	zap.L().Debug("Alipay prepay URL generated successfully", zap.String("url", url.String()))

	// url 类型为 url.URL, 转成 string
	return url.String(), err
}

// GetNotifyPayment 支付宝支付实现应答支付结果通知接口, 包含验签和获取支付结果
func (a *Alipay) GetNotifyPayment(request *http.Request) (bool, *PaymentResult, error) {
	// 文档: https://github.com/smartwalle/alipay/tree/master
	if err := request.ParseForm(); err != nil {
		// 如果 err 不为空，则表示解析表单失败
		return false, nil, fmt.Errorf("alipay notify parse form error: %w", err)
	}

	notif, err := a.Client.DecodeNotification(request.Form)
	if err != nil {
		// 如果 err 不为空，则表示验签失败
		return false, nil, fmt.Errorf("alipay notify verify sign error: %w", err)
	}

	// 为了确保支付状态正确，检查 TradeStatus
	if notif.TradeStatus != alipay.TradeStatusSuccess && notif.TradeStatus != alipay.TradeStatusFinished {
		return false, nil, fmt.Errorf("alipay trade status not success: %s", notif.TradeStatus)
	}

	result := &PaymentResult{
		PayType:       PayTypeAlipay, // 支付宝支付类型
		OrderID:       utils.StrToUint64(notif.OutTradeNo),
		TotalAmount:   utils.StrYuanToInt64Fen(notif.TotalAmount), // 转换为分
		TransactionID: notif.TradeNo,
		TradeState:    TradeStatePaid,
		TradeType:     AlipayTradeTypePage,
		AppID:         notif.AppId,    // 支付宝应用ID
		SellerID:      notif.SellerId, // 支付宝商户ID
	}

	return true, result, nil
}

// ValidateNotifyPayment 支付宝支付实现验证支付结果通知接口
// 主要校验商户订单号、金额、商户号、appid 等信息是否匹配
//
//nolint:dupl
func (a *Alipay) ValidateNotifyPayment(payment *PaymentResult, orderID uint64, amount int64) (bool, *PaymentResult, error) {
	// 校验 payment 是否为 nil
	if payment == nil {
		return false, nil, fmt.Errorf("alipay validate notify payment error: payment is nil")
	}

	// 校验订单号
	if payment.OrderID != orderID {
		return false, nil, fmt.Errorf("alipay validate notify payment error: order ID mismatch, expected %d, got %d", orderID, payment.OrderID)
	}

	// 校验金额
	if payment.TotalAmount != amount {
		return false, nil, fmt.Errorf("alipay validate notify payment error: amount mismatch, expected %d, got %d", amount, payment.TotalAmount)
	}

	// 校验商户号
	if payment.SellerID != a.Conf.SellerID {
		return false, nil, fmt.Errorf("alipay validate notify payment error: seller ID mismatch expected %s, got %s", a.Conf.SellerID, payment.SellerID)
	}

	// 校验应用ID
	if payment.AppID != a.Conf.AppID {
		return false, nil, fmt.Errorf("alipay validate notify payment error: app ID mismatch, expected %s, got %s", a.Conf.AppID, payment.AppID)
	}

	return true, payment, nil
}

// QueryPayment 支付宝支付实现查询支付结果接口
func (a *Alipay) QueryPayment(orderID uint64) (*PaymentResult, error) {
	var p = alipay.TradeQuery{
		OutTradeNo: utils.Uint64ToStr(orderID),
	}

	resultQuery, err := a.Client.TradeQuery(context.Background(), p)
	if err != nil {
		return nil, fmt.Errorf("alipay query payment error: %w", err)
	}

	// 支付结果
	result := &PaymentResult{
		PayType:       PayTypeAlipay, // 支付宝支付类型
		OrderID:       orderID,
		TotalAmount:   utils.StrYuanToInt64Fen(resultQuery.TotalAmount), // 转换为分
		TransactionID: resultQuery.TradeNo,
		TradeType:     AlipayTradeTypePage,
	}

	// 处理没有查询到订单的情况, 说明没有执行支付
	if resultQuery.Code.IsFailure() {
		zap.L().Debug("支付宝支付查询，该订单不存在", zap.Uint64("order_id", orderID))

		result.TradeState = TradeStateUnpaid // 设置为未支付状态

		return result, nil
	}

	var state TradeState

	// 状态对齐
	switch resultQuery.TradeStatus {
	case alipay.TradeStatusWaitBuyerPay: // 等待买家付款
		state = TradeStateUnpaid
	case alipay.TradeStatusClosed: // 交易关闭
		state = TradeStateClosed
	case alipay.TradeStatusSuccess: // 交易支付成功
		state = TradeStatePaid
	case alipay.TradeStatusFinished: // 交易结束，不可退款
		state = TradeStatePaid
	default:
		return nil, fmt.Errorf("alipay trade status not recognized: %s", resultQuery.TradeStatus)
	}

	// 设置支付状态
	result.TradeState = state

	return result, nil
}

// CloseOrder 支付宝支付实现关闭订单接口
func (a *Alipay) CloseOrder(orderID uint64) error {
	var p = alipay.TradeClose{
		OutTradeNo: utils.Uint64ToStr(orderID),
	}

	result, err := a.Client.TradeClose(context.Background(), p)
	if err != nil {
		return fmt.Errorf("alipay cancel order error: %w", err)
	}

	// 用户未进行交互比如扫码或者登录，支付宝远端不会创建订单会得到一个 40004 的错误码
	// 当做正常关单处理
	if result.Code == alipay.CodeBusinessFailed {
		zap.L().Debug("用户未进行交互，支付宝远端未创建交易订单", zap.Uint64("order_id", orderID))
		return nil
	}

	// 如果返回的 code 是失败状态，记录日志并返回错误
	if result.Code.IsFailure() {
		return fmt.Errorf("alipay cancel order failed: code %s, msg %s", result.Code, result.Msg)
	}

	zap.L().Info("Alipay order closed successfully", zap.Uint64("order_id", orderID))

	return nil
}

// Refund 支付宝支付实现退款接口
func (a *Alipay) Refund(orderID, refundID uint64, amount, refundAmount int64, reason string) (*RefundResult, error) {
	refundAmountYuan := utils.Int64FenToStrYuan(refundAmount) // 将分转换为元，保留两位小数

	// 默认当退款金额等于订单金额时，OutRequestNo 使用订单ID
	outRequestNo := utils.Uint64ToStr(orderID)

	// 网站端支付使用 TradePagePay
	var p = alipay.TradeRefund{
		OutTradeNo:   utils.Uint64ToStr(orderID),
		RefundReason: reason,
		RefundAmount: refundAmountYuan, // 退款金额，单位为元
	}

	// 文档: https://opendocs.alipay.com/open/357441a2_alipay.trade.fastpay.refund.query?scene=common&pathHash=01981dca
	// 当退款金额小于订单金额时，需要传入 OutRequestNo
	if refundAmount < amount {
		// 当退款金额小于订单金额时，OutRequestNo 使用退款ID
		outRequestNo = utils.Uint64ToStr(refundID)
		p.OutRequestNo = outRequestNo
	} else {
		// 当退款金额等于订单金额时，OutRequestNo 使用订单ID
		refundID = orderID
	}

	result, err := a.Client.TradeRefund(context.Background(), p)
	if err != nil {
		return nil, fmt.Errorf("alipay refund error: %w", err)
	}

	if result.Code.IsFailure() {
		return nil, fmt.Errorf("alipay refund failed: code %s, msg %s", result.Code, result.Msg)
	}

	zap.L().Debug("Alipay refund successful", zap.Uint64("order_id", orderID), zap.Uint64("refund_id", refundID))

	// 和微信支付不同这里只要 result.Code.IsSuccess() 就表示退款请求成功
	// 没有异步通知
	resultRefund := &RefundResult{
		PayType:             PayTypeAlipay, // 支付宝支付类型
		RefundID:            refundID,
		OrderID:             orderID,
		TransactionID:       result.TradeNo,
		RefundTransactionID: outRequestNo,        // 支付宝没有单独的退款交易ID, 所以使用商家的退款ID
		TotalAmount:         amount,              // 订单总金额，单位为分
		RefundAmount:        refundAmount,        // 退款金额，单位为分
		Reason:              reason,              // 退款原因
		Status:              RefundStatusSuccess, // 退款状态
	}

	return resultRefund, nil
}

// GetNotifyRefund 支付宝支付实现应答退款结果通知接口
func (a *Alipay) GetNotifyRefund(request *http.Request) (bool, *RefundResult, error) {
	// 由于支付宝退款没有异步通知，所以这里直接返回成功
	result := &RefundResult{
		PayType: PayTypeAlipay,       // 支付宝支付类型
		Status:  RefundStatusSuccess, // 退款状态
	}

	return true, result, nil
}

// QueryRefund 支付宝支付实现查询退款结果接口
func (a *Alipay) QueryRefund(orderID, refundID uint64) (*RefundResult, error) {
	var p = alipay.TradeFastPayRefundQuery{
		OutTradeNo:   utils.Uint64ToStr(orderID),
		OutRequestNo: utils.Uint64ToStr(refundID),
	}

	resultQuery, err := a.Client.TradeFastPayRefundQuery(context.Background(), p)
	if err != nil {
		return nil, fmt.Errorf("alipay query refund error: %w", err)
	}

	// 处理没有查询到订单的情况
	if resultQuery.Code.IsFailure() {
		return nil, fmt.Errorf("支付宝退款查询，该订单不存在, 订单id: %d, 退款id: %d", orderID, refundID)
	}

	// 状态对齐
	if resultQuery.RefundStatus != AlipayTradeTypeRefundSuccess {
		return nil, fmt.Errorf("alipay refund status not recognized: %s", resultQuery.RefundStatus)
	}

	resultRefund := &RefundResult{
		PayType:             PayTypeAlipay, // 支付宝支付类型
		RefundID:            refundID,
		OrderID:             orderID,
		TransactionID:       resultQuery.TradeNo,
		RefundTransactionID: resultQuery.OutRequestNo,                          // 支付宝没有单独的退款交易ID, 所以使用商家的退款ID
		TotalAmount:         utils.StrYuanToInt64Fen(resultQuery.TotalAmount),  // 订单总金额，单位为分
		RefundAmount:        utils.StrYuanToInt64Fen(resultQuery.RefundAmount), // 退款金额，单位为分
		Status:              RefundStatusSuccess,                               // 退款状态
	}

	return resultRefund, nil
}
