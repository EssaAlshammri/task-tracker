package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TaskRepository interface {
	Add(description string) (*Task, error)
	Update(id int, description string) error
	Delete(id int) error
	MarkInProgress(id int) error
	MarkDone(id int) error
	List(status string) ([]Task, error)
}

type JsonTaskRepository struct {
	filename string
	tasks    []Task
}

func newJsonTaskRepository(filename string) (*JsonTaskRepository, error) {
	repo := &JsonTaskRepository{
		filename: filename,
	}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JsonTaskRepository) load() error {
	file, err := os.ReadFile(r.filename)
	if os.IsNotExist(err) {
		emptyTasks := []Task{}
		data, err := json.Marshal(emptyTasks)
		if err != nil {
			return err
		}
		return os.WriteFile(r.filename, data, 0644)
	}

	return json.Unmarshal(file, &r.tasks)
}

func (r *JsonTaskRepository) save() error {
	data, err := json.Marshal(r.tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}
	return os.WriteFile(r.filename, data, 0644)

}

func (r *JsonTaskRepository) List(status string) ([]Task, error) {
	if status != "" {
		var filteredTasks []Task
		for _, t := range r.tasks {
			if t.Status == status {
				filteredTasks = append(filteredTasks, t)
			}
		}
		return filteredTasks, nil
	}
	return r.tasks, nil
}

func (r *JsonTaskRepository) Add(description string) (*Task, error) {

	task := Task{
		ID:          len(r.tasks) + 1,
		Description: description,
		Status:      "todo",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	r.tasks = append(r.tasks, task)

	return &task, r.save()

}

func (r *JsonTaskRepository) Update(id int, description string) error {
	for i, t := range r.tasks {
		if t.ID == id {
			r.tasks[i].Description = description
			r.tasks[i].UpdatedAt = time.Now().UTC()
		}
	}
	return r.save()
}

func (r *JsonTaskRepository) Delete(id int) error {
	for i, t := range r.tasks {
		if t.ID == id {
			r.tasks = append(r.tasks[:i], r.tasks[i+1:]...)
		}
	}

	return r.save()
}

func (r *JsonTaskRepository) MarkInProgress(id int) error {

	for i, t := range r.tasks {
		if t.ID == id {
			r.tasks[i].Status = "in-progress"
			r.tasks[i].UpdatedAt = time.Now().UTC()
		}
	}

	return r.save()
}

func (r *JsonTaskRepository) MarkDone(id int) error {

	for i, t := range r.tasks {
		if t.ID == id {
			r.tasks[i].Status = "done"
			r.tasks[i].UpdatedAt = time.Now().UTC()
		}
	}

	return r.save()
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	repo, err := newJsonTaskRepository("tasks.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing repository: %v\n", err)
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		err = add(args, repo)
	case "update":
		err = update(args, repo)
	case "delete":
		err = delete(args, repo)
	case "mark-in-progress":
		err = markInProgress(args, repo)
	case "mark-done":
		err = markDone(args, repo)
	case "list":
		err = list(args, repo)
	default:
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: task-tracker <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  add <description>")
	fmt.Println("  update <id> <description>")
	fmt.Println("  delete <id>")
	fmt.Println("  mark-in-progress <id>")
	fmt.Println("  mark-done <id>")
	fmt.Println("  list [status]")
}

func add(args []string, repo TaskRepository) error {
	if len(args) != 1 {
		return errors.New("help: task-tracker add \"task description\"")
	}

	task, err := repo.Add(args[0])
	if err != nil {
		return errors.New(err.Error())
	}
	fmt.Printf("task added successfully (ID: %d)\n", task.ID)
	return nil
}

func update(args []string, repo TaskRepository) error {
	if len(args) != 2 {
		return errors.New("help: task-tracker update 1 \"new task description\"")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("invalid id")
	}
	return repo.Update(id, args[1])
}

func delete(args []string, repo TaskRepository) error {
	if len(args) != 1 {
		return errors.New("help: task-tracker delete 1")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("invalid id")
	}
	return repo.Delete(id)
}

func markInProgress(args []string, repo TaskRepository) error {
	if len(args) != 1 {
		return errors.New("help: task-tracker mark-in-progress 1")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("invalid id")
	}
	return repo.MarkInProgress(id)
}

func markDone(args []string, repo TaskRepository) error {
	if len(args) != 1 {
		return errors.New("help: task-tracker mark-done 1")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return errors.New("invalid id")
	}
	return repo.MarkDone(id)
}

func list(args []string, repo TaskRepository) error {
	switch len(args) {
	case 0:
		tasks, err := repo.List("")
		if err != nil {
			return errors.New(err.Error())
		}
		for _, task := range tasks {
			fmt.Println(task.ID, task.Description, task.Status)
		}
		return nil
	case 1:
		if !slices.Contains([]string{"todo", "in-progress", "done"}, args[0]) {
			return errors.New("invalid status")
		}
		tasks, err := repo.List(args[0])
		if err != nil {
			return errors.New(err.Error())
		}
		for _, task := range tasks {
			fmt.Println(task.ID, task.Description, task.Status)
		}
		return nil
	}
	if len(args) > 1 {
		return errors.New("help: task-tracker list\ntask-cli list done\ntask-tracker list todo\ntask-tracker list in-progress")
	}
	return nil
}
