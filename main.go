package main

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"
)

func main() {
    r := mux.NewRouter()

	r.HandleFunc("/api/v1/orders", createOrder).Methods("POST")
    r.HandleFunc("/api/v1/orders", getOrders).Methods("GET")
    r.HandleFunc("/api/v1/orders/{id}", getOrder).Methods("GET")
    r.HandleFunc("/api/v1/orders/{id}", updateOrder).Methods("PUT")
    r.HandleFunc("/api/v1/orders/{id}", deleteOrder).Methods("DELETE")

    log.Fatal(http.ListenAndServe(":8080", r))
}