package task

import (
	"aws/internal/common/domain"
	"github.com/jackc/pgx"
	"strings"
)

const errDuplicate = "duplicate key value violates unique constraint \"task_instance_id_uindex\""

type Repo struct {
	connPool *pgx.ConnPool
}

func NewTaskRepo(connPool *pgx.ConnPool) *Repo {
	t := new(Repo)
	t.connPool = connPool
	return t
}

func (r *Repo) Find() ([]domain.Task, error) {
	sql := "SELECT instance_id, instance_name, time_to_start, time_to_end FROM task;"
	rows, err := r.connPool.Query(sql)
	if err != nil {
		return nil, err
	}

	if rows.Err() != nil {
		return nil, err
	}
	var task domain.Task
	var tasks []domain.Task
	for rows.Next() {
		err = rows.Scan(
			&task.Id,
			&task.Name,
			&task.Start,
			&task.Stop,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, err
}

func (r *Repo) FindById(instanceId string) (domain.Task, error) {
	var task domain.Task

	sql := "SELECT instance_id, instance_name, time_to_start, time_to_end FROM task WHERE instance_id = $1;"
	row := r.connPool.QueryRow(sql, instanceId)

	err := row.Scan(
		&task.Id,
		&task.Name,
		&task.Start,
		&task.Stop,
	)

	return task, err
}
func (r *Repo) CheckExist(instanceId string) error {
	var id int

	sql := "SELECT id FROM task WHERE instance_id = $1;"
	row := r.connPool.QueryRow(sql, instanceId)

	err := row.Scan(
		&id,
	)

	s := err.Error()
	if err != nil && s != "no rows in result set" {
		return err
	}

	return nil
}

func (r *Repo) Store(task domain.Task) error {
	sql := "INSERT INTO task (instance_id, instance_name, time_to_start, time_to_end) VALUES ($1, $2, $3, $4);"
	var data []interface{}
	data = append(data, task.Id, task.Name, task.Start, task.Stop)
	tx, err := r.connPool.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(sql, data...)
	if err != nil && !strings.Contains(err.Error(), errDuplicate) {
		return err
	}

	err = tx.Commit()
	if err != nil && !strings.Contains(err.Error(), errDuplicate) {
		return err
	}

	return nil
}
