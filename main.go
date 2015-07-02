package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/codegangsta/negroni"
	"github.com/hashtock/service-tools/serialize"

	"github.com/hashtock/auth/conf"
	"github.com/hashtock/auth/core"
	"github.com/hashtock/auth/storage"
	"github.com/hashtock/auth/webapp"
)

const (
	CommandMakeAdmin = "make_admin"
)

func main() {
	cfg := conf.GetConfig()

	mongoStorage, err := storage.NewMongoStorage(cfg.DB, cfg.DBName)
	if err != nil {
		log.Fatalln("Could not configure storage. ", err)
	}

	ok, err := AdminCommands(mongoStorage)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	} else if ok {
		return
	}

	handlerOptions := webapp.Options{
		Serializer:         &serialize.WebAPISerializer{},
		Storage:            mongoStorage,
		AppAddress:         cfg.AppAddress,
		GoogleClientID:     cfg.GoogleClientID,
		GoogleClientSecret: cfg.GoogleClientSecret,
		SessionSecret:      cfg.SessionSecret,
	}
	handler := webapp.Handlers(handlerOptions)

	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
	)
	n.UseHandler(handler)

	err = http.ListenAndServe(cfg.ServeAddress, n)
	if err != nil {
		log.Fatalln(err)
	}
}

func AdminCommands(adminStorage core.Administrator) (ok bool, err error) {
	flag.Parse()
	if flag.NArg() == 0 {
		return false, nil
	}

	cmd := flag.Arg(0)
	if cmd != CommandMakeAdmin {
		fmt.Fprintf(os.Stderr, "Command %#v not recognised\n", cmd)
		PrintUsage()
	} else if cmd == CommandMakeAdmin {
		if flag.NArg() != 2 {
			fmt.Fprintf(os.Stderr, "Command %#v accept exactly 1 argument. %v given\n", cmd, flag.NArg()-1)
			PrintUsage()
		} else if err = adminStorage.MakeUserAnAdmin(flag.Arg(1)); err != nil {
			return false, err
		}
	}

	return true, nil
}

func PrintUsage() {
	_, appName := filepath.Split(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", appName)
	fmt.Fprintf(os.Stderr, "        %s [command [arguments]]\n\n", appName)
	fmt.Fprintf(os.Stderr, "The commands are:\n\n")
	fmt.Fprintf(os.Stderr, "    [no command]\trun web app\n")
	fmt.Fprintf(os.Stderr, "    %s email\tmark user with given email as an admin\n", CommandMakeAdmin)
	flag.PrintDefaults()
}
