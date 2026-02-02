package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<script src="https://cdn.tailwindcss.com"></script>
</head>
<body>
	<nav class="bg-blue-500 text-white px-6 py-4">
		<a href="/" class="hover:opacity-80">Home</a>
	</nav>
	<div class="p-8">
		<h1 class="text-4xl font-bold">Hello from Layout</h1>
	</div>
</body>
</html>`))
	})

	fmt.Println("http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
