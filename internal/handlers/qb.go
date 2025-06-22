package handlers

import (
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	utils.SuccessResponse(c, gin.H{"message": "Welcome to the Qb API"})
}

func CreateFaculty(c *gin.Context) {
	var faculty models.Faculty
	
	if utils.HandleCreateResource(c, database.DB, utils.Validator, &faculty) {
		return
	}
}

func GetFaculties(c *gin.Context) {
	var faculties []models.Faculty
	
	if utils.HandleGetResources(c, database.DB, &faculties) {
		return
	}
}

func DeleteFaculty(c *gin.Context) {
	if utils.DeleteResourceByID(c, database.DB, "id", "Faculty", &models.Faculty{}) {
		return
	}
}

func CreateLevel(c *gin.Context) {
	var level models.Level

	if utils.HandleCreateResource(c, database.DB, utils.Validator, &level) {
		return
	}
}

func GetLevels(c *gin.Context) {
	var levels []models.Level
	
	if utils.HandleGetResources(c, database.DB, &levels) {
		return
	}
}

func DeleteLevel(c *gin.Context) {
	if utils.DeleteResourceByID(c, database.DB, "id", "Level", &models.Level{}) {
		return
	}
}

func CreateDepartment(c *gin.Context) {
	var department models.Department
	
	if utils.HandleCreateResource(c, database.DB, utils.Validator, &department) {
		return
	}
}

func GetDepartments(c *gin.Context) {
	var departments []models.Department
	
	if err := database.DB.Preload("Faculty").Find(&departments).Error; err != nil {
		utils.HandleError(c, utils.ErrDatabase)
		return
	}
	utils.SuccessResponse(c, departments)
}

func DeleteDepartment(c *gin.Context) {
	if utils.DeleteResourceByStringID(c, database.DB, "id", "Department", &models.Department{}) {
		return
	}
}

func CreateCourse(c *gin.Context) {
	var input models.CreateCourseDTO

	// Step 1: Bind and validate the DTO
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	if err := utils.Validator.Struct(&input); err != nil {
		utils.HandleValidationError(c, err)
		return
	}

	// Step 2: Parse the course code
	departmentCode, level, semester, _, err := input.ParseCourseCode()
	if err != nil {
		utils.HandleError(c, utils.NewValidationError(err.Error()))
		return
	}

	// Step 3: Validate that the department exists
	var department models.Department
	if err := database.DB.Where("id = ?", departmentCode).First(&department).Error; err != nil {
		utils.HandleError(c, utils.NewValidationError("Department with code '"+departmentCode+"' does not exist"))
		return
	}

	// Step 4: Validate that the level exists
	var levelModel models.Level
	if err := database.DB.Where("id = ?", level).First(&levelModel).Error; err != nil {
		utils.HandleError(c, utils.NewValidationError("Level "+strconv.Itoa(level)+" does not exist"))
		return
	}

	// Step 5: Create the course with parsed values (using course code as ID)
	course := models.Course{
		ID:          input.Code,
		Units:       input.Units,
		Title:       input.Title,
		LevelID:     level,
		Semester:    semester,
		Description: input.Description,
		Status:      input.Status,
		Departments: []models.Department{department}, // Associate with the parsed department
	}

	// Step 6: Create the course in the database
	if err := database.DB.Create(&course).Error; err != nil {
		utils.HandleDatabaseError(c, err)
		return
	}

	// Step 7: Return success response
	utils.SuccessResponse(c, course)
}

func GetCourses(c *gin.Context) {
	var courses []models.Course
	
	if utils.HandleGetResources(c, database.DB, &courses) {
		return
	}
}

func DeleteCourse(c *gin.Context) {
	if utils.DeleteResourceByStringID(c, database.DB, "id", "Course", &models.Course{}) {
		return
	}
}

func CreateSession(c *gin.Context) {
	var session models.Session

	if utils.HandleCreateResource(c, database.DB, utils.Validator, &session) {
		return
	}
}

func GetSessions(c *gin.Context) {
	var sessions []models.Session

	if utils.HandleGetResources(c, database.DB, &sessions) {
		return
	}
}

func DeleteSession(c *gin.Context) {
	if utils.DeleteResourceByStringID(c, database.DB, "id", "Session", &models.Session{}) {
		return
	}
}
