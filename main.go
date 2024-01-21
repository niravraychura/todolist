// main.go
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

func init() {
	var err error
	database, err = sql.Open("sqlite3", "./todo.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			description TEXT,
			completed BOOLEAN
		);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("SELECT id, description, completed FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Description, &task.Completed)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, task)
	}

	renderTemplate(w, "index", tasks)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		description := r.FormValue("description")
		_, err := database.Exec("INSERT INTO tasks (description, completed) VALUES (?, false)", description)
		if err != nil {
			log.Fatal(err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	renderTemplate(w, "add", nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmplFiles := []string{"templates/" + tmpl + ".html", "templates/layout.html"}
	t, err := template.ParseFiles(tmplFiles...)
	if err != nil {
		log.Fatal(err)
	}

	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		log.Fatal(err)
	}
}

type Task struct {
	ID          int
	Description string
	Completed   bool
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addTaskHandler)
	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
