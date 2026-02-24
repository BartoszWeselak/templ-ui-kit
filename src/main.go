package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"templ-ui-kit/src/apperrors"
	"templ-ui-kit/src/i18n"
	"templ-ui-kit/src/layouts"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func parseTodoForm(c echo.Context) layouts.TodoFormData {
	ctx := c.Request().Context()
	lang := i18n.GetLanguage(ctx)
	tr := i18n.GetTranslator()

	data := layouts.NewTodoFormData()

	if err := c.Request().ParseForm(); err != nil {
		data.Problems["_generic"] = []string{err.Error()}
		return data
	}

	data.Raw.Title = c.FormValue("title")
	data.Raw.Priority = c.FormValue("priority")

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

// ── Middleware ────────────────────────────────────────────────────────────────

func i18nMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		lang := i18n.DetectLanguage(c.Request())
		ctx := i18n.WithLanguage(c.Request().Context(), lang)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func handleIndex(c echo.Context) error {
	ctx := c.Request().Context()
	return apperrors.RenderTempl(c, http.StatusOK, layouts.TestLayoutI18n(ctx))
}

func handleTodoIndex(c echo.Context) error {
	ctx := c.Request().Context()
	page, _ := strconv.Atoi(c.QueryParam("page"))
	todos, totalPages := pagedTodos(page)
	data := layouts.NewTodoFormData()
	return apperrors.RenderTempl(c, http.StatusOK, layouts.TodoLayout(ctx, todos, data, page, totalPages))
}

func handleTodoCreate(c echo.Context) error {
	ctx := c.Request().Context()
	data := parseTodoForm(c)

	if !data.Valid {
		return apperrors.RenderTempl(c, http.StatusUnprocessableEntity, layouts.TodoForm(ctx, data))
	}

	addTodo(data)
	todos, totalPages := pagedTodos(1)
	return apperrors.RenderTempl(c, http.StatusOK, layouts.TodoCreateSuccess(ctx, todos, 1, totalPages))
}

func handleTodoDelete(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error.bad_request")
	}
	deleteTodo(id)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	todos, totalPages := pagedTodos(page)
	return apperrors.RenderTempl(c, http.StatusOK, layouts.TodoTableInner(ctx, todos, page, totalPages))
}

func handleTodoToggle(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "error.bad_request")
	}
	toggleTodo(id)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	todos, totalPages := pagedTodos(page)
	return apperrors.RenderTempl(c, http.StatusOK, layouts.TodoTableInner(ctx, todos, page, totalPages))
}

func handleSetLanguage(c echo.Context) error {
	lang := c.QueryParam("lang")
	if lang == "" {
		lang = "en"
	}
	i18n.SetLanguageCookie(c.Response().Writer, lang)
	return c.Redirect(http.StatusSeeOther, c.Request().Referer())
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = i18n.GetTranslator()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(i18nMiddleware)
	e.HTTPErrorHandler = apperrors.Handler

	e.Static("/static", "src/static")

	e.GET("/", handleIndex)
	e.GET("/set-language", handleSetLanguage)

	e.GET("/todos", handleTodoIndex)
	e.POST("/todos", handleTodoCreate)
	e.DELETE("/todos/:id", handleTodoDelete)
	e.PATCH("/todos/:id/toggle", handleTodoToggle)

	e.Logger.Fatal(e.Start(":8080"))
}
