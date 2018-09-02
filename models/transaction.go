package models

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/rathvong/talentmob_server/system"
)

const (
	PurchaseStatePurchase                  = 0
	PurchaseStateCancelled                 = 1
	PurchaseStateNotValidated              = 2
	TransactionMerchantGooglePay           = "google_pay"
	TransactionMerchangeApplePay           = "apple_pay"
	TransactionTypeBuy                     = "buy"
	TransactionTypeExchangeToStarPowerGold = "exchange_star_gold"
	TransactionTypeEchangeToUS             = "exchange_to_us"
)

type Transaction struct {
	BaseModel
	UserID            uint64 `json:"user_id"`
	AmountDollar      int64  `json:"amount_dollar"`
	AmountStarPower   int64  `json:"amount_star_power"`
	Merchant          string `json:"merchant"`
	Type              string `json:"type"`
	ItemID            string `json:"item_id"`
	OrderID           string `json:"order_id"`
	PurchaseState     int    `json:"purchase_state"`
	PurchaseTimeMilis uint64 `json:"purchase_time_milis"`
	IsActive          bool   `json:"is_active"`
	PurchaseID        string `json:"purchase_id"`
	ConsumptionState  int    `json:"consumption_state"`
}

func (t *Transaction) merchantValid(merchant string) bool {

	switch merchant {
	case TransactionMerchantGooglePay, TransactionMerchangeApplePay:
		return true
	}

	return false
}

func (t *Transaction) typeValid(tt string) bool {

	switch tt {
	case TransactionTypeBuy,
		TransactionTypeExchangeToStarPowerGold,
		TransactionTypeEchangeToUS:
		return true
	}

	return false
}

func (t *Transaction) createErrors() error {
	if t.OrderID == "" {
		return t.Errors(ErrorMissingValue, "order_id")
	}

	if t.ItemID == "" {
		return t.Errors(ErrorMissingValue, "order_id")
	}

	if !t.merchantValid(t.Merchant) {
		return t.Errors(ErrorIncorrectValue, "merchant")
	}

	if !t.typeValid(t.Type) {
		return t.Errors(ErrorIncorrectValue, "type")
	}

	if t.UserID == 0 {
		return t.Errors(ErrorMissingValue, "user_id")
	}

	if t.validPurchaseState(t.PurchaseState) {
		return t.Errors(ErrorIncorrectValue, "purchase_state: "+fmt.Sprintf("%d, requires 0, 1, 2", t.PurchaseState))
	}

	return nil
}

func (t *Transaction) validPurchaseState(state int) bool {

	switch state {
	case PurchaseStatePurchase, PurchaseStateCancelled, PurchaseStateNotValidated:
		return true
	}

	return false
}

func (t *Transaction) updateError() error {

	if t.ID == 0 {
		return t.Errors(ErrorMissingID, "id")
	}

	return t.createErrors()
}

func (t *Transaction) Create(db *system.DB) error {

	if err := t.createErrors(); err != nil {
		return err
	}

	qry := `INSERT INTO transactions (
			user_id, 
			amount_dollar,
			amount_star_power,
			merchant,
			type,
			item_id,
			order_id,
			purchase_state,
			purchase_time_milis,
			purchase_id,
			consumption_state,
			is_active,
			created_at,
			updated_at	
			) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,$12	
			) RETURNING id
			`

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	if err != nil {
		return err
	}

	err = tx.QueryRow(qry,
		t.UserID,
		t.AmountDollar,
		t.AmountStarPower,
		t.Merchant,
		t.Type,
		t.ItemID,
		t.OrderID,
		t.PurchaseState,
		t.PurchaseTimeMilis,
		t.PurchaseID,
		t.ConsumptionState,
		t.IsActive,
		t.CreatedAt,
		t.UpdatedAt).Scan(t.ID)

	if err != nil {
		log.Printf("Transaction.Create() OrderID: %s \nQuery: %s   \nError: %v", t.OrderID, qry, err)
		return err
	}

	return nil
}

func (t *Transaction) Update(db *system.DB) error {

	if err := t.updateError(); err != nil {
		return err
	}

	qry := `UPDATE transactions SET
			user_id = $2,
			amount_dollar = $3,
			amount_star_power = $4,
			merchant = $5,
			type = $6,
			item_id = $7,
			order_id = $8,
			purchase_state = $9,
			purchase_time_milis = $10,
			consumption_state = $11,
			is_active = $12,
			updated_at = $13,
			purchase_id = $14,
			WHERE id = $1
			`

	tx, err := db.Begin()

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return
		}
	}()

	if err != nil {
		return err
	}

	_, err = tx.Exec(qry,
		t.ID,
		t.UserID,
		t.AmountDollar,
		t.AmountStarPower,
		t.Merchant,
		t.Type,
		t.ItemID,
		t.OrderID,
		t.PurchaseState,
		t.PurchaseTimeMilis,
		t.ConsumptionState,
		t.IsActive,
		t.UpdatedAt,
		t.PurchaseID)

	if err != nil {
		log.Printf("Transaction.Update() OrderID: %s \nQuery: %s   \nError: %v", t.OrderID, qry, err)
		return err
	}

	return nil
}

func (t *Transaction) Get(db *system.DB, id uint64) error {

	qry := `SELECT 
				id,
				user_id, 
				amount_dollar,
				amount_star_power,
				merchant,
				type,
				item_id,
				order_id,
				purchase_state,
				purchase_time_milis,
				consumption_state,
				is_active,
				created_at,
				updated_at,
				purchase_id,	
			FROM transactions
			WHERE id = $1	
			`

	err := db.QueryRow(qry, id).Scan(
		&t.ID,
		&t.UserID,
		&t.AmountDollar,
		&t.AmountStarPower,
		&t.Merchant,
		&t.Type,
		&t.ItemID,
		&t.OrderID,
		&t.PurchaseState,
		&t.PurchaseTimeMilis,
		&t.ConsumptionState,
		&t.IsActive,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.PurchaseID)

	if err != nil {
		log.Printf("Transaction.Get() OrderID: %s \nQuery: %s   \nError: %v", t.OrderID, qry, err)
		return err
	}

	return nil

}

func (t *Transaction) GetAllForUser(db *system.DB, userID uint64, page int) ([]Transaction, error) {

	qry := `SELECT 
				id,
				user_id, 
				amount_dollar,
				amount_star_power,
				merchant,
				type,
				item_id,
				order_id,
				purchase_state,
				purchase_time_milis,
				consumption_state,
				is_active,
				created_at,
				updated_at,
				purchase_id	
			FROM transactions
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
			OFFSET $3	
			`

	rows, err := db.Query(qry, userID, LimitQueryPerRequest, OffSet(page))

	if err != nil {
		log.Printf("Transaction.GetAllForUser()userID: %s \nQuery: %s \nError: %v", userID, qry, err)
		return nil, err
	}

	defer rows.Close()

	return t.parseRows(rows)

}

func (t *Transaction) parseRows(rows *sql.Rows) ([]Transaction, error) {
	var transactions []Transaction

	for rows.Next() {

		transaction := Transaction{}

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.AmountDollar,
			&transaction.AmountStarPower,
			&transaction.Merchant,
			&transaction.Type,
			&transaction.ItemID,
			&transaction.OrderID,
			&transaction.PurchaseState,
			&transaction.PurchaseTimeMilis,
			&transaction.ConsumptionState,
			&transaction.IsActive,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
			&transaction.PurchaseID,
		)

		if err != nil {

			return transactions, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
