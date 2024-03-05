package main

import (
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

	err = http.ListenAndServe(":9988", server)

	return err
}

type User struct {
	Name string `json:"name"`
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
