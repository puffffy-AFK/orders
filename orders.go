package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    _ "github.com/mattn/go-sqlite3"
)

type Order struct {
    ID        int       `json:"id"`
    Product   string    `json:"product"`
    Count     int       `json:"count"`
    Status    string    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

var db *sql.DB

func init() {
    var err error
    db, err = sql.Open("sqlite3", "./orders.db")
    if err != nil {
        log.Fatal(err)
    }

    createTable := `
    CREATE TABLE IF NOT EXISTS orders (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        product TEXT,
        count INTEGER,
        status TEXT,
        created_at DATETIME,
        updated_at DATETIME
    );`
    _, err = db.Exec(createTable)
    if err != nil {
        log.Fatal(err)
    }
}

func createOrder(w http.ResponseWriter, r *http.Request) {
    var order Order
    json.NewDecoder(r.Body).Decode(&order)
    order.CreatedAt = time.Now()
    order.UpdatedAt = time.Now()

    stmt, err := db.Prepare("INSERT INTO orders (product, count, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    res, err := stmt.Exec(order.Product, order.Count, order.Status, order.CreatedAt, order.UpdatedAt)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    id, _ := res.LastInsertId()
    order.ID = int(id)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

func getOrders(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, product, count, status, created_at, updated_at FROM orders")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var order Order
        if err := rows.Scan(&order.ID, &order.Product, &order.Count, &order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        orders = append(orders, order)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(orders)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id := params["id"]

    var order Order
    row := db.QueryRow("SELECT id, product, count, status, created_at, updated_at FROM orders WHERE id = ?", id)
    if err := row.Scan(&order.ID, &order.Product, &order.Count, &order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Order not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id := params["id"]

    var order Order
    json.NewDecoder(r.Body).Decode(&order)
    order.UpdatedAt = time.Now()

    stmt, err := db.Prepare("UPDATE orders SET product = ?, count = ?, status = ?, updated_at = ? WHERE id = ?")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    _, err = stmt.Exec(order.Product, order.Count, order.Status, order.UpdatedAt, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

func deleteOrder(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    id := params["id"]

    stmt, err := db.Prepare("DELETE FROM orders WHERE id = ?")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    _, err = stmt.Exec(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}