package tui

import "github.com/charmbracelet/bubbletea"

func isQuit(msg tea.KeyMsg) bool {
	return msg.String() == "ctrl+c"
}

func isUp(msg tea.KeyMsg) bool {
	k := msg.String()
	return k == "up" || k == "k"
}

func isDown(msg tea.KeyMsg) bool {
	k := msg.String()
	return k == "down" || k == "j"
}

func isEnter(msg tea.KeyMsg) bool {
	return msg.String() == "enter"
}

func isSpace(msg tea.KeyMsg) bool {
	return msg.String() == " "
}

func isEsc(msg tea.KeyMsg) bool {
	return msg.String() == "esc"
}

func isLeft(msg tea.KeyMsg) bool {
	k := msg.String()
	return k == "left" || k == "h"
}

func isRight(msg tea.KeyMsg) bool {
	k := msg.String()
	return k == "right" || k == "l"
}

func isTab(msg tea.KeyMsg) bool {
	return msg.String() == "tab"
}

func isSlash(msg tea.KeyMsg) bool {
	return msg.String() == "/"
}
