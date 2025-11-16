package db

import (
	"context"
	"errors"
	"time"
	"toDoList/internal/domain/task/task_errors"
	"toDoList/internal/domain/task/task_models"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type taskStorage struct {
	db *pgx.Conn
}

func (ts *taskStorage) GetAllTasks(userID string) ([]task_models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := ts.db.Query(ctx, "SELECT * FROM tasks where userid = $1", userID)
	if err != nil {
		return nil, err
	}

	var tasks []task_models.Task

	for rows.Next() {
		var task task_models.Task

		if err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Attributes.Status,
			&task.Attributes.Title,
			&task.Attributes.Description,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (ts *taskStorage) GetTaskByID(taskID string, userID string) (task_models.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var task task_models.Task
	err := ts.db.QueryRow(ctx, "SELECT * FROM tasks WHERE id = $1 AND userid = $2", taskID, userID).
		Scan(
			&task.ID,
			&task.UserID,
			&task.Attributes.Status,
			&task.Attributes.Title,
			&task.Attributes.Description,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return task_models.Task{}, task_errors.FoundNothingErr
		}
		return task_models.Task{}, err
	}

	return task, nil
}

func (ts *taskStorage) AddTask(newTask task_models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ts.db.Exec(ctx, "INSERT INTO tasks (id, userid, status, title, description) VALUES ($1, $2, $3, $4, $5)",
		newTask.ID, newTask.UserID, newTask.Attributes.Status, newTask.Attributes.Title, newTask.Attributes.Description)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return task_errors.ErrorTaskIsAlreadyExist
			}
		}
		return err
	}
	return nil
}

func (ts *taskStorage) UpdateTaskAttributes(task task_models.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd, err := ts.db.Exec(ctx, "UPDATE tasks SET status = $1, title = $2, description = $3 WHERE id = $4",
		task.Attributes.Status, task.Attributes.Title, task.Attributes.Description, task.ID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return task_errors.FoundNothingErr
	}

	return nil
}

func (ts *taskStorage) DeleteTask(taskID string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd, err := ts.db.Exec(ctx, "DELETE FROM tasks WHERE id = $1 AND userid = $2", taskID, userID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return task_errors.FoundNothingErr
	}

	return nil
}

func (ts *taskStorage) MarkTaskToDelete(taskID string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd, err := ts.db.Exec(ctx, "UPDATE tasks SET deleted = true WHERE id = $1 AND userid = $2", taskID, userID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return task_errors.FoundNothingErr
	}

	return nil
}

func (ts *taskStorage) DeleteMarkedTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := ts.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err = tx.Rollback(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Transaction rollback failed")
		}
	}(tx, ctx)

	_, err = tx.Prepare(ctx, "delete_tasks", "DELETE FROM tasks WHERE deleted = true")
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "delete_tasks")
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
