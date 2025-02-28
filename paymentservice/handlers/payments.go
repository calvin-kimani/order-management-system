package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	Logger *zap.Logger
}

func NewPaymentHandler(logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{
		Logger: logger,
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var paymentRequest struct {
		OrderID     uint   `json:"order_id" binding:"required"`
		Amount      string `json:"amount" binding:"required"`
		PhoneNumber string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&paymentRequest); err != nil {
		h.Logger.Error("Invalid payment request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// 1. Get M-Pesa OAuth Token
	token, err := h.getMpesaToken()
	if err != nil {
		h.Logger.Error("Failed to get M-Pesa token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment initialization failed"})
		return
	}

	// 2. Generate STK Push password
	timestamp := time.Now().Format("20060102150405")
	password := base64.StdEncoding.EncodeToString([]byte(
		os.Getenv("MPESA_BUSINESS_SHORTCODE") +
			os.Getenv("MPESA_PASSKEY") +
			timestamp,
	))

	// 3. Create STK Push request
	stkRequest := map[string]interface{}{
		"BusinessShortCode": os.Getenv("MPESA_BUSINESS_SHORTCODE"),
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            paymentRequest.Amount,
		"PartyA":            paymentRequest.PhoneNumber,
		"PartyB":            os.Getenv("MPESA_BUSINESS_SHORTCODE"),
		"PhoneNumber":       paymentRequest.PhoneNumber,
		"CallBackURL":       os.Getenv("MPESA_CALLBACK_URL"),
		"AccountReference":  fmt.Sprintf("ORDER_%d", paymentRequest.OrderID),
		"TransactionDesc":   "Order Payment",
	}

	payload, _ := json.Marshal(stkRequest)
	req, _ := http.NewRequest("POST",
		"https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest",
		bytes.NewBuffer(payload),
	)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		h.Logger.Error("STK Push request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment processing failed"})
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != http.StatusOK {
		h.Logger.Error("STK Push failed",
			zap.Any("response", result),
			zap.Int("status_code", resp.StatusCode),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Payment failed",
			"detail": result,
		})
		return
	}

	h.Logger.Info("Payment initiated successfully",
		zap.Any("response", result),
		zap.Uint("order_id", paymentRequest.OrderID),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":             "Payment initiated",
		"checkout_request_id": result["CheckoutRequestID"],
	})
}

func (h *PaymentHandler) getMpesaToken() (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(
		os.Getenv("MPESA_CONSUMER_KEY") + ":" + os.Getenv("MPESA_CONSUMER_SECRET"),
	))

	req, _ := http.NewRequest("GET",
		"https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials",
		nil,
	)
	req.Header.Add("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return result.AccessToken, nil
}

func (h *PaymentHandler) PaymentCallback(c *gin.Context) {
	var callback struct {
		Body struct {
			StkCallback struct {
				CheckoutRequestID string `json:"CheckoutRequestID"`
				ResultCode        int    `json:"ResultCode"`
				ResultDesc        string `json:"ResultDesc"`
				CallbackMetadata  struct {
					Item []struct {
						Name  string      `json:"Name"`
						Value interface{} `json:"Value"`
					} `json:"Item"`
				} `json:"CallbackMetadata"`
			} `json:"stkCallback"`
		} `json:"Body"`
	}

	if err := c.ShouldBindJSON(&callback); err != nil {
		h.Logger.Error("Invalid callback format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback format"})
		return
	}

	status := "failed"
	if callback.Body.StkCallback.ResultCode == 0 {
		status = "paid"
	}

	// Update order status in Orders Service
	updateURL := fmt.Sprintf("%s/orders/%s/status",
		os.Getenv("ORDERS_SERVICE_URL"),
		callback.Body.StkCallback.CheckoutRequestID,
	)

	_, err := http.Post(updateURL, "application/json",
		strings.NewReader(fmt.Sprintf(`{"status": "%s"}`, status)))

	if err != nil {
		h.Logger.Error("Failed to update order status",
			zap.Error(err),
			zap.String("checkout_id", callback.Body.StkCallback.CheckoutRequestID),
		)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Callback processed"})
}