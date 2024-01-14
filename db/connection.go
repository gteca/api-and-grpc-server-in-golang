package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type Account struct {
	Id           int     `json:"id"`
	Name         string  `json:"name"`
	Balance      float32 `json:"balance"`
	CardNumber   string  `json:"cardnumber"`
	IsCardActive bool    `json:"iscardactive"`
}

func GetAccountByCardNumber(conndb *sql.DB, cardNumber string) (Account, error) {

	query := fmt.Sprintf("select * from accounts where card_number = % v", cardNumber)

	row := conndb.QueryRow(query)
	account := Account{}
	err := row.Scan(&account.Id, &account.Name, &account.Balance, &account.CardNumber, &account.IsCardActive)

	return account, err
}

func GetAccounts(conndb *sql.DB) ([]Account, error) {

	query := "select * from accounts"
	rows, err := conndb.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	accounts := []Account{}

	for rows.Next() {
		account := Account{}
		err = rows.Scan(&account.Id, &account.Name, &account.Balance, &account.CardNumber, &account.IsCardActive)
		if err != nil {
			return accounts, err
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func GetAccountByID(conndb *sql.DB, id int) (Account, error) {

	query := fmt.Sprintf("select * from accounts where id = % v", id)

	row := conndb.QueryRow(query)
	account := Account{}
	err := row.Scan(&account.Id, &account.Name, &account.Balance, &account.CardNumber, &account.IsCardActive)

	return account, err
}

func CreateAccount(conndb *sql.DB, account *Account) error {

	query := fmt.Sprintf("insert into accounts (name, balance, card_number, is_card_active) values ('%v', %v, '%v', %v)",
		account.Name, account.Balance, account.CardNumber, account.IsCardActive)

	result, err := conndb.Exec(query)
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func UpdateAccount(conndb *sql.DB, account *Account) error {

	query := fmt.Sprintf("update accounts set name = '%v', balance = %v, card_number = '%v', is_card_active = %v where id = %v",
		account.Name, account.Balance, account.CardNumber, account.IsCardActive, account.Id)

	result, err := conndb.Exec(query)
	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("No such row exists")
	}

	return err
}

func DeleteAccount(conndb *sql.DB, id int) error {

	query := fmt.Sprintf("delete from accounts where id = %v", id)

	result, err := conndb.Exec(query)
	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("No such row exists")
	}

	return err
}
