package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"time"
	"log"
	"os"
	"github.com/gorilla/mux"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
)

type Redirection struct {
	Pk string
	Url string
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func index(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("Sending index.html")
	http.ServeFile(w, req, "index.html")
}

func generateKey() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
 
    s := make([]rune, 6)
    for i := range s {
        s[i] = letters[rand.Intn(len(letters))]
    }
    return string(s)
}

func main() {
	db_endpoint := os.Getenv("DB_ENDPOINT")

	db, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(db_endpoint))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = db.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	urlCollection := db.Database("test").Collection("urls")

	r := mux.NewRouter()

	// Handle GET requests to /l/{id}
	r.HandleFunc("/l/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			fmt.Println("Received GET to " + r.URL.String())
			
			vars := mux.Vars(r)
			id, ok := vars["id"]
			if !ok {
				fmt.Println("id is missing in parameters")
				http.NotFound(w, r)
			}

			var result Redirection
			urlCollection.FindOne(context.TODO(), bson.D{primitive.E{Key: "pk", Value: id}}).Decode(&result)

			fmt.Println("redirectUrl: " + result.Url)
	
			http.Redirect(w, r, result.Url, http.StatusFound)
		}
	})
	
	// Handle PUT requests to /l
	r.HandleFunc("/l", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Println("Received GET to /l")
		case "POST":
			// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}
			fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
			
			url := r.FormValue("url")
			key := generateKey()
			
			_,err := urlCollection.InsertOne(context.TODO(), bson.D{{Key: "pk", Value: key},{Key: "url", Value: url}})
			if err != nil {
				log.Fatalf("Error happened in DB insertion. Err: %s", err)
			}

			resp := make(map[string]string)
			resp["message"] = "Status Created"
			resp["key"] = key
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonResp)
		default:
			fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
		}
	})

	r.HandleFunc("/", index)
	r.HandleFunc("/headers", headers)

	port := os.Getenv("PORT")
	log.Println("Starting server on port " + port)

	// Ping the primary
	if err := db.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	s := &http.Server{
		Addr:           ":" + port,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}