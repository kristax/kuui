package gui

import (
	"log"
)

func (u *ui) logLifecycle() {
	lc := u.App.Lifecycle()
	lc.SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	lc.SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	lc.SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	lc.SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}
