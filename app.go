package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gteca/bank-app/db"
	"github.com/gteca/bank-app/operations"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

const (
	DBNAME   = "bankOfAmerica"
	DBUSER   = "root"
	DBPASSWD = "fakebank1234"
)

type Api struct {
	Router *mux.Router
	DB     *sql.DB
}

type Grpc struct {
	operations.UnimplementedOperationsServer
	DB *sql.DB
}

const (
	GRPC_SUCCESS               = "Success"
	GRPC_NO_USER_FOUND         = "No User with such credit card found"
	GRPC_INTERNAL_SERVER_ERROR = "Internal Server Error"
)

func InitDB() (*sql.DB, error) {

	connectionInfo := fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v", DBUSER, DBPASSWD, DBNAME)
	var err error
	db, err := sql.Open("mysql", connectionInfo)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (server *Grpc) InitGrpcServer() {

	dbConn, err := InitDB()
	if err != nil {
		log.Println("Error closing database:", err)
	}

	server.DB = dbConn
}

func (api *Api) InitApiServer() error {

	dbConn, err := InitDB()
	if err != nil {
		log.Println("Error closing database:", err)
	}

	api.DB = dbConn

	api.Router = mux.NewRouter().StrictSlash(true)
	api.HandleRoutes()

	return nil
}

func (api *Api) RunApiServer(ipPort string) {

	log.Printf("API Server listening on %v", ipPort)
	log.Fatal(http.ListenAndServe(ipPort, api.Router))
}

func (server *Grpc) RunGrpcServer(ipPort string) {

	log.Printf("GRPC Server listening on %v", ipPort)
	listener, err := net.Listen("tcp", ipPort)
	if err != nil {
		log.Fatalf("Error: %v - Failed to listen on : %v", err, ipPort)
	}

	grpcServer := grpc.NewServer()
	operations.RegisterOperationsServer(grpcServer, server)

	log.Fatal(grpcServer.Serve(listener))

}

func (server *Grpc) ExecutePayment(ctx context.Context, payment *operations.PaymentReq) (*operations.PaymentResp, error) {
	log.Printf("Received transaction request for amount: %v for card: %s", payment.Amount, payment.CardNumber)

	transactionId := uuid.New().String()

	account, result := server.getAccountByCardNumber(payment.CardNumber)
	if result != GRPC_SUCCESS {
		log.Printf("Failed to retrieve account for CardNumber: %s", payment.CardNumber)
		return &operations.PaymentResp{
			Success:       false,
			TransactionId: transactionId,
		}, errors.New(result)
	}

	proceed := server.HasSufficientBalance(account.Balance, payment.Amount)
	log.Printf("Proceed: %v", proceed)

	if !proceed {

		log.Printf("proceed? %v", proceed)

		return &operations.PaymentResp{
			Success:       false,
			TransactionId: transactionId,
		}, errors.New("User has no sufficient funds")
	}

	account.Balance -= payment.Amount

	if err := db.UpdateAccount(server.DB, &account); err != nil {
		return &operations.PaymentResp{
			Success:       false,
			TransactionId: transactionId,
		}, err
	}

	return &operations.PaymentResp{
		Success:       true,
		TransactionId: transactionId,
	}, nil
}

func (server *Grpc) getAccountByCardNumber(cardNumber string) (db.Account, string) {

	account, err := db.GetAccountByCardNumber(server.DB, cardNumber)
	var result string
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			result = GRPC_NO_USER_FOUND
		default:
			result = GRPC_INTERNAL_SERVER_ERROR
		}
	}
	result = GRPC_SUCCESS

	return account, result
}

func (server *Grpc) HasSufficientBalance(balance float32, amount float32) bool {

	return balance-amount > 0.0
}

func sendResponse(w http.ResponseWriter, statusCode int, payload interface{}) {

	response, _ := json.Marshal(payload)
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func sendError(w http.ResponseWriter, statusCode int, err string) {
	error_msg := map[string]string{"error": err}
	sendResponse(w, statusCode, error_msg)
}

func (api *Api) getAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := db.GetAccounts(api.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, accounts)
}

func (api *Api) getAccountByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Cannot parse user id")
	}

	account, err := db.GetAccountByID(api.DB, id)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			sendError(w, http.StatusBadRequest, "User not found")
		default:
			sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendResponse(w, http.StatusOK, account)
}

func (api *Api) createAccount(w http.ResponseWriter, r *http.Request) {

	var account db.Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = db.CreateAccount(api.DB, &account)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusCreated, account)
}

func (api *Api) updateAccount(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Cannot parse user id")
	}

	var account db.Account
	err = json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	account.Id = id
	err = db.UpdateAccount(api.DB, &account)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusNoContent, account)
}

func (api *Api) deleteAccount(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Cannot parse user id")
	}

	err = db.DeleteAccount(api.DB, id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, map[string]string{"result": "successful deletion"})
}

func (api *Api) HandleRoutes() {
	api.Router.HandleFunc("/account", api.getAccounts).Methods("GET")
	api.Router.HandleFunc("/account/{id}", api.getAccountByID).Methods("GET")
	api.Router.HandleFunc("/account", api.createAccount).Methods("POST")
	api.Router.HandleFunc("/account/{id}", api.updateAccount).Methods("PUT")
	api.Router.HandleFunc("/account/{id}", api.deleteAccount).Methods("DELETE")
}
