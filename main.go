package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

type Contact struct {
	ID    int
	Name  string
	Phone string
	Email string
}

var db *sql.DB

func main() {
	var err error

	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=phonebook sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to database!")

	http.HandleFunc("/", showContacts)
	http.HandleFunc("/create", showCreateForm)
	http.HandleFunc("/add", addContact)
	http.HandleFunc("/edit", showEditForm)
	http.HandleFunc("/update", updateContact)

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func showContacts(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, name, phone, email FROM contacts")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Email)
		if err != nil {
			http.Error(w, "Failed to scan row", http.StatusInternalServerError)
			return
		}
		contacts = append(contacts, c)
	}

	tmpl, err := template.ParseFiles("contacts.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, contacts)
}

func showCreateForm(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("create.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func addContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	phone := r.FormValue("phone")
	email := r.FormValue("email")

	_, err := db.Exec("INSERT INTO contacts (name, phone, email) VALUES ($1, $2, $3)", name, phone, email)
	if err != nil {
		http.Error(w, "Failed to insert contact", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showEditForm(w http.ResponseWriter, r *http.Request) {
	indexStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(indexStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var c Contact
	err = db.QueryRow("SELECT id, name, phone, email FROM contacts WHERE id = $1", id).Scan(&c.ID, &c.Name, &c.Phone, &c.Email)
	if err != nil {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("edit.html")
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, c)
}

func updateContact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	idStr := r.FormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	phone := r.FormValue("phone")
	email := r.FormValue("email")

	_, err = db.Exec("UPDATE contacts SET name=$1, phone=$2, email=$3 WHERE id=$4",
		name, phone, email, id)
	if err != nil {
		http.Error(w, "Failed to update contact", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
