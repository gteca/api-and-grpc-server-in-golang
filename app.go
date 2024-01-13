package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/gteca/bank-app/db"
	"github.com/gteca/bank-app/operations"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

const (
	DBNAME   = "bankOfAmerica"
	DBUSER   = "root"
	DBPASSWD = "fakebank1234"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

type bankServer struct {
	operations.UnimplementedOperationsServer
}

func (app *App) Initialise() error {

	connectionInfo := fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v", DBUSER, DBPASSWD, DBNAME)
	var err error
	app.DB, err = sql.Open("mysql", connectionInfo)
	if err != nil {
		return err
	}

	app.Router = mux.NewRouter().StrictSlash(true)
	app.HandleRoutes()

	return nil
}

func (app *App) RunApiServer(ipPort string) {

	log.Printf("API Server listening on %v", ipPort)
	log.Fatal(http.ListenAndServe(ipPort, app.Router))
}

func (app *App) RunGrpcServer(ipPort string) {

	log.Printf("GRPC Server listening on %v", ipPort)
	listener, err := net.Listen("tcp", ipPort)
	if err != nil {
		log.Fatalf("Error: %v - Failed to listen on : %v", err, ipPort)
	}

	grpcServer := grpc.NewServer()
	operations.RegisterOperationsServer(grpcServer, &bankServer{})

	log.Fatal(grpcServer.Serve(listener))
}

func (s bankServer) ExecutePayment(ctx context.Context, payment *operations.Payment) (*operations.PaymentStatus, error) {

	log.Printf("Received transaction request for amount: %d from UserId: %s", payment.Amount, payment.UserId)

	response := &operations.PaymentStatus{
		Result:        "success",
		TransactionId: "ewdewndewn4330439r3329",
	}
	return response, nil
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

func (app *App) getAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := db.GetAccounts(app.DB)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, accounts)
}

func (app *App) getAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Cannot parse user id")
	}

	account, err := db.GetAccount(app.DB, id)
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

func (app *App) createAccount(w http.ResponseWriter, r *http.Request) {

	var account db.Account
	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = db.CreateAccount(app.DB, &account)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusCreated, account)
}

func (app *App) updateAccount(w http.ResponseWriter, r *http.Request) {

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
	err = db.UpdateAccount(app.DB, &account)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusNoContent, account)
}

func (app *App) deleteAccount(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Cannot parse user id")
	}

	err = db.DeleteAccount(app.DB, id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, map[string]string{"result": "successful deletion"})
}

func (app *App) HandleRoutes() {
	app.Router.HandleFunc("/account", app.getAccounts).Methods("GET")
	app.Router.HandleFunc("/account/{id}", app.getAccount).Methods("GET")
	app.Router.HandleFunc("/account", app.createAccount).Methods("POST")
	app.Router.HandleFunc("/account/{id}", app.updateAccount).Methods("PUT")
	app.Router.HandleFunc("/account/{id}", app.deleteAccount).Methods("DELETE")
}
