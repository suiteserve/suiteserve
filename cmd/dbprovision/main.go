package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/crypto/ssh/terminal"
	r "gopkg.in/rethinkdb/rethinkdb-go.v6"
	"io/ioutil"
	"net/url"
	"os"
	"syscall"
)

var (
	noCreateFlag = flag.Bool("nocreate", false,
		"Whether to skip creating the specified database and tables if they" +
		"don't already exist")
	passFileFlag = flag.String("pass", "",
		"The file at which to read the password for the new user; leaving "+
			"this empty will prompt for the password")
	urlFlag = flag.String("url", "admin@localhost:28015",
		"The RethinkDB connection URL at which to provision the database; "+
			"takes the form user[:pass]@host:port")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [options] user db tables...\n\nOptions:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 3 {
		flag.Usage()
		return
	}
	user, db, tables := flag.Arg(0), flag.Arg(1), flag.Args()[2:]
	userStr, dbStr := userStr(user), dbStr(db)
	connUrl, err := url.Parse("rethinkdb://" + *urlFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"-url must take the form user[:pass]@host:port: %v\n", err)
		return
	}

	var pass []byte
	if *passFileFlag == "" {
		fmt.Printf("Enter new password for user %s: ", userStr)
		if pass, err = terminal.ReadPassword(syscall.Stdin); err != nil {
			fmt.Fprintf(os.Stderr, "read password: %v\n", err)
			os.Exit(74)
		}
		fmt.Println()
	} else if pass, err = ioutil.ReadFile(*passFileFlag); err != nil {
		fmt.Fprintf(os.Stderr, "read password file: %v\n", err)
		os.Exit(74)
	}

	opts := r.ConnectOpts{
		Address:  connUrl.Host,
		Username: connUrl.User.Username(),
	}
	if pass, ok := connUrl.User.Password(); ok {
		opts.Password = pass
	}
	session, err := r.Connect(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to database: %v\n", err)
		os.Exit(1)
	}

	// create user
	err = r.DB("rethinkdb").Table("users").Insert(map[string]string{
		"id":       user,
		"password": string(pass),
	}).Exec(session)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create user: %v\n", err)
		os.Exit(1)
	}
	printPlus()
	fmt.Printf("Upserted user %s\n", userStr)

	// create database
	if !*noCreateFlag {
		dbList, err := r.DBList().Contains(db).Run(session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "list databases: %v\n", err)
			os.Exit(1)
		}
		var dbExists bool
		if err := dbList.One(&dbExists); err != nil {
			fmt.Fprintf(os.Stderr, "read list databases response: %v\n", err)
			os.Exit(1)
		}

		if dbExists {
			printDash()
			fmt.Printf("Database %s already exists, not creating...\n", dbStr)
		} else {
			if err := r.DBCreate(db).Exec(session); err != nil {
				fmt.Fprintf(os.Stderr, "create database %q: %v\n", db, err)
				os.Exit(1)
			}
			printPlus()
			fmt.Printf("Created database %s\n", dbStr)
		}
	}

	// grant user permissions
	for _, table := range tables {
		tableStr := tableStr(db, table)
		if !*noCreateFlag {
			tableList, err := r.DB(db).TableList().Contains(table).Run(session)
			if err != nil {
				fmt.Fprintf(os.Stderr, "list tables: %v\n", err)
				os.Exit(1)
			}
			var tableExists bool
			if err := tableList.One(&tableExists); err != nil {
				fmt.Fprintf(os.Stderr, "read list tables response: %v\n", err)
				os.Exit(1)
			}

			if tableExists {
				printDash()
				fmt.Printf("Table %s already exists, not creating...\n",
					tableStr)
			} else {
				err := r.DB(db).TableCreate(table).Exec(session)
				if err != nil {
					fmt.Fprintf(os.Stderr, "create table %s: %v\n",
						tableStr, err)
					os.Exit(1)
				}
				printPlus()
				fmt.Printf("Created table %s\n", tableStr)
			}
		}
		err = r.DB(db).Table(table).Grant(user, map[string]bool{
			"read":  true,
			"write": true,
		}).Exec(session)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"grant user r+w permissions on table %s: %v\n",
				tableStr, err)
			os.Exit(1)
		}
		printPlus()
		fmt.Printf("Granted user %s r+w permissions on table %s\n",
			userStr, tableStr)
	}
}

func printPlus() {
	color.New(color.Bold, color.FgHiGreen).Print("+ ")
}

func printDash() {
	color.New(color.Bold, color.FgHiBlack).Print("- ")
}

func dbStr(db string) string {
	return color.New(color.FgCyan).Sprint(db)
}

func tableStr(db, table string) string {
	return color.New(color.FgCyan).Sprintf("%s.%s", db, table)
}

func userStr(user string) string {
	return color.New(color.FgGreen).Sprint(user)
}
