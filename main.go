package main

import (
	"fmt"
	"log"
	"os"
	database "todo/db"

	_ "github.com/glebarez/go-sqlite"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func main() {
	app := &cli.App{
		Name:  "todo",
		Usage: "Create and manage todo lists!",
		Commands: []*cli.Command{
			{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "List all of the todo lists you have",
				Args:      true,
				ArgsUsage: "<optional name of list>",
				Action: func(ctx *cli.Context) error {
					db := database.Open()
					defer db.Close()

					switch ctx.NArg() {
					case 0:
						table, err := database.ListTodos(db)
						if err != nil {
							log.Fatal(err)
						}

						table.Print()

					case 1:
						todo := ctx.Args().First()

						todoList, err := database.ListEntries(db, todo)
						if err != nil {
							log.Fatal(err)
						}

						todoList.Print()
					}
					return nil
				},
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Create a new top level todo list",
				Args:    true,
				Action: func(ctx *cli.Context) error {
					switch ctx.NArg() {
					case 0:
						return cli.Exit("Useage: todo create <title>", 2)

					case 1:
						title := ctx.Args().First()
						if title == "" {
							return cli.Exit("Useage: todo create <title>", 2)
						}

						db := database.Open()
						defer db.Close()

						err := database.CreateTodo(db, title)
						if err != nil {
							log.Fatal(err)
						}

					case 2:
						todo, entry := ctx.Args().Get(0), ctx.Args().Get(1)
						if todo == "" && entry != "" || todo == "" && entry == "" {
							return cli.Exit("Useage: todo create <title> <entry>", 2)
						}

						db := database.Open()
						defer db.Close()

						err := database.InsertEntry(db, todo, entry)
						if err != nil {
							log.Fatal(err)
						}

					default:
						return cli.Exit("Useage: todo create <title>", 2)
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initialModel() model {
	return model{
		choices:  []string{"Buy carrots", "Buy celery", "buy", "kohlrabi"},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		}
	}

	return m, nil
}

func (m model) View() string {
	// The header
	s := "What should we buy at the market?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
