package main

import (
	"context"
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

	Get(server, "/users", getUsers)
	Post(server, "/users", postUsers)

	server.Mux.Handle("GET /groups", Get2(server, getGroups))

	apiMux := http.NewServeMux()
	server.Mux.Handle("/api/", http.StripPrefix("/api", apiMux))

	apiMux.Handle("/users", Get2(server, apiGetUsers))
	apiMux.Handle("/users/{user}", Get2(server, apiGetUsersUser))

	apiMux.Handle("GET /groups", HandleOutFunc(server, apiGetGroups))

	err = http.ListenAndServe(":9988", server)

	return err
}

type User struct {
	Name  string `json:"name"`
	Login string `json:"login"`
}

func getUsers(r *Request) ([]User, error) {
	var users []User
	users = append(users, User{Name: "John"})
	return users, nil
}

func postUsers(r *Request, newUser User) (User, error) {
	user := User{Name: "New John"}
	return user, nil
}

type Group struct {
	Name string `json:"name"`
}

func getGroups(ctx context.Context, s *Server, r *Request) ([]Group, error) {
	var groups []Group
	groups = append(groups, Group{Name: "sudoers"})
	return groups, nil
}

func apiGetUsers(ctx context.Context, s *Server, r *Request) ([]User, error) {
	var users []User
	users = append(users, User{Name: "John (API)"})
	return users, nil
}

func apiGetUsersUser(ctx context.Context, s *Server, r *Request) (User, error) {
	user := User{
		Name:  "John Doe",
		Login: r.PathValue("user"),
	}
	return user, nil
}

func apiGetGroups(ctx context.Context, s *Server, r *Request) ([]Group, error) {
	var groups []Group
	groups = append(groups, Group{Name: "sudoers"})
	return groups, nil
}
