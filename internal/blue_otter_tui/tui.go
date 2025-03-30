// package main

// func main() {

// }

package blue_otter_tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// CreateUI builds a layout with:
//  - A title view at the top (header)
//  - A chat view on the left
//  - A system log view on the right
//  - An input field at the bottom (footer)
//
// It returns the root layout plus references to each component so that
// you can update them (e.g., writing messages to the chat or system log).
func CreateUI(username string, roomName string) (rootLayout *tview.Flex,
    titleView *tview.TextView,
    chatView *tview.TextView,
    systemLogView *tview.TextView,
    inputField *tview.InputField) {

	tview.Styles.PrimitiveBackgroundColor = tcell.NewHexColor(0x06385b)
	tview.Styles.PrimaryTextColor = tcell.NewHexColor(0x60d79d)


    // Title at the top (header)
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

    // Chat area on the left
    chatView = tview.NewTextView()
	chatView.SetTitle(" Chat ").
		SetBorder(true)
        
	chatView.SetTextColor(tcell.ColorWhite)

    // System log area on the right
    systemLogView = tview.NewTextView()
	systemLogView.SetTitle(" System Log ").
        SetBorder(true)

	systemLogView.SetTextColor(tcell.ColorWhite)

    // User input bar at the bottom (footer)
    inputField = tview.NewInputField().
        SetLabel(fmt.Sprintf("[%s] <%s>: ", roomName, username)).
        SetFieldWidth(0). // Allow for full-width text input
		SetFieldBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorWhite)


    // Main horizontal split: Chat on left, System Log on right
    mainContent := tview.NewFlex().
        SetDirection(tview.FlexColumn).
        AddItem(chatView, 0, 2, false).  // "2" weight for chat
        AddItem(systemLogView, 0, 1, false) // "1" weight for system log

    // Top-level vertical layout:
    //  title (header)
    //  mainContent (chat + system log)
    //  inputField (footer)
    rootLayout = tview.NewFlex().
        SetDirection(tview.FlexRow).
        AddItem(titleView, 12, 1, false).
        AddItem(mainContent, 0, 1, false).
        AddItem(inputField, 1, 1, true)

    return
}