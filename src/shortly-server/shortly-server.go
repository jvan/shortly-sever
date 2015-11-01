package main

import (
  "net/http"
  "github.com/labstack/echo"
  mw "github.com/labstack/echo/middleware"

  "database/sql"
  _ "github.com/mattn/go-sqlite3"

  "github.com/rs/cors"

  "strconv"

  "shortly"
)

var db *sql.DB  // database handle


//----- Data Structures -----------------------------------

type User struct {
  Id int        `json:"id"`
  Name string   `json:"name"`
	Links []int   `json:"links"`
}

type Link struct {
  Id int        `json:"id"`
  Url string    `json:"url"`
  Key string    `json:"key"`
}


type UserArray []User

//----- Routes --------------------------------------------

func ping(c *echo.Context) error {
  return c.String(http.StatusOK, "Ok")
}

func encode_value(c *echo.Context) error {
  // Convert the `val` parameter to an integer.
  value := c.Param("val")
  i, _ := strconv.Atoi(value)

  // Return the base-62 encoded value.
  return c.String(http.StatusOK, shortly.Encode(i))
}

func decode_value(c *echo.Context) error {
  // Decode the base-62 `val` parameter.
  value := c.Param("val")
  i := shortly.Decode(value)

  // Convert the base-10 integer to a string and return.
  s := strconv.Itoa(i)
  return c.String(http.StatusOK, s)
}

func get_all_users(c *echo.Context) error {
  // Query the database for all user records.
  rows, _ := db.Query("SELECT id, username FROM users")

  users := UserArray{}

  // For each row in the result create a new user object and add it to the
  // users array.
  for rows.Next() {
    var user_id int
    var username string
    rows.Scan(&user_id, &username)

    new_user := User{Id: user_id, Name: username}
    users = append(users, new_user)
  }

  // For ember-data compatibility we need to construct a response that 
  // contains a `users` object.
  type UserResponse struct {
    Users UserArray `json:"users"`
  }

  response := UserResponse{Users: users}

  return c.JSON(http.StatusOK, response)
}

func get_user(c *echo.Context) error {
	// Query the database for the specified user.
	id := c.Param("user_id")
	row, _ := db.Query("SELECT id, username FROM users WHERE id=$1 LIMIT 1", id)

	// TODO: Return NOT FOUND if no rows are returned.

	// Get first (and only) record.
	row.Next()

	var user_id int
	var username string

	row.Scan(&user_id, &username)

	// Query the database for all links associated with the user.
	rows, _ := db.Query("SELECT id FROM links WHERE user_id=$1", user_id)

	links := []int{}

	for rows.Next() {
		var link_id int
		rows.Scan(&link_id)
		links = append(links, link_id)
	}

	// Initialize a user object.
	user := User{Id: user_id, Name: username, Links: links}

	// For ember-data compatibility we need to construct a response that
	// contains a `user` object.
	type UserResponse struct {
		User User `json:"user"`
	}

	response := UserResponse{User: user}
	return c.JSON(http.StatusOK, response)
}

func get_link(c *echo.Context) error {
	// Query the database for the specified user.
	id := c.Param("link_id")
	row, _ := db.Query("SELECT id, url FROM links WHERE id=$1 LIMIT 1", id)

	// TODO: Return NOT FOUND if no rows are returned.

	// Get the first (and only) record
	row.Next()

	var uid int
	var url string

	row.Scan(&uid, &url)

	// Initialize a link object. Use the `Encode` function to base-62
	// encode the link id.
	link := Link{Id: uid, Url: url, Key: shortly.Encode(uid)}

	// For ember-data compatibility we need to construct a response that
	// contains a `user` object.
	type LinkResponse struct {
		Link Link `json:"link"`
	}

	response := LinkResponse{Link: link}

	return c.JSON(http.StatusOK, response)
}


func main() {
  e := echo.New()

  e.Use(mw.Logger())
  e.Use(mw.Recover())

  // Enable CORS.
  e.Use(cors.Default().Handler)

  // Initialize the database connection.
  db, _ = sql.Open("sqlite3", "test.db")

  e.Get("/ping", ping)

  e.Get("/encode/:val", encode_value)
  e.Get("/decode/:val", decode_value)

  e.Get("/users", get_all_users)
	e.Get("/users/:user_id", get_user)

	e.Get("/links/:link_id", get_link)

  e.Run(":1323")
}
