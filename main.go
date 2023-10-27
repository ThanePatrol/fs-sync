package main

import (
	"log"
	"os"
	"time"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
//	"github.com/go-git/go-git/v5/plumbing/transport"
//	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Pass in directory to watch

func main() {
	directory := os.Getenv("DIRECTORY")
	name := os.Getenv("NAME")
	email := os.Getenv("EMAIL")
	remote := os.Getenv("REMOTE")
	remoteURL := os.Getenv("REMOTEURL")
	username := os.Getenv("USERNAME")
	pat := os.Getenv("PAT")

	authen := &http.BasicAuth{
		Username: username,
		Password: pat,
	}

	repo, err := git.PlainOpen(directory)

	if err != nil {
		log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	worktree, err := repo.Worktree()

	ret := worktree.Pull(&git.PullOptions{
		Auth: authen,
		RemoteName: remote,
		RemoteURL: remoteURL,
	}) 
	log.Println(ret)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if !strings.HasPrefix(event.Name, ".") && event.Has(fsnotify.Write) {
					_, err := worktree.Add(".")
					if err != nil {
						log.Println("Error in adding file:", err)
					}
					log.Println("File added:", event.Name)
					commitMsg := "Auto commit of changes at " + time.Now().Format("2006-01-02 15:04:05") + " of " + event.Name
					commit, err := worktree.Commit(commitMsg, &git.CommitOptions{
						Author: &object.Signature {
							Name:  name,
							Email: email,
							When:  time.Now(),
						},
					})
					if err != nil {
						log.Println("Error committing changes:", err)
					} else {
						//TODO - better error handling
						repo.CommitObject(commit)
						err = repo.Push(&git.PushOptions{
							RemoteName: remote,
							RemoteURL: remoteURL,
							Auth: authen,
						})
						if err != nil {
							log.Println("Error pushing: ", err)
						}
					}

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Error:", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(err)
	}

	<-make(chan struct{})

}

