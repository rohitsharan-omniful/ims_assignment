package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omniful/ims_rohit/inventory"
	"github.com/omniful/ims_rohit/pkg/sku"
)

func CreateSKUHandler(c *gin.Context) {
	var req inventory.SKU
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.SKUCode == "" {
		skuCode := sku.GenerateSKUCode(req.Name, "SKU-")
		req.SKUCode = skuCode
	}

	id, err := inventory.CreateSKU(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.ID = id
	c.JSON(http.StatusCreated, req)
}

func GetSKUHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sku id"})
		return
	}
	sku, err := inventory.GetSKU(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sku == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "sku not found"})
		return
	}
	c.JSON(http.StatusOK, sku)
}

func UpdateSKUHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sku id"})
		return
	}
	var req inventory.SKU
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id
	if err := inventory.UpdateSKU(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

func DeleteSKUHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sku id"})
		return
	}
	if err := inventory.DeleteSKU(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func ListSKUsHandler(c *gin.Context) {
	// Optional query params: tenant_id, seller_id, sku_code (comma separated)
	var (
		tenantID *int64
		sellerID *int64
		skuCodes []string
	)
	if tid := c.Query("tenant_id"); tid != "" {
		if v, err := strconv.ParseInt(tid, 10, 64); err == nil {
			tenantID = &v
		}
	}
	if sid := c.Query("seller_id"); sid != "" {
		if v, err := strconv.ParseInt(sid, 10, 64); err == nil {
			sellerID = &v
		}
	}
	if codes := c.Query("sku_code"); codes != "" {
		skuCodes = splitAndTrim(codes)
	}
	skus, err := inventory.ListSKUs(c.Request.Context(), tenantID, sellerID, skuCodes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, skus)
}

func splitAndTrim(s string) []string {
	// Helper: splits comma-separated and trims spaces
	var out []string
	for _, v := range strings.Split(s, ",") {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// Inventory Handlers

type UpsertInventoryRequest struct {
	HubID int64 `json:"hub_id" binding:"required"`
	SKUID int64 `json:"sku_id" binding:"required"`
	Qty   int64 `json:"qty" binding:"required"`
}

func UpsertInventoryHandler(c *gin.Context) {
	var req UpsertInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := inventory.UpsertInventory(c.Request.Context(), req.HubID, req.SKUID, req.Qty); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

type ViewInventoryRequest struct {
	HubID  int64   `json:"hub_id" binding:"required"`
	SKUIDs []int64 `json:"sku_ids"`
}

func ViewInventoryHandler(c *gin.Context) {
	var req ViewInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	invs, err := inventory.ViewInventory(c.Request.Context(), req.HubID, req.SKUIDs)
	fmt.Println("invs", invs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invs)
}

type CheckSKUsExistenceRequest struct {
	SKUIDs []int64 `json:"sku_ids" binding:"required"`
}

type CheckSKUsExistenceResponse struct {
	Existence map[int64]bool `json:"existence"`
	Invalid   []int64        `json:"invalid"`
}

func CheckSKUsExistenceHandler(c *gin.Context) {
	var req CheckSKUsExistenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	existence, invalid, err := inventory.CheckSKUsExistence(c.Request.Context(), req.SKUIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp := CheckSKUsExistenceResponse{
		Existence: existence,
		Invalid:   invalid,
	}
	c.JSON(http.StatusOK, resp)
}
