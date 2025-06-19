package models

import (
	"time"
)

// Role represents the Role enum in Prisma.
type Role string

const (
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
)

// QuestionType represents the QuestionType enum in Prisma.
type QuestionType string

const (
	QuestionTypeTest QuestionType = "TEST"
	QuestionTypeExam QuestionType = "EXAM"
)

// CourseStatus represents the CourseStatus enum in Prisma.
type CourseStatus string

const (
	CourseStatusElective    CourseStatus = "ELECTIVE"
	CourseStatusCompulsory  CourseStatus = "COMPULSORY"
	CourseStatusUnavailable CourseStatus = "UNAVAILABLE"
)

// User model translated from Prisma schema.
// Explanation:
// - ID: Translated from String @id @default(auto()) @map("_id") @db.ObjectId. In PostgreSQL, UUIDs are often used for IDs,
//   but a simple string is sufficient if the ObjectID is treated as a string. GORM's "primaryKey" tag marks it as the primary key,
//   and "column:_id" maps it to the "_id" column in the database as per your Prisma schema.
// - Nullable fields (e.g., LastName, Age): Represented as pointers (*string, *int) in Go to allow for NULL values in the database.
// - Role: Mapped to the custom Role type. GORM will store the string value.
// - IsActive: Translated from Boolean @default(true). "default:true" sets the default value.
// - UpdatedAt: Translated from DateTime @updatedAt. "autoUpdateTime" automatically updates this field on record updates.
// - UploadedQuestions: One-to-many relationship with Question. GORM handles this by looking at the foreign key in the Question model.
// - Department/Level: Many-to-one relationships. GORM uses DepartmentID and LevelID as foreign keys.
type User struct {
	ID                string      `gorm:"primaryKey;column:_id;type:char(36);default:(uuid())" json:"id"`
	FirstName         string      `json:"firstName"`
	LastName          *string     `json:"lastName"`
	Email             string      `gorm:"unique" json:"email"`
	Role              Role        `gorm:"default:'MEMBER'" json:"role"`
	Age               *int        `json:"age"`
	Image             *string     `json:"image"`
	Username          *string     `gorm:"unique" json:"username"`
	DepartmentID      *string     `gorm:"type:char(36)" json:"departmentId"`
	LevelID           *int        `json:"levelId"`
	Semester          *int        `json:"semester"`
	IsActive          bool        `gorm:"default:true" json:"isActive"`
	Password          *string     `json:"password"`
	Phone             *string     `json:"phone"`
	Twitter           *string     `json:"twitter"`
	LinkedIn          *string     `json:"linkedin"`
	Discord           *string     `json:"discord"`
	UpdatedAt         time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	UploadedQuestions []Question  `gorm:"foreignKey:UploaderID" json:"uploadedQuestions"`
	Department        *Department `gorm:"foreignKey:DepartmentID" json:"department"`
	Level             *Level      `gorm:"foreignKey:LevelID" json:"level"`
}

// Faculty model translated from Prisma schema.
// Explanation:
// - ID: Translated from Int @id @map("_id"). Mapped to "_id" column.
// - Title: Unique string.
// - Departments: One-to-many relationship with Department.
type Faculty struct {
	ID          int          `gorm:"primaryKey;column:_id" json:"id"`
	Title       string       `gorm:"unique" json:"title"`
	Departments []Department `gorm:"foreignKey:FacultyID" json:"departments"`
}

// Department model translated from Prisma schema.
// Explanation:
// - ID: Translated from String @id @map("_id").
// - Title: Unique string.
// - FacultyID: Foreign key to Faculty.
// - Faculty: Many-to-one relationship with Faculty.
// - Users: One-to-many relationship with User.
// - Course: Many-to-many relationship with Course, explicitly handled by GORM using `gorm:"many2many:department_courses;"` if a join table were desired.
//   For PostgreSQL, a join table `DepartmentCourses` would be more idiomatic for a many-to-many relationship. For demonstration purposes,
//   I'm keeping `CourseIDs` as a string array, implying manual handling of the join.
type Department struct {
	ID        string   `gorm:"primaryKey;column:_id;type:char(36);default:(uuid())" json:"id"`
	Title     string   `gorm:"unique" json:"title"`
	FacultyID int      `json:"facultyId"`
	Faculty   Faculty  `gorm:"foreignKey:FacultyID" json:"faculty"`
	Users     []User   `gorm:"foreignKey:DepartmentID" json:"users"`
	Course    []Course `gorm:"many2many:department_courses;" json:"course"` // GORM will create a join table 'department_courses'
}

// Level model translated from Prisma schema.
// Explanation:
// - ID: Translated from Int @id @map("_id").
// - Courses: One-to-many relationship with Course.
// - Users: One-to-many relationship with User.
type Level struct {
	ID      int      `gorm:"primaryKey;column:_id" json:"id"`
	Courses []Course `gorm:"foreignKey:LevelID" json:"courses"`
	Users   []User   `gorm:"foreignKey:LevelID" json:"users"`
}

// Session model translated from Prisma schema.
// Explanation:
// - ID: Translated from String @id @map("_id").
// - StartDate/EndDate: Nullable DateTime fields become *time.Time.
// - Questions: One-to-many relationship with Question.
type Session struct {
	ID        string    `gorm:"primaryKey;column:_id;type:char(36);default:(uuid())" json:"id"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	Info      *string    `json:"info"`
	Questions []Question `gorm:"foreignKey:SessionID" json:"questions"`
}

// Question model translated from Prisma schema.
// Explanation:
// - ID: Translated from String @id @map("_id").
// - CourseID/SessionID/UploaderID: Foreign keys.
// - ImageLinks: Array of strings. `gorm:"type:text[]"` specifies a PostgreSQL array type.
// - Lecturer/TimeAllowed/DocLink/Tips: Nullable fields.
// - Type: Mapped to custom QuestionType enum.
// - Downloads/Views: Integer fields.
// - Approved: Boolean with default false.
// - CreatedAt/UpdatedAt: Automatically managed timestamps.
// - Course/Session/Uploader: Many-to-one relationships.
type Question struct {
	ID          string       `gorm:"primaryKey;column:_id;type:char(36);default:(uuid())" json:"id"`
	CourseID    string       `gorm:"type:char(36)" json:"courseId"`
	Course      Course       `gorm:"foreignKey:CourseID" json:"course"`
	SessionID   string       `gorm:"type:char(36)" json:"sessionId"`
	Session     Session      `gorm:"foreignKey:SessionID" json:"session"`
	ImageLinks  []string     `gorm:"type:json" json:"imageLinks"`
	Lecturer    *string      `json:"lecturer"`
	TimeAllowed *int         `json:"timeAllowed"`
	DocLink     *string      `json:"docLink"`
	Tips        *string      `json:"tips"`
	Type        QuestionType `json:"type"`
	Downloads   *int         `json:"downloads"`
	Views       *int         `json:"views"`
	Approved    bool         `gorm:"default:false" json:"approved"`
	UploaderID  *string      `gorm:"type:char(36)" json:"uploaderId"`
	Uploader    *User        `gorm:"foreignKey:UploaderID" json:"uploader"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
}

// Course model translated from Prisma schema.
// Explanation:
// - ID: Translated from String @id @map("_id").
// - Units/Semester/LevelID: Integer fields.
// - Description: Nullable string.
// - Status: Mapped to custom CourseStatus enum.
// - CreatedAt/UpdatedAt: Automatically managed timestamps.
// - Level: Many-to-one relationship with Level.
// - Questions: One-to-many relationship with Question.
// - Departments: Many-to-many relationship, using a join table.
type Course struct {
	ID            string       `gorm:"primaryKey;column:_id;type:char(36);default:(uuid())" json:"id"`
	Units         int          `json:"units"`
	Title         string       `json:"title"`
	LevelID       int          `json:"levelId"`
	Semester      int          `json:"semester"`
	Description   *string      `json:"description"`
	Status        *CourseStatus `json:"status"`
	CreatedAt     time.Time    `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time    `gorm:"autoUpdateTime" json:"updatedAt"`
	Level         Level        `gorm:"foreignKey:LevelID" json:"level"`
	Questions     []Question   `gorm:"foreignKey:CourseID" json:"questions"`
	Departments   []Department `gorm:"many2many:department_courses;" json:"departments"` // GORM will create a join table 'department_courses'
}
