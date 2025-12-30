package db

import (
	"context"
	"errors"
	"toDoList/internal"
	"toDoList/internal/domain/task/taskerrors"
	"toDoList/internal/domain/task/taskmodels"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

type taskStorage struct {
	db PgxIface
}

func (ts *taskStorage) GetAllTasks(userID string) ([]taskmodels.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()
	rows, err := ts.db.Query(
		ctx,
		"SELECT id, userid, status, title, description, deleted FROM tasks where userid = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}

	var tasks []taskmodels.Task

	for rows.Next() {
		var task taskmodels.Task
		if err = rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Attributes.Status,
			&task.Attributes.Title,
			&task.Attributes.Description,
			&task.Deleted,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (ts *taskStorage) GetTaskByID(taskID string, userID string) (taskmodels.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	var task taskmodels.Task
	err := ts.db.QueryRow(ctx, "SELECT id, userid, status, title, description, deleted FROM tasks WHERE id = $1 AND userid = $2", taskID, userID).
		Scan(
			&task.ID,
			&task.UserID,
			&task.Attributes.Status,
			&task.Attributes.Title,
			&task.Attributes.Description,
			&task.Deleted,
		)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return taskmodels.Task{}, taskerrors.ErrFoundNothing
		}
		return taskmodels.Task{}, err
	}

	return task, nil
}

func (ts *taskStorage) AddTask(newTask taskmodels.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	_, err := ts.db.Exec(
		ctx,
		"INSERT INTO tasks (id, userid, status, title, description) VALUES ($1, $2, $3, $4, $5)",
		newTask.ID,
		newTask.UserID,
		newTask.Attributes.Status,
		newTask.Attributes.Title,
		newTask.Attributes.Description,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return taskerrors.ErrTaskIsAlreadyExist
			}
		}
		return err
	}
	return nil
}

func (ts *taskStorage) UpdateTaskAttributes(task taskmodels.Task) error {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	cmd, err := ts.db.Exec(
		ctx,
		"UPDATE tasks SET status = $1, title = $2, description = $3 WHERE id = $4",
		task.Attributes.Status,
		task.Attributes.Title,
		task.Attributes.Description,
		task.ID,
	)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return taskerrors.ErrFoundNothing
	}

	return nil
}

func (ts *taskStorage) DeleteTask(taskID string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	cmd, err := ts.db.Exec(ctx, "DELETE FROM tasks WHERE id = $1 AND userid = $2", taskID, userID)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return taskerrors.ErrFoundNothing
	}

	return nil
}

func (ts *taskStorage) MarkTaskToDelete(taskID string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
	defer cancel()

	cmd, err := ts.db.Exec(
		ctx,
		"UPDATE tasks SET deleted = true WHERE id = $1 AND userid = $2",
		taskID,
		userID,
	)

	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return taskerrors.ErrFoundNothing
	}

	return nil
}

func (ts *taskStorage) DeleteMarkedTasks() error {
	ctx, cancel := context.WithTimeout(context.Background(), internal.SecFive)
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
