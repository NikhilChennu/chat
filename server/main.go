package main

import (
	dbmodule "chat/server/db"
	"chat/server/jwt"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var port = flag.String("port", "9000", "port")

var (
	username    string
	password    string
	dbName      string
	dbHost      string
	tokenSecret string
	dbport      string
)

type contextUserKey string

const userId contextUserKey = "user"

func init() {

	e := godotenv.Load() //Load .env file
	if e != nil {
		fmt.Print(e)
	}

	username = os.Getenv("db_user")
	password = os.Getenv("db_pass")
	dbName = os.Getenv("db_name")
	dbHost = os.Getenv("db_host")
	dbport = os.Getenv("db_port")
	tokenSecret = os.Getenv("secret_key")

}

func createUser(db *dbmodule.DBObject, w http.ResponseWriter, r *http.Request) {

	account := &dbmodule.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid input!"))
	}

	err = db.AddAccount(account)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed account creation!"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))

}

func validateMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X_Auth")
		if token == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		_, mobile, err := jwt.ValidateToken(token, tokenSecret)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("invalid token!"))
			return
		}

		fmt.Println("Mobile bumber ", *mobile)

		ctx := context.WithValue(r.Context(), userId, *mobile)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func loginUser(db *dbmodule.DBObject, w http.ResponseWriter, r *http.Request) {

	account := &dbmodule.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid input!"))
	}

	fmt.Printf("%+v ", account)

	resp, err := db.GetAccount(account.Mobile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed account creation!User does not exists"))
	}
	if resp == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("User does not exists"))
	} else {

		fmt.Printf("%s %s\n ", resp.UserID, resp.Mobile)

		token := jwt.GetToken(resp.UserID, resp.Mobile, tokenSecret)
		fmt.Printf("%s\n", token)
		w.Header().Set("X_Auth", token)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	}
}

func main() {

	var wg sync.WaitGroup

	hub := newHub()

	db, err := dbmodule.GetDBObject(dbHost, dbport, username, password, dbName)
	if err != nil {
		fmt.Println("Failed DB creation", err.Error())
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())

	go hub.run()
	router := mux.NewRouter()

	router.Handle("/create", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createUser(db, w, r)
	})).Methods("POST")

	router.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loginUser(db, w, r)
	})).Methods("POST")

	router.Handle("/chat", validateMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveWs(ctx, &wg, hub, w, r)
	})))

	server := &http.Server{Addr: ":" + *port, Handler: router}

	wg.Add(1)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
		wg.Done()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	cancel()
	wg.Wait()
}
