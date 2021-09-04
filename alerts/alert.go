package alerts

import (
	"log"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/wintltr/login-api/models"
)

func WatchFile(filepath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	var old, current, dif string
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Println(err)
	}
	old = strings.Trim(string(data), "\n")
	done := make(chan bool)
	go func() {
		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				data, err := os.ReadFile(filepath)
				if err != nil {
					log.Println(err)
				}

				current = strings.Trim(string(data), "\n")
				//sometimes, the loop get repeated??, do this to avoid printing same string
				if dif != strings.ReplaceAll(current, old, "") && current != old {
					dif = strings.ReplaceAll(current, old, "")
					if dif != "" {
						dif = strings.Trim(dif, "\n")
						models.SendTelegramMessage("NEW ALERT FROM LOG " + filepath + ": \n" + dif)
						//log.Println("Different:", dif)
						old = current
					}
				}

				//log.Println("event:", event)

				// if event.Op&fsnotify.Write == fsnotify.Write {
				// 	//log.Println("modified file:", event.Name)
				// }
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	//log.Println("#filepath added:" + filepath)
	err = watcher.Add(filepath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
