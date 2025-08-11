package inventory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/omniful/ims_rohit/pkg/pg"
)

// --- Hub CRUD ---

type Hub struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateHub(ctx context.Context, hub *Hub) (int64, error) {
	db := pg.GetClient().DB
	query := `INSERT INTO hubs (name, address) VALUES ($1, $2) RETURNING id`
	err := db.QueryRowContext(ctx, query, hub.Name, hub.Address).Scan(&hub.ID)
	return hub.ID, err
}

func GetHub(ctx context.Context, id int64) (*Hub, error) {
	db := pg.GetClient().DB
	query := `SELECT id, name, address, created_at, updated_at FROM hubs WHERE id = $1`
	h := &Hub{}
	err := db.QueryRowContext(ctx, query, id).Scan(&h.ID, &h.Name, &h.Address, &h.CreatedAt, &h.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return h, err
}

func UpdateHub(ctx context.Context, hub *Hub) error {
	db := pg.GetClient().DB
	query := `UPDATE hubs SET name = $1, address = $2, updated_at = NOW() WHERE id = $3`
	_, err := db.ExecContext(ctx, query, hub.Name, hub.Address, hub.ID)
	return err
}

func DeleteHub(ctx context.Context, id int64) error {
	db := pg.GetClient().DB
	query := `DELETE FROM hubs WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

func ListHubs(ctx context.Context) ([]*Hub, error) {
	db := pg.GetClient().DB
	query := `SELECT id, name, address, created_at, updated_at FROM hubs`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hubs []*Hub
	for rows.Next() {
		h := &Hub{}
		if err := rows.Scan(&h.ID, &h.Name, &h.Address, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		hubs = append(hubs, h)
	}
	return hubs, nil
}

// --- SKU CRUD & Filtering ---

type SKU struct {
	ID        int64     `json:"id"`
	TenantID  int64     `json:"tenant_id"`
	SellerID  int64     `json:"seller_id"`
	SKUCode   string    `json:"sku_code"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateSKU(ctx context.Context, sku *SKU) (int64, error) {
	db := pg.GetClient().DB
	query := `INSERT INTO skus (tenant_id, seller_id, sku_code, name) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.QueryRowContext(ctx, query, sku.TenantID, sku.SellerID, sku.SKUCode, sku.Name).Scan(&sku.ID)
	return sku.ID, err
}

func GetSKU(ctx context.Context, id int64) (*SKU, error) {
	db := pg.GetClient().DB
	query := `SELECT id, tenant_id, seller_id, sku_code, name, created_at, updated_at FROM skus WHERE id = $1`
	s := &SKU{}
	err := db.QueryRowContext(ctx, query, id).Scan(&s.ID, &s.TenantID, &s.SellerID, &s.SKUCode, &s.Name, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return s, err
}

func UpdateSKU(ctx context.Context, sku *SKU) error {
	db := pg.GetClient().DB
	query := `UPDATE skus SET name = $1, updated_at = NOW() WHERE id = $2`
	_, err := db.ExecContext(ctx, query, sku.Name, sku.ID)
	return err
}

func DeleteSKU(ctx context.Context, id int64) error {
	db := pg.GetClient().DB
	query := `DELETE FROM skus WHERE id = $1`
	_, err := db.ExecContext(ctx, query, id)
	return err
}

func ListSKUs(ctx context.Context, tenantID, sellerID *int64, skuCodes []string) ([]*SKU, error) {
	db := pg.GetClient().DB
	var (
		conds  []string
		args   []interface{}
		argIdx = 1
	)
	if tenantID != nil {
		conds = append(conds, fmt.Sprintf("tenant_id = $%d", argIdx))
		args = append(args, *tenantID)
		argIdx++
	}
	if sellerID != nil {
		conds = append(conds, fmt.Sprintf("seller_id = $%d", argIdx))
		args = append(args, *sellerID)
		argIdx++
	}
	if len(skuCodes) > 0 {
		placeholders := []string{}
		for i := range skuCodes {
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIdx))
			args = append(args, skuCodes[i])
			argIdx++
		}
		conds = append(conds, fmt.Sprintf("sku_code IN (%s)", strings.Join(placeholders, ",")))
	}
	query := `SELECT id, tenant_id, seller_id, sku_code, name, created_at, updated_at FROM skus`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var skus []*SKU
	for rows.Next() {
		s := &SKU{}
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SellerID, &s.SKUCode, &s.Name, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		skus = append(skus, s)
	}
	return skus, nil
}

// --- Inventory APIs ---

type Inventory struct {
	HubID int64 `json:"hub_id"`
	SKUID int64 `json:"sku_id"`
	Qty   int64 `json:"quantity"`
}

func UpsertInventory(ctx context.Context, hubID, skuID, qty int64) error {
	const query = `
	INSERT INTO inventory (hub_id, sku_id, quantity, updated_at)
	VALUES ($1, $2, $3, NOW())
	ON CONFLICT (hub_id, sku_id)
	DO UPDATE 
	SET quantity = inventory.quantity + EXCLUDED.quantity,
	    updated_at = NOW();
	`
	db := pg.GetClient().DB

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	_, err = tx.ExecContext(ctx, query, hubID, skuID, qty)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("upsert inventory failed: %w", err)
	}
	return nil
}

// View inventory for a hub and list of SKUs. Missing entries default to 0.
func ViewInventory(ctx context.Context, hubID int64, skuIDs []int64) ([]*Inventory, error) {
	db := pg.GetClient().DB
	var (
		rows *sql.Rows
		err  error
		invs []*Inventory
	)

	if len(skuIDs) == 0 {
		// Return all inventory for the hub
		const query = `SELECT sku_id, quantity FROM inventory WHERE hub_id = $1`
		rows, err = db.QueryContext(ctx, query, hubID)
	} else {
		// Return inventory for specific SKUs, defaulting to 0 if missing
		placeholders := make([]string, len(skuIDs))
		args := make([]interface{}, 0, len(skuIDs)+1)
		args = append(args, hubID)

		for i, id := range skuIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, id)
		}

		query := fmt.Sprintf(`
			SELECT s.id, COALESCE(i.quantity, 0)
			FROM skus s
			LEFT JOIN inventory i 
				ON i.sku_id = s.id AND i.hub_id = $1
			WHERE s.id IN (%s)
		`, strings.Join(placeholders, ","))

		rows, err = db.QueryContext(ctx, query, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		inv := &Inventory{HubID: hubID}
		if err := rows.Scan(&inv.SKUID, &inv.Qty); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invs, nil
}

// CheckSKUsExistence checks which SKUs from the given list exist in the database.
// Returns a map of skuID to bool (true if exists), and a slice of invalid (non-existent) skuIDs.
func CheckSKUsExistence(ctx context.Context, skuIDs []int64) (map[int64]bool, []int64, error) {
	db := pg.GetClient().DB
	if len(skuIDs) == 0 {
		return map[int64]bool{}, nil, nil
	}

	// Build query: SELECT id FROM skus WHERE id IN (...)
	placeholders := make([]string, len(skuIDs))
	args := make([]interface{}, len(skuIDs))
	for i, id := range skuIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	query := fmt.Sprintf("SELECT id FROM skus WHERE id IN (%s)", strings.Join(placeholders, ","))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	found := make(map[int64]bool, len(skuIDs))
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, nil, err
		}
		found[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	// Prepare result: for each input skuID, check if found
	existence := make(map[int64]bool, len(skuIDs))
	var invalid []int64
	for _, id := range skuIDs {
		if found[id] {
			existence[id] = true
		} else {
			existence[id] = false
			invalid = append(invalid, id)
		}
	}

	return existence, invalid, nil
}
