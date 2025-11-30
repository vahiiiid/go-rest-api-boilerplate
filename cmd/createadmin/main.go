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

	"golang.org/x/term"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vahiiiid/go-rest-api-boilerplate/internal/config"
	"github.com/vahiiiid/go-rest-api-boilerplate/internal/user"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) > 255 {
		return fmt.Errorf("name is too long (max 255 characters)")
	}
	return nil
}

func checkPasswordsMatch(password, confirmPassword string) error {
	if password != confirmPassword {
		return fmt.Errorf("passwords do not match")
	}
	return nil
}

func promoteUserToAdmin(ctx context.Context, service user.Service, userID uint) error {
	existingUser, err := service.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if existingUser.IsAdmin() {
		fmt.Printf("User %s (%s) is already an admin\n", existingUser.Name, existingUser.Email)
		return nil
	}

	if err := service.PromoteToAdmin(ctx, userID); err != nil {
		return fmt.Errorf("failed to promote user: %w", err)
	}

	fmt.Printf("Successfully promoted %s (%s) to admin\n", existingUser.Name, existingUser.Email)
	return nil
}

func registerAndPromoteUser(ctx context.Context, service user.Service, email, password, name string) (*user.User, error) {
	registerReq := user.RegisterRequest{
		Email:    email,
		Password: password,
		Name:     name,
	}

	newUser, err := service.RegisterUser(ctx, registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if err := service.PromoteToAdmin(ctx, newUser.ID); err != nil {
		return nil, fmt.Errorf("failed to promote user to admin: %w", err)
	}

	return newUser, nil
}

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
	if err := promoteUserToAdmin(ctx, service, userID); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func createNewAdmin(ctx context.Context, service user.Service) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter admin email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	if err := validateEmail(email); err != nil {
		log.Fatalf("Invalid email: %v", err)
	}

	fmt.Print("Enter admin name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	if err := validateName(name); err != nil {
		log.Fatalf("Invalid name: %v", err)
	}

	fmt.Println("\nPassword requirements:")
	fmt.Println("  • Minimum 8 characters")
	fmt.Println("  • At least one uppercase letter (A-Z)")
	fmt.Println("  • At least one lowercase letter (a-z)")
	fmt.Println("  • At least one digit (0-9)")
	fmt.Println("  • At least one special character (!@#$%^&*()_+-=[]{}...)")
	fmt.Println()

	password := readPassword("Enter admin password: ")
	if err := validatePassword(password); err != nil {
		log.Fatalf("Invalid password: %v", err)
	}

	confirmPassword := readPassword("Confirm password: ")
	if err := checkPasswordsMatch(password, confirmPassword); err != nil {
		log.Fatalf("Password mismatch: %v", err)
	}

	newUser, err := registerAndPromoteUser(ctx, service, email, password, name)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("\nAdmin user created successfully:\n")
	fmt.Printf("ID: %d\n", newUser.ID)
	fmt.Printf("Email: %s\n", newUser.Email)
	fmt.Printf("Name: %s\n", newUser.Name)
	fmt.Printf("Roles: admin, user\n")
}

func readPassword(prompt string) string {
	fmt.Print(prompt)

	// Check if stdin is a terminal
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		// Fallback to regular input if not a terminal
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Failed to read password: %v", err)
		}
		return strings.TrimSpace(password)
	}

	// Use secure password reading if terminal is available
	bytePassword, err := term.ReadPassword(fd)
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
