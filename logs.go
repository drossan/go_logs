package go_logs

import (
	"github.com/fatih/color"
	"log"
)

func FatalLog(message string) {
	color.Set(color.FgRed)
	saveLog(" 💣 "+message, "FATAL")
	color.Unset()
	log.Fatal(" 💣  " + message)
}

func ErrorLog(message string) {
	color.Set(color.FgRed)
	log.Println(" 🚨 ", message)
	color.Unset()
	saveLog(" 🚨 "+message, "ERROR")
}

func InfoLog(message string) {
	color.Set(color.FgYellow)
	log.Println(" ⚠️ ", message)
	color.Unset()
	saveLog(" ⚠️  "+message, "INFO")
}

func SuccessLog(message string) {
	color.Set(color.FgGreen)
	log.Println(" ✅ ", message)
	color.Unset()
	saveLog(" ✅ "+message, "SUCCESS")
}
