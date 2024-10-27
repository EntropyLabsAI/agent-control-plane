package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	sentinel "github.com/entropylabsai/sentinel/server"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresqlStore struct {
	db *sql.DB
}

// Check if PostgresqlStore implements sentinel.Store
var _ sentinel.Store = &PostgresqlStore{}

// NewPostgresqlStore creates a new PostgreSQL store
func NewPostgresqlStore(connStr string) (*PostgresqlStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to the database: %w", err)
	}

	return &PostgresqlStore{db: db}, nil
}

// Close closes the database connection
func (s *PostgresqlStore) Close() error {
	return s.db.Close()
}

// ProjectStore implementation
func (s *PostgresqlStore) CreateProject(ctx context.Context, project sentinel.Project) error {
	query := `
		INSERT INTO project (id, name, created_at)
		VALUES ($1, $2, $3)`

	fmt.Printf("Creating project: %+v\n", project)

	_, err := s.db.ExecContext(ctx, query, project.Id, project.Name, project.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetProject(ctx context.Context, id uuid.UUID) (*sentinel.Project, error) {
	query := `
		SELECT id, name, created_at
		FROM project
		WHERE id = $1`

	var project sentinel.Project
	err := s.db.QueryRowContext(ctx, query, id).Scan(&project.Id, &project.Name, &project.CreatedAt)
	if err == sql.ErrNoRows {
		log.Printf("no rows found for project ID: %s\n", id)
		return nil, fmt.Errorf("no rows found for project ID: %s", id)
	}
	if err != nil {
		log.Printf("error getting project: %v\n", err)
		return nil, fmt.Errorf("error getting project: %w", err)
	}

	log.Printf("found project: %s\n", project.Name) // Add this line to see what we got
	return &project, nil
}

func (s *PostgresqlStore) GetProjects(ctx context.Context) ([]sentinel.Project, error) {
	query := `
		SELECT id, name, created_at
		FROM project
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing projects: %w", err)
	}
	defer rows.Close()

	var projects []sentinel.Project
	for rows.Next() {
		var project sentinel.Project
		if err := rows.Scan(&project.Id, &project.Name, &project.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning project: %w", err)
		}
		projects = append(projects, project)
	}

	// If no rows were found, projects will be empty slice
	return projects, nil
}

func (s *PostgresqlStore) GetRuns(ctx context.Context, projectId uuid.UUID) ([]sentinel.Run, error) {
	query := `
		SELECT id, project_id, created_at
		FROM run
		WHERE project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, projectId)
	if err != nil {
		return nil, fmt.Errorf("error getting runs: %w", err)
	}
	defer rows.Close()

	var runs []sentinel.Run
	for rows.Next() {
		var run sentinel.Run
		if err := rows.Scan(&run.Id, &run.ProjectId, &run.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning run: %w", err)
		}
		runs = append(runs, run)
	}

	// If no rows were found, runs will be empty slice
	return runs, nil
}

func (s *PostgresqlStore) GetProjectRuns(ctx context.Context, id uuid.UUID) ([]sentinel.Run, error) {
	query := `
		SELECT id, project_id, created_at
		FROM run
		WHERE project_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting project runs: %w", err)
	}
	defer rows.Close()

	var runs []sentinel.Run
	for rows.Next() {
		var run sentinel.Run
		if err := rows.Scan(&run.Id, &run.ProjectId, &run.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning run: %w", err)
		}
		runs = append(runs, run)
	}

	// If no rows were found, runs will be empty slice
	return runs, nil
}

// ReviewStore implementation
// Todo:
// store messages
// store tool requests
// store review status
// check result of review is stored somewhere
// ensure that the review status is updated 3 times (timeout, pending, completed)

// for _, toolRequest := range request.ToolRequests {
// err := store.CreateToolRequest(ctx, reviewID, toolRequest)
// if err != nil {
// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 	return
// }
// }
// for _, message := range request.Messages {
// 	err := store.CreateMessage(ctx, reviewID, message)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// }

func (s *PostgresqlStore) CreateReviewRequest(ctx context.Context, request sentinel.ReviewRequest) (uuid.UUID, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Store the llm_messages first and keep track of the IDs
	var messageIDs []uuid.UUID
	for _, message := range request.Messages {
		messageID := uuid.New()
		query := `
			INSERT INTO llm_message (id, role, content)
			VALUES ($1, $2, $3)`

		_, err = tx.ExecContext(ctx, query, messageID, message.Role, message.Content)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("error creating llm message: %w", err)
		}
		messageIDs = append(messageIDs, messageID)
	}

	// Store the review request
	// Marshal the TaskState map to JSON if it's a map
	taskStateJSON, err := json.Marshal(request.TaskState)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error marshalling task state: %w", err)
	}

	query := `
		INSERT INTO reviewrequest (id, run_id, task_state)
		VALUES ($1, $2, $3)`

	requestID := uuid.New()
	_, err = tx.ExecContext(ctx, query, requestID, request.RunId, taskStateJSON)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating review: %w", err)
	}

	// Store the tool requests
	for i, toolRequest := range request.ToolRequests {
		toolRequestID := uuid.New()

		query := `
			INSERT INTO toolrequest (id, reviewrequest_id, tool_id, message_id, arguments)
			VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.ExecContext(
			ctx, query, toolRequestID, requestID, toolRequest.ToolId, messageIDs[i], toolRequest.Arguments,
		)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("error creating tool request: %w", err)
		}
	}

	status := sentinel.ReviewStatus{Status: sentinel.Pending, CreatedAt: time.Now()}

	// Store a review status pending
	err = s.createReviewStatus(ctx, requestID, status, tx)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating review status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return requestID, nil
}

func (s *PostgresqlStore) createReviewStatus(ctx context.Context, requestID uuid.UUID, status sentinel.ReviewStatus, tx *sql.Tx) error {
	query := `
		INSERT INTO reviewrequest_status (id, reviewrequest_id, status, created_at)
		VALUES ($1, $2, $3, $4)`

	id := uuid.New()
	_, err := tx.ExecContext(ctx, query, id, requestID, status.Status, status.CreatedAt)
	if err != nil {
		return fmt.Errorf("error creating review status: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetReview(ctx context.Context, id uuid.UUID) (*sentinel.Review, error) {
	query := `
		SELECT r.id, r.run_id, r.task_state, rs.status
		FROM reviewrequest r
		LEFT JOIN reviewrequest_status rs ON r.id = rs.reviewrequest_id
		WHERE r.id = $1
		ORDER BY rs.created_at DESC
		LIMIT 1`

	var review sentinel.Review
	var status sentinel.ReviewStatus
	var taskStateJSON []byte // Add temporary variable for JSON data
	err := s.db.QueryRowContext(ctx, query, id).Scan(&review.Id, &review.RunId, &taskStateJSON, &status.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting review: %w", err)
	}

	// Parse the JSON task state
	if err := json.Unmarshal(taskStateJSON, &review.TaskState); err != nil {
		return nil, fmt.Errorf("error parsing task state: %w", err)
	}

	if status.Status != "" {
		review.Status = &status
	}
	return &review, nil
}

func (s *PostgresqlStore) GetReviewResults(ctx context.Context, id uuid.UUID) ([]*sentinel.ReviewResult, error) {
	query := `
		SELECT rr.id, rr.reviewrequest_id, rr.created_at, rr.decision, rr.reasoning, 
		rr.toolrequest_id, tr.tool_id, tr.message_id, tr.arguments
		FROM reviewresult rr
		LEFT JOIN toolrequest tr ON rr.toolrequest_id = tr.id
		WHERE rr.reviewrequest_id = $1`

	var tr sentinel.ToolRequest

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting review results: %w", err)
	}
	defer rows.Close()

	var results []*sentinel.ReviewResult
	for rows.Next() {
		var result sentinel.ReviewResult
		if err := rows.Scan(
			&result.Id, &result.ReviewRequestId, &result.CreatedAt, &result.Decision, &result.Reasoning,
			&tr.Id, &tr.ToolId, &tr.MessageId, &tr.Arguments,
		); err != nil {
			return nil, fmt.Errorf("error scanning review result: %w", err)
		}
		result.Toolrequest = &tr
		results = append(results, &result)
	}

	return results, nil

}

func (s *PostgresqlStore) UpdateReview(ctx context.Context, review sentinel.Review) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	if err = tx.Rollback(); err != nil {
		return fmt.Errorf("error rolling back transaction: %w", err)
	}

	// Update review request
	query1 := `
		UPDATE reviewrequest 
		SET task_state = $1
		WHERE id = $2`

	_, err = tx.ExecContext(ctx, query1, review.TaskState, review.Id)
	if err != nil {
		return fmt.Errorf("error updating review: %w", err)
	}

	// Insert new status
	query2 := `
		INSERT INTO reviewrequest_status (id, reviewrequest_id, created_at, status)
		VALUES ($1, $2, CURRENT_TIMESTAMP, $3)`

	_, err = tx.ExecContext(ctx, query2, review.Id, review.Id, review.Status.Status)
	if err != nil {
		return fmt.Errorf("error updating review status: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresqlStore) GetReviews(ctx context.Context) ([]sentinel.Review, error) {
	query := `
		SELECT rr.id, rr.run_id, rr.task_state, rs.id, rs.status, rs.created_at
		FROM reviewrequest rr
		LEFT JOIN reviewrequest_status rs ON rr.id = rs.reviewrequest_id
		ORDER BY rs.created_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting reviews: %w", err)
	}
	defer rows.Close()

	var reviews []sentinel.Review
	for rows.Next() {
		var review sentinel.Review
		var status sentinel.ReviewStatus
		var taskStateJSON []byte
		if err := rows.Scan(&review.Id, &review.RunId, &taskStateJSON, &status.Id, &status.Status, &status.CreatedAt); err != nil {
			return nil, fmt.Errorf("error scanning review: %w", err)
		}

		// Parse the JSON task state
		if err := json.Unmarshal(taskStateJSON, &review.TaskState); err != nil {
			return nil, fmt.Errorf("error parsing task state: %w", err)
		}

		review.Status = &status
		reviews = append(reviews, review)
	}

	return reviews, nil
}

func (s *PostgresqlStore) DeleteReview(ctx context.Context, id uuid.UUID) error {
	fmt.Printf("Stub: DeleteReview called with ID: %s\n", id)
	return nil
}

func (s *PostgresqlStore) CountReviews(ctx context.Context) (int, error) {
	fmt.Println("Stub: CountReviews called")
	return 0, nil
}

// ProjectToolStore implementation
func (s *PostgresqlStore) GetProjectTools(ctx context.Context, id uuid.UUID) ([]sentinel.Tool, error) {
	fmt.Printf("Stub: GetProjectTools called with project ID: %s\n", id)
	return []sentinel.Tool{}, nil
}

func (s *PostgresqlStore) CreateProjectTool(ctx context.Context, id uuid.UUID, tool sentinel.Tool) error {
	fmt.Printf("Stub: CreateProjectTool called with project ID: %s and tool ID: %s\n", id, tool.Id)
	return nil
}

func (s *PostgresqlStore) GetTool(ctx context.Context, id uuid.UUID) (*sentinel.Tool, error) {
	query := `
		SELECT id, name, attributes, description
		FROM tool
		WHERE id = $1`

	var tool sentinel.Tool
	var attributesJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(&tool.Id, &tool.Name, &attributesJSON, &tool.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting tool: %w", err)
	}

	// Parse the JSON attributes
	if err := json.Unmarshal(attributesJSON, &tool.Attributes); err != nil {
		return nil, fmt.Errorf("error parsing tool attributes: %w", err)
	}

	return &tool, nil
}

func (s *PostgresqlStore) GetTools(ctx context.Context) ([]sentinel.Tool, error) {
	query := `
		SELECT id, name, attributes, description
		FROM tool`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting tools: %w", err)
	}
	defer rows.Close()

	var tools []sentinel.Tool
	for rows.Next() {
		var tool sentinel.Tool
		var attributesJSON []byte
		if err := rows.Scan(&tool.Id, &tool.Name, &attributesJSON, &tool.Description); err != nil {
			return nil, fmt.Errorf("error scanning tool: %w", err)
		}

		// Parse the JSON attributes
		if err := json.Unmarshal(attributesJSON, &tool.Attributes); err != nil {
			return nil, fmt.Errorf("error parsing tool attributes: %w", err)
		}

		tools = append(tools, tool)
	}

	return tools, nil
}

func (s *PostgresqlStore) GetSupervisorFromToolID(ctx context.Context, id uuid.UUID) (*sentinel.Supervisor, error) {
	query := `
		SELECT s.id, s.description, s.created_at, s.type
		FROM supervisor s
		INNER JOIN tool_supervisor ts ON s.id = ts.supervisor_id
		INNER JOIN tool t ON ts.tool_id = t.id
		WHERE t.id = $1`

	var supervisor sentinel.Supervisor
	err := s.db.QueryRowContext(ctx, query, id).Scan(&supervisor.Id, &supervisor.Description, &supervisor.CreatedAt, &supervisor.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor: %w", err)
	}
	return &supervisor, nil
}

func (s *PostgresqlStore) CreateSupervisor(ctx context.Context, supervisor sentinel.Supervisor) (uuid.UUID, error) {
	if supervisor.Code == nil {
		return uuid.UUID{}, fmt.Errorf("can't create supervisor, code is required")
	}

	var id uuid.UUID

	// Try and find a supervisor with the same code
	query := `
		SELECT id
		FROM supervisor
		WHERE code = $1`

	var existingSupervisorId uuid.UUID
	err := s.db.QueryRowContext(ctx, query, supervisor.Code).Scan(&existingSupervisorId)
	if err != nil && err != sql.ErrNoRows {
		return uuid.UUID{}, fmt.Errorf("error checking if supervisor already exists: %w", err)
	}

	// If the supervisor already exists, just use the existing ID, else create a new one
	if existingSupervisorId != uuid.Nil {
		return existingSupervisorId, nil
	}

	id = uuid.New()

	query = `
		INSERT INTO supervisor (id, description, created_at, type, code)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = s.db.ExecContext(ctx, query, id, supervisor.Description, supervisor.CreatedAt, supervisor.Type, supervisor.Code)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating supervisor: %w", err)
	}

	return id, nil
}

func (s *PostgresqlStore) CreateRun(ctx context.Context, run sentinel.Run) (uuid.UUID, error) {
	id := uuid.New()

	query := `
		INSERT INTO run (id, project_id, created_at)
		VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, query, id, run.ProjectId, run.CreatedAt)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating run: %w", err)
	}
	return id, nil
}

func (s *PostgresqlStore) CreateRunTool(ctx context.Context, runId uuid.UUID, tool sentinel.Tool) (uuid.UUID, error) {
	var id uuid.UUID

	if tool.Name == "" || tool.Description == "" || tool.Attributes == nil {
		return uuid.UUID{}, fmt.Errorf("can't create run tool, tool name, description, and attributes are required. Values: %+v", tool)
	}

	attributes, err := json.Marshal(tool.Attributes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error marshalling tool attributes: %w", err)
	}

	// Check if there is a tool already with the same name, description, and attributes
	query := `
		SELECT id
		FROM tool
		WHERE name = $1 AND description = $2 AND attributes = $3`

	var existingToolId uuid.UUID
	err = s.db.QueryRowContext(ctx, query, tool.Name, tool.Description, attributes).Scan(&existingToolId)
	if err != nil && err != sql.ErrNoRows {
		return uuid.UUID{}, fmt.Errorf("error checking if tool already exists: %w", err)
	}

	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// If the tool already exists, just use the existing ID, else create a new one
	if existingToolId != uuid.Nil {
		id = existingToolId
	} else {
		id = uuid.New()

		query = `
		INSERT INTO tool (id, name, description, attributes)
		VALUES ($1, $2, $3, $4)`

		_, err = tx.ExecContext(ctx, query, id, tool.Name, tool.Description, attributes)
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("error creating tool: %w", err)
		}
	}

	// Insert a connection between the run and the tool into run_tool
	query = `
		INSERT INTO run_tool (run_id, tool_id)
		VALUES ($1, $2)`

	_, err = tx.ExecContext(ctx, query, runId, id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating run tool: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error committing transaction: %w", err)
	}

	return id, nil
}

func (s *PostgresqlStore) CreateTool(ctx context.Context, tool sentinel.Tool) (uuid.UUID, error) {
	id := uuid.New()

	if tool.Name == "" || tool.Description == "" || tool.Attributes == nil {
		return uuid.UUID{}, fmt.Errorf("can't create tool, tool name, description, and attributes are required. Values: %+v", tool)
	}

	attributes, err := json.Marshal(tool.Attributes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error marshalling tool attributes: %w", err)
	}

	query := `
		INSERT INTO tool (id, name, description, attributes)
		VALUES ($1, $2, $3, $4)`

	_, err = s.db.ExecContext(ctx, query, id, tool.Name, tool.Description, attributes)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("error creating tool: %w", err)
	}

	return id, nil
}

func (s *PostgresqlStore) GetReviewToolRequests(ctx context.Context, id uuid.UUID) ([]sentinel.ToolRequest, error) {
	query := `
		SELECT id, reviewrequest_id, tool_id, message_id, arguments
		FROM toolrequest
		WHERE reviewrequest_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting review tool requests: %w", err)
	}
	defer rows.Close()

	var toolRequests []sentinel.ToolRequest
	for rows.Next() {
		var toolRequest sentinel.ToolRequest
		if err := rows.Scan(&toolRequest.Id, &toolRequest.ReviewRequestId, &toolRequest.ToolId, &toolRequest.MessageId, &toolRequest.Arguments); err != nil {
			return nil, fmt.Errorf("error scanning tool request: %w", err)
		}
		toolRequests = append(toolRequests, toolRequest)
	}
	return toolRequests, nil

}

func (s *PostgresqlStore) GetRun(ctx context.Context, projectId uuid.UUID, id uuid.UUID) (*sentinel.Run, error) {
	query := `
		SELECT id, project_id, created_at
		FROM run
		WHERE id = $1 AND project_id = $2`

	var run sentinel.Run
	err := s.db.QueryRowContext(ctx, query, id, projectId).Scan(&run.Id, &run.ProjectId, &run.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting run: %w", err)
	}

	return &run, nil
}

func (s *PostgresqlStore) AssignSupervisorToTool(ctx context.Context, supervisorID uuid.UUID, toolID uuid.UUID) error {
	query := `
		INSERT INTO tool_supervisor (tool_id, supervisor_id)
		VALUES ($1, $2)`

	_, err := s.db.ExecContext(ctx, query, toolID, supervisorID)
	if err != nil {
		return fmt.Errorf("error assigning supervisor to tool: %w", err)
	}

	return nil
}

func (s *PostgresqlStore) GetRunTools(ctx context.Context, id uuid.UUID) ([]sentinel.Tool, error) {
	query := `
		SELECT tool.id, tool.name, tool.description, tool.attributes
		FROM tool
		INNER JOIN run_tool ON tool.id = run_tool.tool_id
		WHERE run_tool.run_id = $1`

	rows, err := s.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting run tools: %w", err)
	}
	defer rows.Close()

	var tools []sentinel.Tool
	for rows.Next() {
		var tool sentinel.Tool
		if err := rows.Scan(&tool.Id, &tool.Name, &tool.Description, &tool.Attributes); err != nil {
			return nil, fmt.Errorf("error scanning tool: %w", err)
		}
		tools = append(tools, tool)
	}

	// If no rows were found, tools will be empty slice
	return tools, nil
}

func (s *PostgresqlStore) GetSupervisor(ctx context.Context, id uuid.UUID) (*sentinel.Supervisor, error) {
	query := `
		SELECT id, description, created_at, type
		FROM supervisor
		WHERE id = $1`

	var supervisor sentinel.Supervisor
	err := s.db.QueryRowContext(ctx, query, id).Scan(&supervisor.Id, &supervisor.Description, &supervisor.CreatedAt, &supervisor.Type)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error getting supervisor: %w", err)
	}

	return &supervisor, nil
}

func (s *PostgresqlStore) GetSupervisors(ctx context.Context) ([]sentinel.Supervisor, error) {
	query := `
		SELECT id, description, created_at, type
		FROM supervisor`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error getting supervisors: %w", err)
	}
	defer rows.Close()

	var supervisors []sentinel.Supervisor
	for rows.Next() {
		var supervisor sentinel.Supervisor
		if err := rows.Scan(&supervisor.Id, &supervisor.Description, &supervisor.CreatedAt, &supervisor.Type); err != nil {
			return nil, fmt.Errorf("error scanning supervisor: %w", err)
		}
		supervisors = append(supervisors, supervisor)
	}

	return supervisors, nil
}
