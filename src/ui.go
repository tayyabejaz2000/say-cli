package say

import (
	"fmt"
	"log"
	"time"

	"github.com/marcusolsson/tui-go"
)

type UI struct {
	ChatHistory *tui.Box
	Input       *tui.Entry

	ui tui.UI
}

func CreateUI(appConfig *Config, inputCallback func(string)) *UI {
	chatHistory := tui.NewVBox()

	chatScroll := tui.NewScrollArea(chatHistory)
	chatScroll.SetAutoscrollToBottom(true)

	chatBox := tui.NewVBox(chatScroll)
	chatBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	input.OnSubmit(func(entry *tui.Entry) {
		text := entry.Text()
		inputCallback(text)
		chatHistory.Append(tui.NewHBox(
			tui.NewLabel(time.Now().Format("15:04")),
			tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", appConfig.Name))),
			tui.NewLabel(text),
			tui.NewSpacer(),
		))
		input.SetText("")
	})

	name := tui.NewLabel(appConfig.Name + "> ")

	inputBox := tui.NewHBox(name, input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(chatBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	root := tui.NewHBox(chat)

	ui, _ := tui.New(root)
	ui.SetKeybinding("Esc", func() { ui.Quit() })

	return &UI{
		ChatHistory: chatHistory,
		Input:       input,
		ui:          ui,
	}
}

func (ui *UI) AddMessage(sender string, message string) {
	ui.ChatHistory.Append(tui.NewHBox(
		tui.NewLabel(time.Now().Format("15:04")),
		tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", sender))),
		tui.NewLabel(message),
		tui.NewSpacer(),
	))
	ui.ui.Repaint()
}

func (ui *UI) Run(exitCallback func()) {
	if err := ui.ui.Run(); err != nil {
		log.Fatalf("[Error: %s]: Something went wrong while running UI\n", err.Error())
	}
	exitCallback()
}
