package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var err error

	log.Printf("api-rest (%d args)\n", len(args))

	server := NewServer()

	server.AddMiddleware(logger)

	server.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello.")
	})

	apiMux := http.NewServeMux()
	server.Mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	apiMux.Handle("/users", Handle(apiGetUsers))
	apiMux.Handle("/users/{user}", Handle(apiGetUsersUser))

	apiMux.Handle("POST /users", Handle(apiPostUsers))

	apiMux.Handle("GET /groups", Handle(apiGetGroups))

	// apiMux.Handle("GET /roles", Handle(apiGetRoles))
	// apiMux.Handle("GET /roles/{role}", Handle(apiGetRolesRole))
	// apiMux.Handle("POST /roles", Handle(apiPostRoles, OnlyRoot))

	apiMux.Handle("GET /roles", Handle(apiGetRoles))
	// apiMux.Handle("GET /roles/{role}", Handle(apiGetRolesRole))
	apiMux.Handle("POST /roles", Handle(apiPostRoles, OnlyRoot))

	err = http.ListenAndServe(":9988", server)

	return err
}

type User struct {
	Name  string `json:"name"`
	Login string `json:"login"`
}

type Group struct {
	Name string `json:"name"`
}

func apiGetUsers(r *Request, _ None) ([]User, error) {
	var users []User
	users = append(users, User{Name: "John (API)"})
	return nil, sql.ErrNoRows
	return users, nil
}

func apiGetUsersUser(r *Request, _ None) (User, error) {
	user := User{
		Name:  fmt.Sprintf("John Doe"),
		Login: r.PathValue("user"),
	}
	return user, nil
}

func apiGetGroups(r *Request, _ None) ([]Group, error) {
	var groups []Group
	groups = append(groups, Group{Name: "sudoers"})
	return groups, nil
}

func apiPostUsers(r *Request, newUser User) (User, error) {
	return newUser, nil
}

type Role struct {
	Name string
}

func apiGetRoles(r *Request, _ None) ([]Role, error) {
	return []Role{
		Role{Name: "Alpha"},
		Role{Name: "Bravo"},
	}, nil
}

func apiPostRoles(r *Request, role Role) (Role, error) {
	return Role{}, nil
}
