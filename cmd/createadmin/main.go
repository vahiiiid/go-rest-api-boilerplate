package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func main() {
	promoteID := flag.Int("promote", 0, "Promote existing user ID to admin")
	flag.Parse()

	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := user.NewRepository(db)
	service := user.NewService(repo)

	ctx := context.Background()

	if *promoteID > 0 {
		promoteExistingUser(ctx, service, uint(*promoteID))
	} else {
		createNewAdmin(ctx, service)
	}
}

func promoteExistingUser(ctx context.Context, service user.Service, userID uint) {
	existingUser, err := service.GetUserByID(ctx, userID)
	if err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}

	if existingUser.IsAdmin() {
		fmt.Printf("User %s (%s) is already an admin\n", existingUser.Name, existingUser.Email)
		return
	}

	if err := service.PromoteToAdmin(ctx, userID); err != nil {
		log.Fatalf("Failed to promote user: %v", err)
	}

	fmt.Printf("Successfully promoted %s (%s) to admin\n", existingUser.Name, existingUser.Email)
}

func createNewAdmin(ctx context.Context, service user.Service) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter admin email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if !emailRegex.MatchString(email) {
		log.Fatal("Invalid email format")
	}

	fmt.Print("Enter admin name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if name == "" {
		log.Fatal("Name cannot be empty")
	}

	password := readPassword("Enter admin password: ")
	if err := validatePassword(password); err != nil {
		log.Fatalf("Invalid password: %v", err)
	}

	confirmPassword := readPassword("Confirm password: ")
	if password != confirmPassword {
		log.Fatal("Passwords do not match")
	}

	registerReq := user.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}

	newUser, err := service.RegisterUser(ctx, registerReq)
	if err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}

	if err := service.PromoteToAdmin(ctx, newUser.ID); err != nil {
		log.Fatalf("Failed to promote user to admin: %v", err)
	}

	fmt.Printf("\nAdmin user created successfully:\n")
	fmt.Printf("ID: %d\n", newUser.ID)
	fmt.Printf("Email: %s\n", newUser.Email)
	fmt.Printf("Name: %s\n", newUser.Name)
	fmt.Printf("Roles: admin, user\n")
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(syscall.Stdin)
	fmt.Println() // Print newline after password input
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}
	return strings.TrimSpace(string(bytePassword))
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Enforce strong password policy for admin accounts
	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(password)
		hasDigit   = regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	)

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
