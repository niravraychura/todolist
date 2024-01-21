// main.go
package main

import (
	"database/sql"
	"encoding/json"
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
func deletePageHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := database.Query("SELECT id, description FROM tasks")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Description)
		if err != nil {
			log.Fatal(err)
		}
		tasks = append(tasks, task)
	}

	renderTemplate(w, "delete", tasks)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var tasksToDelete []string

		if err := json.NewDecoder(r.Body).Decode(&tasksToDelete); err != nil {
			http.Error(w, "Error decoding JSON data", http.StatusBadRequest)
			return
		}

		if len(tasksToDelete) == 0 {
			http.Error(w, "No tasks selected for deletion", http.StatusBadRequest)
			return
		}

		// Loop through tasksToDelete and delete each task
		for _, id := range tasksToDelete {
			_, err := database.Exec("DELETE FROM tasks WHERE id = ?", id)
			if err != nil {
				log.Printf("Error deleting task with ID %s: %s", id, err.Error())
				http.Error(w, fmt.Sprintf("Error deleting task with ID %s", id), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
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
	tmplParsed, err := template.ParseFiles(tmplFiles...)
	if err != nil {
		log.Fatal(err)
	}

	err = tmplParsed.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Fatal(err)
	}
}

func updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		completed := r.FormValue("completed") == "true"

		_, err := database.Exec("UPDATE tasks SET completed = ? WHERE id = ?", completed, id)
		if err != nil {
			log.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
}

type Task struct {
	ID          int
	Description string
	Completed   bool
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addTaskHandler)
	http.HandleFunc("/update", updateTaskStatusHandler)
	http.HandleFunc("/delete", deleteTaskHandler)
	http.HandleFunc("/delete-page", deletePageHandler)
	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}
