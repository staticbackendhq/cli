package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/staticbackendhq/backend-go"
	"github.com/staticbackendhq/core/model"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage scheduled tasks.",
	Long: fmt.Sprintf(`
%s

Create, update, list, inspect, and delete recurring scheduled tasks.
Tasks run on the backend scheduler using the base root user.
	`,
		clbold("Manage scheduled tasks"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scheduled tasks",
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		tasks, err := taskList(tok)
		if err != nil {
			printError("error listing tasks: %v", err)
			return
		}

		fmt.Printf("%s result(s)\n\n", clbold(len(tasks)))
		taskPrintList(tasks)
	},
}

var taskAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a scheduled task",
	Long: fmt.Sprintf(`
%s

Examples:

backend task add --name nightly-cleanup --type function --value cleanup --interval "0 2 * * *"
backend task add --name ping --type http --value https://example.com/hook --interval "*/15 * * * *" --meta '{"method":"POST","ct":"application/json","data":"{\"ok\":true}"}'
	`,
		clbold("Create a scheduled task"),
	),
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		task, ok := taskFromFlags(cmd, model.Task{})
		if !ok {
			return
		}

		created, err := taskAdd(tok, task)
		if err != nil {
			printError("error creating task: %v", err)
			return
		}

		printSuccess("Task %s created successfully", clbold(created.Name))
		fmt.Printf("Task ID: %s\n", clbold(created.ID))
	},
}

var taskUpdateCmd = &cobra.Command{
	Use:   "update id",
	Short: "Update a scheduled task",
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		if len(args) != 1 {
			printError("argument mismatch: one task id should be specified")
			return
		}

		current, err := taskInfo(tok, args[0])
		if err != nil {
			printError("error retrieving task: %v", err)
			return
		}

		task, ok := taskFromFlags(cmd, current)
		if !ok {
			return
		}

		updated, err := taskUpdate(tok, args[0], task)
		if err != nil {
			printError("error updating task: %v", err)
			return
		}

		printSuccess("Task %s updated successfully", clbold(updated.Name))
	},
}

var taskInfoCmd = &cobra.Command{
	Use:   "info id",
	Short: "Display scheduled task details",
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		if len(args) != 1 {
			printError("argument mismatch: one task id should be specified")
			return
		}

		task, err := taskInfo(tok, args[0])
		if err != nil {
			printError("error retrieving task: %v", err)
			return
		}

		taskPrintInfo(task)
	},
}

var taskDeleteCmd = &cobra.Command{
	Use:   "delete id",
	Short: "Delete a scheduled task",
	Run: func(cmd *cobra.Command, args []string) {
		if setBackend() == false {
			return
		}

		tok, ok := getRootToken()
		if !ok {
			return
		}

		if len(args) != 1 {
			printError("argument mismatch: one task id should be specified")
			return
		}

		if err := taskDelete(tok, args[0]); err != nil {
			printError("error deleting task: %v", err)
			return
		}

		printSuccess("Task %s has been deleted", clbold(args[0]))
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskAddCmd)
	taskCmd.AddCommand(taskUpdateCmd)
	taskCmd.AddCommand(taskInfoCmd)
	taskCmd.AddCommand(taskDeleteCmd)

	addTaskFlags(taskAddCmd)
	addTaskFlags(taskUpdateCmd)
}

func addTaskFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "task name")
	cmd.Flags().String("type", "", "task type: function, message, or http")
	cmd.Flags().String("value", "", "task value: function name, message type, or HTTP URL")
	cmd.Flags().String("interval", "", "cron interval in UTC")
	cmd.Flags().String("meta", "", "optional JSON metadata")
}

func taskFromFlags(cmd *cobra.Command, task model.Task) (model.Task, bool) {
	required := task.ID == ""

	if required || cmd.Flags().Changed("name") {
		name, _ := cmd.Flags().GetString("name")
		if len(name) == 0 {
			printError("missing parameter: the --name option is required")
			return task, false
		}
		task.Name = name
	}

	if required || cmd.Flags().Changed("type") {
		typ, _ := cmd.Flags().GetString("type")
		if len(typ) == 0 {
			printError("missing parameter: the --type option is required")
			return task, false
		}
		if !taskValidType(typ) {
			printError("invalid task type: must be function, message, or http")
			return task, false
		}
		task.Type = strings.ToLower(typ)
	}

	if required || cmd.Flags().Changed("value") {
		value, _ := cmd.Flags().GetString("value")
		if len(value) == 0 {
			printError("missing parameter: the --value option is required")
			return task, false
		}
		task.Value = value
	}

	if required || cmd.Flags().Changed("interval") {
		interval, _ := cmd.Flags().GetString("interval")
		if len(interval) == 0 {
			printError("missing parameter: the --interval option is required")
			return task, false
		}
		task.Interval = interval
	}

	if cmd.Flags().Changed("meta") {
		meta, _ := cmd.Flags().GetString("meta")
		if len(meta) > 0 && !json.Valid([]byte(meta)) {
			printError("invalid metadata: --meta must be valid JSON")
			return task, false
		}
		task.Meta = meta
	}

	return task, true
}

func taskValidType(typ string) bool {
	switch strings.ToLower(typ) {
	case model.TaskTypeFunction, model.TaskTypeMessage, model.TaskTypeHTTP:
		return true
	default:
		return false
	}
}

func taskList(token string) (tasks []model.Task, err error) {
	err = backend.Get(token, "/task", &tasks)
	return
}

func taskInfo(token, id string) (task model.Task, err error) {
	err = backend.Get(token, "/task/"+id, &task)
	return
}

func taskAdd(token string, task model.Task) (model.Task, error) {
	var created model.Task
	err := backend.Post(token, "/task", task, &created)
	return created, err
}

func taskUpdate(token, id string, task model.Task) (model.Task, error) {
	var updated model.Task
	err := backend.Post(token, "/task/"+id, task, &updated)
	return updated, err
}

func taskDelete(token, id string) error {
	return backend.Del(token, "/task/"+id)
}

func taskPrintList(tasks []model.Task) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "ID\tNAME\tTYPE\tVALUE\tINTERVAL\tLAST RUN\n")
	for _, task := range tasks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			task.ID,
			task.Name,
			task.Type,
			task.Value,
			task.Interval,
			taskFormatTime(task.LastRun),
		)
	}
	w.Flush()
}

func taskPrintInfo(task model.Task) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', tabwriter.DiscardEmptyColumns)
	fmt.Fprintf(w, "ID\tNAME\tTYPE\tVALUE\tINTERVAL\tLAST RUN\n")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
		task.ID,
		task.Name,
		task.Type,
		task.Value,
		task.Interval,
		taskFormatTime(task.LastRun),
	)
	w.Flush()

	if len(task.Meta) > 0 {
		fmt.Printf("\nMETA\n%s\n", task.Meta)
	}
}

func taskFormatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format("2006/01/02 15:04")
}
