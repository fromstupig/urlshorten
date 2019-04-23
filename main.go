package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/smapig/urlshorten/dac"
)

var mappings dac.URL

func init() {
	rand.Seed(time.Now().UnixNano())

	db := dac.ConnectToDB("./url.db")
	mappings = dac.URL{"mappings", "id INTEGER PRIMARY KEY AUTOINCREMENT, url TEXT NOT NULL, redirection TEXT NOT NULL UNIQUE, numberOfUses INTEGER", db}
	dac.Create(mappings, db)
}

func main() {
	removingShortcut := flag.String("d", "", "Remove redirection config.")
	listShortcut := flag.Bool("l", false, "List all redirection.")
	usageInfo := flag.Bool("h", false, "Usage info.")

	configureCommand := flag.NewFlagSet("configure", flag.ExitOnError)
	url := configureCommand.String("u", "", "URL.")
	shortcut := configureCommand.String("a", "", "URL redirection.")
	configureCommand.Usage = func() {
		fmt.Fprint(os.Stderr, "configure [options] [<args>]\n")
		configureCommand.PrintDefaults()
	}

	runCommand := flag.NewFlagSet("run", flag.ExitOnError)
	port := runCommand.String("p", "8080", "HTTP server will be running on a given port.")
	runCommand.Usage = func() {
		fmt.Fprint(os.Stderr, "run [options] [<args>]\n")
		runCommand.PrintDefaults()
	}

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: urlshorten [options] [<args>] | urlshorten <commands> [options] [<args>]\n")
		fmt.Fprint(os.Stderr, "Options are:\n")
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "Commands are:\n")
		configureCommand.Usage()
		runCommand.Usage()
	}

	flag.Parse()

	if *usageInfo {
		flag.Usage()
		return
	}

	if *removingShortcut != "" {
		mappings.RemoveShortcut(*removingShortcut)
		return
	}

	if *listShortcut {
		mappings.ListAll()
		return
	}

	if flag.NArg() == 0 {
		invalidCommand()
		return
	}

	switch flag.Args()[0] {
	case "configure":
		configureCommand.Parse(flag.Args()[1:])
	case "run":
		runCommand.Parse(flag.Args()[1:])
	default:
		invalidCommand()
		os.Exit(1)
	}

	if configureCommand.Parsed() {
		configure(*url, *shortcut)
		return
	}

	if runCommand.Parsed() {
		run(*port)
		return
	}
}

func invalidCommand() {
	fmt.Fprintf(os.Stderr, "INVALID COMMAND! \nRun urlshorten -h to know more.\n")
}

func configure(url string, shortcut string) {
	if url == "" {
		fmt.Fprint(os.Stderr, "URL must be provided.\n")
		return
	}

	if shortcut == "" {
		shortcut = randomString(8)
	}

	mappings.Insert(url, shortcut)
}

func run(port string) {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	shortcut := strings.TrimPrefix(r.URL.Path, "/")

	if shortcut == "" {
		w.Write([]byte("Welcome to url shorten! Let config your redirection to make your life more simple."))
	} else {
		url := mappings.GetURL(shortcut)

		if url == "" {
			w.Write([]byte(shortcut + " not found."))
		} else {
			http.Redirect(w, r, url, http.StatusSeeOther)
			defer mappings.LogRedirection(shortcut)
		}
	}
}

func randomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		n := rand.Intn(35)

		if n < 10 {
			bytes[i] = byte(48 + n)
		} else if n > 25 {
			bytes[i] = byte(65 + n - 10)
		} else {
			bytes[i] = byte(65 + n)
		}
	}
	return string(bytes)
}
