package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PanymentHandler struct {
}

func NewPaymentHandler() *PanymentHandler {
	return &PanymentHandler{}
}

func (h *PanymentHandler) RegisterRoutes(c *gin.Engine) {
	c.POST("/api/webhook", h.handleWebhook)
}

func (h *PanymentHandler) handleWebhook(c *gin.Context) {
	logrus.Info("Got webhook from stripe")
}
