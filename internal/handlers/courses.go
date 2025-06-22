package handlers

import (
	"fmt"
	"qb/pkg/database"
	"qb/pkg/models"
	"qb/pkg/utils"

	"github.com/gin-gonic/gin"
)

func FilterCourses(c *gin.Context) {
    dept, level, semester, handled := utils.GetCourseFilterParams(c)
	if handled {
		return
	}

    var courses []models.Course
    err := database.DB.
        Joins("JOIN department_courses ON courses.id = department_courses.course_id").
        Where("department_courses.department_id = ? AND courses.level_id = ? AND courses.semester = ?", 
              dept, level, semester).
        Find(&courses).Error

    if err != nil {
        utils.HandleDatabaseError(c, err)
        return
    }

    utils.SuccessResponse(c, courses)
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
		utils.HandleError(c, utils.NewValidationError("Level "+fmt.Sprintf("%d", level)+" does not exist"))
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
