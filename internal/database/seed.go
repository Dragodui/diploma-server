package database

import (
	"log"
	"time"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/pkg/security"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// SeedDatabase populates the database with test data for development purposes
// WARNING: This should NEVER be used in production environments
// Test users use weak passwords (password123) and should only be used for local development
func SeedDatabase(db *gorm.DB) error {
	// Check if data already exists
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	log.Println("Seeding database with test data...")

	// Create users
	passwordHash, _ := security.HashPassword("password123")

	admin := models.User{
		Email:         "admin@example.com",
		EmailVerified: true,
		Name:          "Admin User",
		PasswordHash:  passwordHash,
	}

	user1 := models.User{
		Email:         "john@example.com",
		EmailVerified: true,
		Name:          "John Doe",
		PasswordHash:  passwordHash,
	}

	user2 := models.User{
		Email:         "jane@example.com",
		EmailVerified: true,
		Name:          "Jane Smith",
		PasswordHash:  passwordHash,
	}

	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	if err := db.Create(&user1).Error; err != nil {
		return err
	}
	if err := db.Create(&user2).Error; err != nil {
		return err
	}

	// Create home
	home := models.Home{
		Name:       "Test Home",
		InviteCode: "TEST1234",
	}
	if err := db.Create(&home).Error; err != nil {
		return err
	}

	// Create memberships
	memberships := []models.HomeMembership{
		{HomeID: home.ID, UserID: admin.ID, Role: "admin"},
		{HomeID: home.ID, UserID: user1.ID, Role: "member"},
		{HomeID: home.ID, UserID: user2.ID, Role: "member"},
	}
	if err := db.Create(&memberships).Error; err != nil {
		return err
	}

	// Create rooms
	rooms := []models.Room{
		{HomeID: home.ID, Name: "Living Room"},
		{HomeID: home.ID, Name: "Kitchen"},
		{HomeID: home.ID, Name: "Bedroom"},
		{HomeID: home.ID, Name: "Bathroom"},
	}
	if err := db.Create(&rooms).Error; err != nil {
		return err
	}

	// Create tasks
	dueDate := time.Now().Add(24 * time.Hour)
	dueDate2 := time.Now().Add(48 * time.Hour)
	dueDate3 := time.Now().Add(72 * time.Hour)

	tasks := []models.Task{
		{HomeID: home.ID, RoomID: &rooms[0].ID, Name: "Clean living room", Description: "Vacuum and dust the living room", ScheduleType: "once", DueDate: &dueDate},
		{HomeID: home.ID, RoomID: &rooms[1].ID, Name: "Wash dishes", Description: "Clean all dishes in the sink", ScheduleType: "daily", DueDate: &dueDate},
		{HomeID: home.ID, RoomID: &rooms[2].ID, Name: "Change bed sheets", Description: "Replace bed sheets with fresh ones", ScheduleType: "weekly", DueDate: &dueDate2},
		{HomeID: home.ID, RoomID: &rooms[3].ID, Name: "Clean bathroom", Description: "Scrub toilet, sink and shower", ScheduleType: "weekly", DueDate: &dueDate3},
		{HomeID: home.ID, Name: "Take out trash", Description: "Empty all trash bins", ScheduleType: "daily", DueDate: &dueDate},
	}
	if err := db.Create(&tasks).Error; err != nil {
		return err
	}

	// Create task assignments
	assignments := []models.TaskAssignment{
		{TaskID: tasks[0].ID, UserID: user1.ID, Status: "assigned"},
		{TaskID: tasks[1].ID, UserID: user2.ID, Status: "assigned"},
		{TaskID: tasks[2].ID, UserID: admin.ID, Status: "completed"},
		{TaskID: tasks[3].ID, UserID: user1.ID, Status: "assigned"},
		{TaskID: tasks[4].ID, UserID: user2.ID, Status: "assigned"},
	}
	if err := db.Create(&assignments).Error; err != nil {
		return err
	}

	// Create polls
	endsAt := time.Now().Add(7 * 24 * time.Hour)
	polls := []models.Poll{
		{
			HomeID:      home.ID,
			Question:    "What color should we paint the kitchen?",
			Type:        "public",
			Status:      "open",
			AllowRevote: true,
			EndsAt:      &endsAt,
		},
		{
			HomeID:      home.ID,
			Question:    "Should we get a pet?",
			Type:        "public",
			Status:      "open",
			AllowRevote: false,
			EndsAt:      &endsAt,
		},
	}
	if err := db.Create(&polls).Error; err != nil {
		return err
	}

	// Create poll options
	options := []models.Option{
		{PollID: polls[0].ID, Title: "Blue"},
		{PollID: polls[0].ID, Title: "Green"},
		{PollID: polls[0].ID, Title: "White"},
		{PollID: polls[1].ID, Title: "Yes, a cat"},
		{PollID: polls[1].ID, Title: "Yes, a dog"},
		{PollID: polls[1].ID, Title: "No pets"},
	}
	if err := db.Create(&options).Error; err != nil {
		return err
	}

	// Create some votes
	votes := []models.Vote{
		{UserID: admin.ID, OptionID: options[0].ID},
		{UserID: user1.ID, OptionID: options[1].ID},
		{UserID: user2.ID, OptionID: options[0].ID},
		{UserID: admin.ID, OptionID: options[3].ID},
		{UserID: user1.ID, OptionID: options[4].ID},
	}
	if err := db.Create(&votes).Error; err != nil {
		return err
	}

	// Create bill categories
	billCategories := []models.BillCategory{
		{HomeID: home.ID, Name: "Electricity", Color: "#FFD700"},
		{HomeID: home.ID, Name: "Water", Color: "#4169E1"},
		{HomeID: home.ID, Name: "Internet", Color: "#32CD32"},
		{HomeID: home.ID, Name: "Gas", Color: "#FF6347"},
	}
	if err := db.Create(&billCategories).Error; err != nil {
		return err
	}

	// Create bills
	startDate := time.Now().AddDate(0, -1, 0)
	endDate := time.Now()
	bills := []models.Bill{
		{HomeID: home.ID, BillCategoryID: &billCategories[0].ID, Type: "electricity", Payed: true, TotalAmount: 85.50, Start: startDate, End: endDate, UploadedBy: admin.ID, OCRData: datatypes.JSON([]byte(`{}`))},
		{HomeID: home.ID, BillCategoryID: &billCategories[1].ID, Type: "water", Payed: false, TotalAmount: 45.00, Start: startDate, End: endDate, UploadedBy: admin.ID, OCRData: datatypes.JSON([]byte(`{}`))},
		{HomeID: home.ID, BillCategoryID: &billCategories[2].ID, Type: "internet", Payed: true, TotalAmount: 59.99, Start: startDate, End: endDate, UploadedBy: user1.ID, OCRData: datatypes.JSON([]byte(`{}`))},
	}
	if err := db.Create(&bills).Error; err != nil {
		return err
	}

	// Create shopping categories
	groceryIcon := "cart"
	cleaningIcon := "spray"
	shoppingCategories := []models.ShoppingCategory{
		{HomeID: home.ID, Name: "Groceries", Icon: &groceryIcon, Color: "#90EE90"},
		{HomeID: home.ID, Name: "Cleaning Supplies", Icon: &cleaningIcon, Color: "#87CEEB"},
	}
	if err := db.Create(&shoppingCategories).Error; err != nil {
		return err
	}

	// Create shopping items
	shoppingItems := []models.ShoppingItem{
		{CategoryID: shoppingCategories[0].ID, Name: "Milk", UploadedBy: admin.ID, IsBought: false},
		{CategoryID: shoppingCategories[0].ID, Name: "Bread", UploadedBy: user1.ID, IsBought: false},
		{CategoryID: shoppingCategories[0].ID, Name: "Eggs", UploadedBy: user2.ID, IsBought: true},
		{CategoryID: shoppingCategories[1].ID, Name: "Dish soap", UploadedBy: admin.ID, IsBought: false},
		{CategoryID: shoppingCategories[1].ID, Name: "Paper towels", UploadedBy: user1.ID, IsBought: false},
	}
	if err := db.Create(&shoppingItems).Error; err != nil {
		return err
	}

	log.Println("Database seeded successfully!")
	log.Println("Test users created:")
	log.Println("  - admin@example.com (Admin)")
	log.Println("  - john@example.com")
	log.Println("  - jane@example.com")
	log.Println("Default password for test users is set in seed.go (DO NOT use in production)")

	return nil
}
