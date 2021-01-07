package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection
var ctx = context.TODO()

// Task struct
type Task struct {
	ID        primitive.ObjectID `bson:_id`
	CreatedAt time.Time          `bson:created_at`
	UpdatedAt time.Time          `bson:updated_at`
	Text      string             `bson:text`
	Completed bool               `bson:completed`
}

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("tasker").Collection("tasks")
}

func main() {
	fmt.Println("Tasker")
	app := &cli.App{
		Name:  "tasker",
		Usage: "A Simple program to manage your tasks from a terminal",
		Action: func(c *cli.Context) error {
			tasks, err := getPending()
			if err != nil {
				if err == mongo.ErrNoDocuments {
					fmt.Print("There is no task here\nIf you want to add a new task run `add 'task'`")
					return nil
				}

				return err
			}

			printTasks(tasks)

			return nil
		},
		Commands: []cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add a task to your task list",
				Action: func(c *cli.Context) error {
					str := c.Args().First()
					if str == "" {
						return errors.New("Please add a non empty task")
					}
					task := &Task{
						ID:        primitive.NewObjectID(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
						Text:      str,
						Completed: false,
					}
					return createTask(task)
				},
			},
			{
				Name:    "all",
				Aliases: []string{"l"},
				Usage:   "list all your tasks",
				Action: func(c *cli.Context) error {
					tasks, err := getAllTasks()
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("There is no task here\nIf you want to add a new task run `add 'task'`")
							return nil
						}
						return err
					}
					printTasks((tasks))
					return nil
				},
			},
			{
				Name:    "done",
				Aliases: []string{"d"},
				Usage:   "Complete a task from your task list",
				Action: func(c *cli.Context) error {
					text := c.Args().First()
					return completeTask(text)
				},
			},
			{
				Name:    "finished",
				Aliases: []string{"f"},
				Usage:   "List your unfinished Tasks",
				Action: func(c *cli.Context) error {
					tasks, err := getFinished()
					if err != nil {
						if err == mongo.ErrNoDocuments {
							fmt.Print("There is no task here\nIf you want to add a new task run `add 'task'`")
							return nil
						}
						return err
					}
					printTasks((tasks))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func printTasks(tasks []*Task) {
	for i, v := range tasks {
		if v.Completed {
			color.Green("%d: %s", i+1, v.Text)
		} else {
			color.Yellow("%d: %s", i+1, v.Text)
		}

	}
}

func createTask(task *Task) error {
	_, err := collection.InsertOne(ctx, task)
	fmt.Println("Task created")
	return err
}

func getAllTasks() ([]*Task, error) {
	filter := bson.D{{}}
	return filterTasks(filter)
}

func filterTasks(filter interface{}) ([]*Task, error) {
	var tasks []*Task

	cur, err := collection.Find(ctx, filter)
	if err != nil {
		return tasks, err
	}

	for cur.Next(ctx) {
		var t Task
		err = cur.Decode(&t)
		if err != nil {
			return tasks, err
		}
		tasks = append(tasks, &t)
	}

	if err := cur.Err(); err != nil {
		return tasks, err
	}

	cur.Close(ctx)

	if len(tasks) == 0 {
		return tasks, mongo.ErrNoDocuments
	}

	return tasks, nil
}

func completeTask(text string) error {
	filter := bson.D{primitive.E{Key: "text", Value: text}}

	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "completed", Value: true}}}}

	t := &Task{}

	return collection.FindOneAndUpdate(ctx, filter, update).Decode(t)
}

func getPending() ([]*Task, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: false},
	}

	return filterTasks(filter)
}

func getFinished() ([]*Task, error) {
	filter := bson.D{
		primitive.E{Key: "completed", Value: true},
	}

	return filterTasks(filter)
}
