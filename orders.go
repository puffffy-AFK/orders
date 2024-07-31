package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "strconv"
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
type OrderStore interface {
    Create(order *Order) error
    GetAll() ([]Order, error)
    GetByID(id int) (*Order, error)
    Update(order *Order) error
    Delete(id int) error
}
type SQLiteOrderStore struct {
    db *sql.DB
}

func NewSQLiteOrderStore(db *sql.DB) *SQLiteOrderStore {
    return &SQLiteOrderStore{db: db}
}

func (store *SQLiteOrderStore) Create(order *Order) error {
    order.CreatedAt = time.Now()
    order.UpdatedAt = time.Now()
    stmt, err := store.db.Prepare("INSERT INTO orders (product, count, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?)")
    if err != nil {
        return err
    }
    res, err := stmt.Exec(order.Product, order.Count, order.Status, order.CreatedAt, order.UpdatedAt)
    if err != nil {
        return err
    }
    id, err := res.LastInsertId()
    if err != nil {
        return err
    }
    order.ID = int(id)
    return nil
}

func (store *SQLiteOrderStore) GetAll() ([]Order, error) {
    rows, err := store.db.Query("SELECT id, product, count, status, created_at, updated_at FROM orders")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var orders []Order
    for rows.Next() {
        var order Order
        if err := rows.Scan(&order.ID, &order.Product, &order.Count, &order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
            return nil, err
        }
        orders = append(orders, order)
    }
    return orders, nil
}

func (store *SQLiteOrderStore) GetByID(id int) (*Order, error) {
    var order Order
    row := store.db.QueryRow("SELECT id, product, count, status, created_at, updated_at FROM orders WHERE id = ?", id)
    if err := row.Scan(&order.ID, &order.Product, &order.Count, &order.Status, &order.CreatedAt, &order.UpdatedAt); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &order, nil
}

func (store *SQLiteOrderStore) Update(order *Order) error {
    order.UpdatedAt = time.Now()
    stmt, err := store.db.Prepare("UPDATE orders SET product = ?, count = ?, status = ?, updated_at = ? WHERE id = ?")
    if err != nil {
        return err
    }
    _, err = stmt.Exec(order.Product, order.Count, order.Status, order.UpdatedAt, order.ID)
    if err != nil {
        return err
    }
    return nil
}

func (store *SQLiteOrderStore) Delete(id int) error {
    stmt, err := store.db.Prepare("DELETE FROM orders WHERE id = ?")
    if err != nil {
        return err
    }
    _, err = stmt.Exec(id)
    if err != nil {
        return err
    }
    return nil
}

var store OrderStore

func init() {
    var err error
    db, err := sql.Open("sqlite3", "./orders.db")
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

    store = NewSQLiteOrderStore(db)
}

func createOrder(w http.ResponseWriter, r *http.Request) {
    var order Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if err := store.Create(&order); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

func getOrders(w http.ResponseWriter, r *http.Request) {
    orders, err := store.GetAll()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(orders)
}

func getOrder(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    idStr := params["id"]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    order, err := store.GetByID(id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    if order == nil {
        http.Error(w, "Order not found", http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

func updateOrder(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    idStr := params["id"]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    var order Order
    if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    order.ID = id
    if err := store.Update(&order); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}

func deleteOrder(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    idStr := params["id"]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid order ID", http.StatusBadRequest)
        return
    }

    if err := store.Delete(id); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}