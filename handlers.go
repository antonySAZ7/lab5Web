package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func handleRequest(conn net.Conn, db *sql.DB) {
	defer conn.Close()

	buffer := make([]byte, 4096)
	n, _ := conn.Read(buffer)
	request := string(buffer[:n])

	lines := strings.Split(request, "\r\n")
	firstLine := strings.Split(lines[0], " ")

	method := firstLine[0]
	fullPath := firstLine[1]

	parts := strings.Split(fullPath, "?")
	path := parts[0]

	params := url.Values{}
	if len(parts) > 1 {
		params, _ = url.ParseQuery(parts[1])
	}

	if strings.HasPrefix(path, "/static/") {
		serveStatic(conn, path)
		return
	}

	if method == "GET" && path == "/" {
		handleHome(conn, db, params)
		return
	}

	if method == "GET" && path == "/create" {
		handleCreateForm(conn)
		return
	}

	if method == "POST" && path == "/create" {
		handleCreate(conn, request, db)
		return
	}

	if method == "POST" && path == "/update" {
		id := params.Get("id")
		db.Exec(`UPDATE series 
		         SET current_episode = current_episode + 1 
		         WHERE id = ? AND current_episode < total_episodes`, id)
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nok"))
		return
	}

	if method == "POST" && path == "/decrement" {
		id := params.Get("id")
		db.Exec(`UPDATE series 
		         SET current_episode = current_episode - 1 
		         WHERE id = ? AND current_episode > 1`, id)
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nok"))
		return
	}
}

func handleHome(conn net.Conn, db *sql.DB, params url.Values) {

	sort := params.Get("sort")
	orderBy := "id"

	if sort == "name" {
		orderBy = "name"
	}
	if sort == "current" {
		orderBy = "current_episode"
	}

	rows, _ := db.Query("SELECT id, name, current_episode, total_episodes FROM series ORDER BY " + orderBy)
	defer rows.Close()

	html := `
<!DOCTYPE html>
<html>
<head>
<link rel="stylesheet" href="/static/style.css">
<script src="/static/app.js"></script>
<title>tracker de series que ando viendo</title>
</head>
<body>

<h1>tracker series</h1>

<a href="/create">Agregar nueva serie</a>

<table>
<tr>
<th>#</th>
<th><a href="/?sort=name">nombre</a></th>
<th><a href="/?sort=current">episodio actual</a></th>
<th>total de episodios</th>
<th>progreso</th>
<th>accion</th>
</tr>
`

	for rows.Next() {
		var id int
		var name string
		var current int
		var total int

		rows.Scan(&id, &name, &current, &total)

		percentage := (current * 100) / total

		rowClass := ""
		if current == total {
			rowClass = "complete"
		}

		html += fmt.Sprintf(`
<tr class="%s">
<td>%d</td>
<td>%s</td>
<td>%d</td>
<td>%d</td>
<td>
<div class="progress">
<div class="bar" style="width:%d%%"></div>
</div>
</td>
<td>
<button onclick="prevEpisode(%d)">-1</button>
<button onclick="nextEpisode(%d)">+1</button>
</td>
</tr>
`, rowClass, id, name, current, total, percentage, id, id)
	}

	html += `
</table>
</body>
</html>
`

	response := "HTTP/1.1 200 OK\r\n"
	response += "Content-Type: text/html; charset=utf-8\r\n"
	response += "Content-Length: " + strconv.Itoa(len(html)) + "\r\n"
	response += "\r\n"
	response += html

	conn.Write([]byte(response))
}

func handleCreateForm(conn net.Conn) {

	html := `
<!DOCTYPE html>
<html>
<head>
<link rel="stylesheet" href="/static/style.css">
<title>Crear Serie</title>
</head>
<body>

<h1>Agregar Serie</h1>

<form method="POST" action="/create">
<input type="text" name="series_name" placeholder="Nombre" required>
<input type="number" name="current_episode" min="1" value="1" required>
<input type="number" name="total_episodes" min="1" required>
<button type="submit">Crear</button>
</form>

<a href="/">Volver</a>

</body>
</html>
`

	response := "HTTP/1.1 200 OK\r\n"
	response += "Content-Type: text/html\r\n\r\n"
	response += html

	conn.Write([]byte(response))
}

func handleCreate(conn net.Conn, request string, db *sql.DB) {

	bodyIndex := strings.Index(request, "\r\n\r\n")
	body := request[bodyIndex+4:]

	values, _ := url.ParseQuery(body)

	name := values.Get("series_name")
	current := values.Get("current_episode")
	total := values.Get("total_episodes")

	db.Exec("INSERT INTO series (name, current_episode, total_episodes) VALUES (?, ?, ?)",
		name, current, total)

	response := "HTTP/1.1 303 See Other\r\n"
	response += "Location: /\r\n\r\n"

	conn.Write([]byte(response))
}

func serveStatic(conn net.Conn, path string) {
	filePath := "." + path
	data, err := os.ReadFile(filePath)
	if err != nil {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
	}

	contentType := "text/plain"
	if strings.HasSuffix(path, ".css") {
		contentType = "text/css"
	}
	if strings.HasSuffix(path, ".js") {
		contentType = "application/javascript"
	}

	response := "HTTP/1.1 200 OK\r\n"
	response += "Content-Type: " + contentType + "\r\n\r\n"

	conn.Write([]byte(response))
	conn.Write(data)
}
