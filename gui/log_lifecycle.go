package gui

import (
	"log"
)

func (u *ui) logLifecycle() {
	u.app.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	u.app.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	u.app.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	u.app.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}
