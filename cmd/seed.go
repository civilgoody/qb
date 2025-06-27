package main

import (
	"fmt"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"
)

func main() {
	// Load environment variables
	utils.LoadDotEnv()

	// Initialize database connection
	database.ConnectDB()

	fmt.Println("🌱 Starting database seeding...")

	// Seed data
	seedFaculties()
	seedDepartments()
	seedLevels()
	seedSessions()
	// seedUsers()
	// seedCourses()
	// seedQuestions()

	fmt.Println("✅ Database seeding completed!")
}

func seedFaculties() {
	faculties := []models.Faculty{
		{ID: 1, Title: "Faculty of Engineering"},
		{ID: 2, Title: "Faculty of Science"},
		{ID: 3, Title: "Faculty of Arts"},
		{ID: 4, Title: "Faculty of Medicine"},
	}

	for _, faculty := range faculties {
		database.DB.FirstOrCreate(&faculty, models.Faculty{ID: faculty.ID})
	}
	fmt.Println("📚 Seeded faculties")
}

func seedDepartments() {
	departments := []models.Department{
		{ID: "CEG", Title: "Civil Engineering", FacultyID: 1},
		{ID: "EEG", Title: "Electrical & Electronics Engineering", FacultyID: 1},
		{ID: "MEG", Title: "Mechanical Engineering", FacultyID: 1},
		{ID: "CSC", Title: "Computer Science", FacultyID: 2},
		{ID: "MTH", Title: "Mathematics", FacultyID: 2},
		{ID: "PHY", Title: "Physics", FacultyID: 2},
		{ID: "ENG", Title: "English", FacultyID: 3},
		{ID: "HIS", Title: "History", FacultyID: 3},
	}

	for _, dept := range departments {
		database.DB.FirstOrCreate(&dept, models.Department{ID: dept.ID})
	}
	fmt.Println("🏢 Seeded departments")
}

func seedLevels() {
	levels := []models.Level{
		{ID: 100},
		{ID: 200},
		{ID: 300},
		{ID: 400},
		{ID: 500},
	}

	for _, level := range levels {
		database.DB.FirstOrCreate(&level, models.Level{ID: level.ID})
	}
	fmt.Println("📊 Seeded levels")
}

func seedUsers() {
	users := []models.User{
		{
			ID:           "user1",
			FirstName:    "John",
			LastName:     stringPtr("Doe"),
			Email:        "john.doe@example.com",
			Role:         models.RoleAdmin,
			Age:          intPtr(25),
			Username:     stringPtr("johndoe"),
			DepartmentID: stringPtr("CEG"),
			LevelID:      intPtr(400),
			Semester:     intPtr(1),
			Phone:        stringPtr("+1234567890"),
		},
		{
			ID:           "user2",
			FirstName:    "Jane",
			LastName:     stringPtr("Smith"),
			Email:        "jane.smith@example.com",
			Role:         models.RoleMember,
			Age:          intPtr(23),
			Username:     stringPtr("janesmith"),
			DepartmentID: stringPtr("EEE"),
			LevelID:      intPtr(300),
			Semester:     intPtr(2),
		},
		{
			ID:        "user3",
			FirstName: "Alice",
			Email:     "alice@example.com",
			Role:      models.RoleMember,
			LevelID:   intPtr(200),
			Semester:  intPtr(1),
		},
	}

	for _, user := range users {
		database.DB.FirstOrCreate(&user, models.User{ID: user.ID})
	}
	fmt.Println("👥 Seeded users")
}

func seedSessions() {
	sessions := []models.Session{
		{
			ID:        "23-24",
			StartDate: 2023,
			EndDate:   2024,
			Info:      stringPtr("Academic Session 2023/2024"),
		},
		{
			ID:        "24-25",
			StartDate: 2024,
			EndDate:   2025,
			Info:      stringPtr("Academic Session 2024/2025"),
		},
		{
			ID:        "22-23",
			StartDate: 2022,
			EndDate:   2023,
			Info:      stringPtr("Academic Session 2022/2023"),
		},
	}

	for _, session := range sessions {
		database.DB.FirstOrCreate(&session, models.Session{ID: session.ID})
	}
	fmt.Println("📅 Seeded sessions")
}

func seedCourses() {
	compulsory := models.CourseStatusCompulsory
	elective := models.CourseStatusElective

	courses := []models.Course{
		{
			ID:          "CEG543",
			Units:       3,
			Title:       "Structural Analysis",
			LevelID:     500,
			Semester:    2,
			Description: stringPtr("Advanced structural analysis techniques"),
			Status:      &compulsory,
		},
		{
			ID:          "EEE321",
			Units:       4,
			Title:       "Digital Electronics",
			LevelID:     300,
			Semester:    1,
			Description: stringPtr("Introduction to digital circuits and systems"),
			Status:      &elective,
		},
		{
			ID:          "CSC412",
			Units:       3,
			Title:       "Database Systems",
			LevelID:     400,
			Semester:    1,
			Description: stringPtr("Database design and management"),
			Status:      &compulsory,
		},
		{
			ID:          "MTH204",
			Units:       2,
			Title:       "Linear Algebra",
			LevelID:     200,
			Semester:    2,
			Description: stringPtr("Vector spaces and linear transformations"),
			Status:      &compulsory,
		},
	}

	for _, course := range courses {
		database.DB.FirstOrCreate(&course, models.Course{ID: course.ID})
	}
	fmt.Println("📖 Seeded courses")
}

func seedQuestions() {
	questions := []models.Question{
		{
			ID:          "q1",
			CourseID:    "CEG543",
			SessionID:   "23-24",
			ImageLinks:  []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
			Lecturer:    stringPtr("Prof. Johnson"),
			TimeAllowed: intPtr(180),
			Tips:        stringPtr("Focus on beam analysis"),
			Type:        models.QuestionTypeExam,
			Downloads:   intPtr(45),
			Views:       intPtr(123),
			Approved:    true,
			UploaderID:  stringPtr("user1"),
		},
		{
			ID:          "q2",
			CourseID:    "EEE321",
			SessionID:   "23-24",
			ImageLinks:  []string{"https://example.com/image3.jpg"},
			Lecturer:    stringPtr("Dr. Williams"),
			TimeAllowed: intPtr(120),
			Type:        models.QuestionTypeTest,
			Downloads:   intPtr(23),
			Views:       intPtr(67),
			Approved:    false,
			UploaderID:  stringPtr("user2"),
		},
		{
			ID:        "q3",
			CourseID:  "CSC412",
			SessionID: "24-25",
			Type:      models.QuestionTypeExam,
			Downloads: intPtr(12),
			Views:     intPtr(34),
			Approved:  true,
		},
	}

	for _, question := range questions {
		database.DB.FirstOrCreate(&question, models.Question{ID: question.ID})
	}
	fmt.Println("❓ Seeded questions")
}

// Helper functions for pointer creation
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
} 
