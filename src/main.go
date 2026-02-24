package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"templ-ui-kit/src/i18n"
	"templ-ui-kit/src/layouts"
)

// ── In-memory store ───────────────────────────────────────────────────────────

const pageSize = 5

var (
	todoStore  []layouts.TodoItem
	todoMu     sync.Mutex
	todoNextID = 1
)

func init() {
	seeds := []struct {
		title    string
		priority layouts.TodoPriority
	}{
		{"Buy groceries", layouts.TodoPriorityLow},
		{"Finish project report", layouts.TodoPriorityHigh},
		{"Call the doctor", layouts.TodoPriorityMedium},
	}
	for _, s := range seeds {
		todoStore = append(todoStore, layouts.TodoItem{
			ID:       todoNextID,
			Title:    s.title,
			Priority: s.priority,
		})
		todoNextID++
	}
}

func allTodos() []layouts.TodoItem {
	todoMu.Lock()
	defer todoMu.Unlock()
	cp := make([]layouts.TodoItem, len(todoStore))
	copy(cp, todoStore)
	return cp
}

func pagedTodos(page int) ([]layouts.TodoItem, int) {
	all := allTodos()
	total := len(all)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	return all[start:end], totalPages
}

func addTodo(data layouts.TodoFormData) {
	todoMu.Lock()
	defer todoMu.Unlock()
	todoStore = append(todoStore, layouts.TodoItem{
		ID:       todoNextID,
		Title:    data.Parsed.Title,
		Priority: layouts.TodoPriority(data.Parsed.Priority),
	})
	todoNextID++
}

func deleteTodo(id int) {
	todoMu.Lock()
	defer todoMu.Unlock()
	for i, t := range todoStore {
		if t.ID == id {
			todoStore = append(todoStore[:i], todoStore[i+1:]...)
			return
		}
	}
}

func toggleTodo(id int) {
	todoMu.Lock()
	defer todoMu.Unlock()
	for i, t := range todoStore {
		if t.ID == id {
			todoStore[i].Done = !t.Done
			return
		}
	}
}

// ── Form parsing ──────────────────────────────────────────────────────────────

func parseTodoForm(r *http.Request) layouts.TodoFormData {
	ctx := r.Context()
	lang := i18n.GetLanguage(ctx)
	tr := i18n.GetTranslator()

	data := layouts.NewTodoFormData()

	if err := r.ParseForm(); err != nil {
		data.Problems["_generic"] = []string{err.Error()}
		return data
	}

	data.Raw.Title = r.PostFormValue("title")
	data.Raw.Priority = r.PostFormValue("priority")

	data.Parsed.Title = strings.TrimSpace(data.Raw.Title)
	if len(data.Parsed.Title) == 0 {
		data.Problems["title"] = []string{tr.T(lang, "required")}
	} else if len(data.Parsed.Title) < 3 {
		data.Problems["title"] = []string{tr.Tf(lang, "min", 3)}
	}

	priorityStr := strings.TrimSpace(data.Raw.Priority)
	if len(priorityStr) == 0 {
		data.Problems["priority"] = []string{tr.T(lang, "required")}
	} else {
		p, err := strconv.Atoi(priorityStr)
		if err != nil || p < 1 || p > 3 {
			data.Problems["priority"] = []string{tr.T(lang, "invalid")}
		} else {
			data.Parsed.Priority = p
		}
	}

	data.Valid = len(data.Problems) == 0
	return data
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func handleIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := layouts.TestLayoutI18n(ctx).Render(ctx, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTodos(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleTodoIndex(w, r)
	case http.MethodPost:
		handleTodoCreate(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTodoIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	todos, totalPages := pagedTodos(page)
	data := layouts.NewTodoFormData()
	if err := layouts.TodoLayout(ctx, todos, data, page, totalPages).Render(ctx, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTodoCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := parseTodoForm(r)

	if !data.Valid {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := layouts.TodoForm(ctx, data).Render(ctx, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	addTodo(data)

	todos, totalPages := pagedTodos(1)
	if err := layouts.TodoCreateSuccess(ctx, todos, 1, totalPages).Render(ctx, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTodoByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		handleTodoDelete(w, r)
	case http.MethodPatch:
		// /todos/{id}/toggle — sprawdź suffix
		handleTodoToggle(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTodoDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	deleteTodo(id)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	todos, totalPages := pagedTodos(page)
	if err := layouts.TodoTableInner(ctx, todos, page, totalPages).Render(ctx, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleTodoToggle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	toggleTodo(id)
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	todos, totalPages := pagedTodos(page)
	if err := layouts.TodoTableInner(ctx, todos, page, totalPages).Render(ctx, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleSetLanguage(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "en"
	}
	i18n.SetLanguageCookie(w, lang)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = i18n.GetTranslator()

	mux := http.NewServeMux()
	lang := i18n.LanguageMiddleware

	// Components demo
	mux.Handle("/", lang(http.HandlerFunc(handleIndex)))
	mux.HandleFunc("/set-language", handleSetLanguage)

	// Todo — bez metody w ścieżce, dispatch ręcznie
	mux.Handle("/todos", lang(http.HandlerFunc(handleTodos)))
	mux.Handle("/todos/{id}", lang(http.HandlerFunc(handleTodoByID)))
	mux.Handle("/todos/{id}/toggle", lang(http.HandlerFunc(handleTodoToggle)))

	// Static
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))))

	log.Println("Server  → http://localhost:8080")
	log.Println("Todos   → http://localhost:8080/todos")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
