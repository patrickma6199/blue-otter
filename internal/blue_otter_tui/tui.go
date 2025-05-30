package blue_otter_tui

// tui.go contains all functions related to the creation and management of the TUI client interface

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateUI(username string, roomName string) (rootLayout *tview.Flex,
    titleView *tview.TextView,
    chatView *tview.TextView,
    systemLogView *tview.TextView,
    inputField *tview.InputField) {

	tview.Styles.PrimitiveBackgroundColor = tcell.NewHexColor(0x06385b)
	tview.Styles.PrimaryTextColor = tcell.NewHexColor(0x60d79d)

    titleView = tview.NewTextView()
		titleView.SetText(`
    ____  __    __  ______   ____ _______________  ____     _____ __      ____
   / __ )/ /   / / / / __/  / __ /_  __/_  __/ __/ / __ \   / ___// /    /  _/
  / __  / /   / / / / /_   / / / // /   / / / /_  / /_/ /  / /   / /     / /  
/ /_/ / /___/ /_/ / __/  / /_/ // /   / / / __/ / _, _/  / /_  / /___  / / 
/_____/_____/\____/___/   \____//_/   /_/ /___/ /_/ |_|  /____//_____//___/  
                                                                            
																	CLIENT NODE - v0.1.0                                                                           
					`).
		SetTextAlign(tview.AlignCenter).
		SetWrap(true).
		SetBorder(true)

    chatView = tview.NewTextView()
	chatView.SetTitle(" Chat ").
		SetBorder(true)
        
	chatView.SetTextColor(tcell.ColorWhite)

    systemLogView = tview.NewTextView()
	systemLogView.SetTitle(" System Log ").
        SetBorder(true)

	systemLogView.SetTextColor(tcell.ColorWhite)


    inputField = tview.NewInputField().
        SetLabel(fmt.Sprintf("[%s] <%s>: ", roomName, username)).
        SetFieldWidth(0). // Allow for full-width text input
		SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorWhite)


    mainContent := tview.NewFlex().
        SetDirection(tview.FlexColumn).
        AddItem(chatView, 0, 2, false).  // "2" weight for chat
        AddItem(systemLogView, 0, 1, false) // "1" weight for system log

    rootLayout = tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(titleView, 12, 1, false).
        AddItem(mainContent, 0, 1, false).
        AddItem(inputField, 1, 1, true)

    return
}