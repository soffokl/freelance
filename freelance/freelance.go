package freelance

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/soffokl/freelance/db"
)

type Exchange interface {
	UserList(w http.ResponseWriter, r *http.Request)
	UserAdd(w http.ResponseWriter, r *http.Request)

	OrderList(w http.ResponseWriter, r *http.Request)
	OrderAdd(w http.ResponseWriter, r *http.Request)
	OrderUpdate(w http.ResponseWriter, r *http.Request)

	Close() error
}

type exchange struct {
	db db.Database
}

func NewExchange() Exchange {
	db := db.NewDB()
	return &exchange{db: db}
}

func (ex *exchange) OrderList(w http.ResponseWriter, r *http.Request) {
	list := ex.db.ListOrders()

	data, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(data)
}

func (ex *exchange) OrderAdd(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.URL.Query()["user_id"] // Getting user ID should be done via some auth process, passing it manually for now.
	if !ok || len(userID) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var newOrder db.Order

	err := json.NewDecoder(r.Body).Decode(&newOrder)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(userID[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		return
	}

	err = ex.db.AddOrder(db.User{ID: id}, newOrder)
	if err != nil {
		w.WriteHeader(http.StatusConflict)

		data, _ := json.Marshal(newOrder)
		log.Println(err, string(data))
	}
}

func (ex *exchange) OrderUpdate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.URL.Query()["user_id"] // Getting user ID should be done via some auth process, passing it manually for now.
	if !ok || len(userID) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)

	oid, err := strconv.ParseInt(vars["order_id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uid, err := strconv.ParseInt(userID[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newOrder := db.Order{ID: oid, Assigned: uid}

	switch vars["status"] {
	case "done":
		newOrder.Done = time.Now()
	case "reserve":
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ex.db.UpdateOrder(newOrder)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		data, _ := json.Marshal(newOrder)
		log.Println(err, string(data))
	}
}

func (ex *exchange) UserAdd(w http.ResponseWriter, r *http.Request) {
	var newUser db.User

	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = ex.db.AddUser(newUser)
	if err != nil {
		w.WriteHeader(http.StatusConflict)

		data, _ := json.Marshal(newUser)
		log.Println(err, string(data))
	}
}

func (ex *exchange) UserList(w http.ResponseWriter, r *http.Request) {
	list := ex.db.ListUsers()

	data, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(data)
}

func (ex *exchange) Close() error {
	return ex.db.Close()
}
