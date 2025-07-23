package repository

import (
	"database/sql"
	"encoding/json"
	"log"
	"server/internal/models"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type projectPSQL struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) ProjectRepo {
	return &projectPSQL{db: db}
}

// DeleteProject deletes a project and its corresponding table from the database by the provided project name.
func (l *projectPSQL) DeleteProject(name string) error {
	err := l.db.Delete(&models.Project{}, "name = ?", name).Error
	if err != nil {
		return err
	}
	return nil
}

// GetAllProjects retrieves a list of projects from the database with pagination options based on page and limit values.
func (l *projectPSQL) GetAllProjects(page int, limit int) ([]*models.Project, error) {
	var projects []*models.Project
	err := l.db.Offset((page-1)*limit).Limit(limit).Select("name", "description", "id", "created_at").Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProjectByID retrieves a project by its ID from the database and returns the project object or an error if not found.
func (l *projectPSQL) GetProjectByID(id string) (*models.Project, error) {
	var project models.Project
	err := l.db.First(&project, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// GetProjectByName retrieves a project from the database based on the provided project name.
// Returns the project and any error encountered during the operation.
func (l *projectPSQL) GetProjectByName(name string) (*models.Project, error) {
	var project models.Project
	err := l.db.First(&project, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

// CreateProject adds a new project to the database, initializes a log table, and retrieves the complete project record.
func (l *projectPSQL) CreateProject(project *models.Project) (*models.Project, error) {
	//project.ID = uuid.NewString() // generate project id
	//project.Active = true
	res := l.db.Create(&project)
	if res.Error != nil {
		return nil, res.Error
	}

	var fullRecord models.Project
	if err := l.db.First(&fullRecord, "id = ?", project.ID).Error; err != nil {
		return nil, err
	}
	return &fullRecord, nil
}

// GetProjectsCount retrieves the total number of projects from the database and returns the count along with any error encountered.
func (l *projectPSQL) GetProjectsCount() (int64, error) {
	var count int64
	err := l.db.Model(&models.Project{}).Count(&count).Error
	return count, err
}

// UpdateProject updates an existing project in the database based on its name and returns the updated project details.
func (l *projectPSQL) UpdateProject(project *models.Project) (*models.Project, error) {

	existingProject, err := l.GetProjectByName(project.Name) // check if a project exists
	if err != nil {
		return nil, err
	}

	if existingProject.Active != project.Active { // if active is changed
		err := l.db.Model(&project).Where("name = ?", project.Name).
			Select("active").Updates(project).Error
		if err != nil {
			return nil, err
		}
	}

	err = l.db.Where("name = ?", project.Name).Updates(project).Error // update project
	if err != nil {
		return nil, err
	}
	var updatedProject models.Project
	err = l.db.First(&updatedProject, "name = ?", project.Name).Error // get an updated project

	if err != nil {
		return nil, err
	}
	return &updatedProject, nil
}

// GetLogs retrieves logs from the database for the specified project name and returns them as a slice of Log models.
func (l *projectPSQL) GetLogs(projectName string) ([]*models.Log, error) {
	var logs []*models.Log
	var tempBuildDetails string // store build details as a string
	rows, err := l.db.Table(projectName).Select("service_name, level, message, stack, request_url, request_method, request_id, ip_address, user_agent, timestamp, build_details").Rows()
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) { // close rows
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)
	// iterate over rows
	for rows.Next() {
		lg := &models.Log{}
		err := rows.Scan(
			&lg.ServiceName,
			&lg.Level,
			&lg.Message,
			&lg.Stack,
			&lg.RequestUrl,
			&lg.RequestMethod,
			&lg.RequestId,
			&lg.IpAddress,
			&lg.UserAgent,
			&lg.Timestamp,
			&tempBuildDetails,
		)
		if err != nil {
			return nil, err
		}

		if tempBuildDetails != "" {
			buildDetails := &models.BuildDetails{}
			if err := json.Unmarshal([]byte(tempBuildDetails), buildDetails); err != nil { // unmarshal build details
				return nil, err
			}
			lg.BuildDetails = *buildDetails
		}

		logs = append(logs, lg)

	}
	return logs, nil
}

// GetRecentProjects retrieves a list of recent projects from the database based on the provided comma-separated project names.
func (l *projectPSQL) GetRecentProjects(projectNames string) ([]*models.Project, error) {
	projectsArr := strings.Split(projectNames, ",")
	var projects []*models.Project
	// find projects that are present in projectsArr
	err := l.db.Select("name", "description", "created_at", "id", "active").Where("name IN ?", projectsArr).Find(&projects).Error
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// UpsertKeyStore saves a new KeyStore record in the database.
func (l *projectPSQL) UpsertKeyStore(keyStore *models.KeyStore) error {
	return l.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "timestamp"}),
	}).Create(keyStore).Error
}
