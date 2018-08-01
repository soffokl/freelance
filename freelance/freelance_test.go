package freelance

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/soffokl/freelance/db"
)

const (
	n = 200
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestFreelance(t *testing.T) {
	ex := NewExchange()
	defer ex.Close()

	r := mux.NewRouter()
	r.HandleFunc("/users/", ex.UserList).Methods(http.MethodGet)
	r.HandleFunc("/users/", ex.UserAdd).Methods(http.MethodPost)

	r.HandleFunc("/orders/", ex.OrderList).Methods(http.MethodGet)
	r.HandleFunc("/orders/", ex.OrderAdd).Methods(http.MethodPost)
	r.HandleFunc("/orders/{order_id}/{status}", ex.OrderUpdate).Methods(http.MethodPut)

	ts := httptest.NewServer(r)

	createUsers("test", ts.URL+"/users/", t)
	balance0 := listUsers(ts.URL+"/users/", t)

	if balance0 != n*1000 {
		t.Fatalf("wrong total users balance, expected %d, got %f", n*1000, balance0)
	}

	createOrders("torder", ts.URL+"/orders/", t)
	listOrders(ts.URL+"/orders/", t)

	balance1 := listUsers(ts.URL+"/users/", t)

	if balance1 != n*1000-n*10 {
		t.Fatalf("wrong total users balance after orders placing, expected %d, got %f", n*1000-n*10, balance1)
	}

	doneOrders(ts.URL+"/orders/", t)

	balance2 := listUsers(ts.URL+"/users/", t)
	if balance0 != balance2 {
		t.Fatalf("wrong total users balance after orders done, expected %f, got %f", balance0, balance2)
	}

}

func doneOrders(url string, t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 1; i <= n; i++ {
		go func(i int) {
			req, err := http.NewRequest(http.MethodPut, url+strconv.Itoa(i)+"/done", nil)
			if err != nil {
				t.Errorf("failed to create http request")
			}

			q := req.URL.Query()
			q.Add("user_id", strconv.Itoa(rand.Intn(n)+1))
			req.URL.RawQuery = q.Encode()

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("failed to execute http request")
				return
			}

			if res.StatusCode != http.StatusOK {
				t.Errorf("bad status code: %d", res.StatusCode)
			}

			res.Body.Close()
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func createOrders(name, url string, t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 1; i <= n; i++ {
		func(i int) {
			order := db.Order{Title: name + strconv.Itoa(i), Fee: 10}

			data, err := json.Marshal(order)
			if err != nil {
				t.Errorf("failed to marshal order json")
			}

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			if err != nil {
				t.Errorf("failed to create http request")
			}

			q := req.URL.Query()
			q.Add("user_id", strconv.Itoa(i))
			req.URL.RawQuery = q.Encode()

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("failed to execute http request")
				return
			}

			if res.StatusCode != http.StatusOK {
				t.Errorf("bad status code: %d", res.StatusCode)
			}

			res.Body.Close()
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func listOrders(url string, t *testing.T) {
	var orders []db.Order

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Errorf("failed to create http request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("failed to execute http request")
	}

	err = json.NewDecoder(res.Body).Decode(&orders)
	if err != nil {
		t.Errorf("failed to decode users")
	}

	if len(orders) != n {
		t.Fatalf("number of returned uses not equal to %d, received %d", n, len(orders))
	}
}

func listUsers(url string, t *testing.T) float64 {
	var users []db.User

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Errorf("failed to create http request")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("failed to execute http request")
	}

	err = json.NewDecoder(res.Body).Decode(&users)
	if err != nil {
		t.Errorf("failed to decode users")
	}

	if len(users) != n {
		t.Fatalf("number of returned uses not equal to %d, received %d", n, len(users))
	}

	sum := 0.0
	for i := range users {
		sum += users[i].Balance
	}
	return sum
}

func createUsers(name, url string, t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			user := db.User{Name: name + strconv.Itoa(i), Balance: 1000}

			data, err := json.Marshal(user)
			if err != nil {
				t.Errorf("failed to marshal user json")
			}

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			if err != nil {
				t.Errorf("failed to create http request")
			}

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("failed to execute http request")
				return
			}

			if res.StatusCode != http.StatusOK {
				t.Errorf("bad status code: %d", res.StatusCode)
			}

			res.Body.Close()
			wg.Done()
		}(i)
	}

	wg.Wait()
}
