package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type User struct {
	UserID   uint16 `yaml:"userID"`
	Password string `yaml:"password"`
	Usage    uint64 `yaml:"usage"`
}

func RunUserManager() {
	fmt.Println("Welcome to the user management tool")
	for {
		fmt.Println("Please select an option:")
		fmt.Println("1. Create a new user")
		fmt.Println("2. Remove a new user")
		fmt.Println("3. Exit")

		var choice int
		fmt.Print("Choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			fmt.Print("Enter a new password: ")
			password, err := getInput()
			if err != nil {
				log.Fatal("Error reading password: ", err)
			}

			userID := generateUserID()
			hashedPassword := HashSha256(password)
			fmt.Println(hashedPassword)
			if err != nil {
				log.Fatal("Error hashing password: ", err)
			}

			newUser := User{
				UserID:   userID,
				Password: string(hashedPassword),
				Usage:    0,
			}

			saveUser(newUser)

			fmt.Printf("User created with userID %d\n", userID)
		case 2:
			fmt.Print("Enter User ID to remove: ")
			idStr, err := getInput()
			if err != nil {
				log.Fatal("Error reading User ID: ", err)
			}
			id64, err := strconv.ParseUint(idStr, 10, 16)
			if err != nil {
				fmt.Println("Error converting string to uint16:", err)
				return
			}
			id := uint16(id64)
			err = removeUserFromYamlFile(id)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("User removed.")
		case 3:
			fmt.Println("Exiting...")
			os.Exit(0)

		default:
			fmt.Println("Invalid choice, please try again")
		}
	}
}

func getInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	password, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(password), nil
}

func generateUserID() uint16 {
	var userID uint16
	rand.Seed(time.Now().UnixNano())
	userID = uint16(rand.Intn(1 << 16))
	return userID
}

// TODO make it generic for every user database e.g. sql, json
func saveUser(user User) error {
	file, err := os.OpenFile(USERS_FILE_PATH, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	userString := fmt.Sprintf("- userID: %d\n  password: %s\n  usage: %d\n", user.UserID, user.Password, user.Usage)
	_, err = io.WriteString(file, userString)
	if err != nil {
		return err
	}
	return nil
}

func removeUserFromYamlFile(userID uint16) error {
	// Read the YAML file into a byte slice
	userBytes, err := ioutil.ReadFile(USERS_FILE_PATH)
	if err != nil {
		return err
	}

	// Parse the YAML data into a slice of User structs
	var users []User
	err = yaml.Unmarshal(userBytes, &users)
	if err != nil {
		return err
	}

	// Find the user with the given ID and remove it from the slice
	found := false
	for i, user := range users {
		if user.UserID == userID {
			users = append(users[:i], users[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		// TODO find out why this does not work
		return fmt.Errorf("user with ID %d not found", userID)
	}

	// Encode the updated user slice into YAML format
	updatedUserBytes, err := yaml.Marshal(&users)
	if err != nil {
		return err
	}

	// Write the updated user data to the file
	err = ioutil.WriteFile(USERS_FILE_PATH, updatedUserBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func findUserById(id uint16) (User, error) {
	// Read the YAML file into a byte slice
	userBytes, err := ioutil.ReadFile(USERS_FILE_PATH)
	if err != nil {
		return User{}, err
	}

	// Parse the YAML data into a slice of User structs
	var users []User
	err = yaml.Unmarshal(userBytes, &users)
	if err != nil {
		return User{}, err
	}
	// Find the user with the given ID and remove it from the slice
	found := false
	u := User{}
	for _, user := range users {
		if user.UserID == id {
			u = user
			found = true
			break
		}
	}
	if !found {
		// TODO find out why this does not work
		return User{}, fmt.Errorf("user with ID %d not found", id)
	}
	return u, nil
}
