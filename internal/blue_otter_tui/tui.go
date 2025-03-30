package main

func main() {
    
}

// package blue_otter_tui

// import (
//     "context"
//     "fmt"
//     "log"

//     "github.com/gdamore/tcell"
//     "github.com/libp2p/go-libp2p-pubsub"
//     "github.com/rivo/tview"
// )

// // ChatMessage is your JSON struct for messages
// type ChatMessage struct {
//     Sender string `json:"sender"`
//     Text   string `json:"text"`
// }

// func runTUI(
//     ctx context.Context,
//     cancel context.CancelFunc,
//     quitCh chan struct{},
//     sub *pubsub.Subscription,
//     topic *pubsub.Topic,
//     username string,
// ) error {
//     // Initialize tview Application
//     app := tview.NewApplication()

//     // A text view for incoming messages
//     messagesView := tview.NewTextView().
//         SetDynamicColors(true).
//         SetScrollable(true).
//         SetRegions(true).
//         SetWrap(true).
//         SetBorder(true).
//         SetTitle(" Messages ")

//     // An input field for user typing
//     inputField := tview.NewInputField().
//         SetLabel("Type> ").
//         SetFieldWidth(0). // 0 = no fixed limit
//         SetDoneFunc(func(key tcell.Key) {
//             if key == tcell.KeyEnter {
//                 text := inputField.GetText()
//                 inputField.SetText("")

//                 if text == "/quit" {
//                     // signal the main goroutine to shut down
//                     close(quitCh)
//                     cancel()
//                     return
//                 }
//                 if text == "" {
//                     return
//                 }

//                 // Construct a ChatMessage with username
//                 msg := ChatMessage{Sender: username, Text: text}
//                 // Marshal to JSON and publish (example from your existing code)
//                 data, err := json.Marshal(msg)
//                 if err != nil {
//                     addLine(messagesView, "Error encoding message: %v", err)
//                     return
//                 }
//                 topic.Publish(ctx, data)
//             }
//         }).
//         SetBorder(true).
//         SetTitle(" Input ")

//     // Layout: top half = messages, bottom = input
//     flex := tview.NewFlex().
//         SetDirection(tview.FlexRow).
//         AddItem(messagesView, 0, 1, false). // 0 height => expand
//         AddItem(inputField, 3, 1, true)     // 3 lines tall

//     app.SetRoot(flex, true)

//     // Start a goroutine to read subscription messages
//     go func() {
//         for {
//             select {
//             case <-ctx.Done():
//                 return
//             default:
//                 m, err := sub.Next(ctx)
//                 if err != nil {
//                     // Subscription closed
//                     return
//                 }

//                 // Distinguish self vs others
//                 if m.ReceivedFrom == topic.(*pubsub.TopicImpl).Router().ID() {
//                     addLine(messagesView, "[You]: %s", string(m.Data))
//                 } else {
//                     // Attempt JSON parse
//                     var cm ChatMessage
//                     if err := json.Unmarshal(m.Data, &cm); err != nil {
//                         addLine(messagesView, "Msg from %s (unparsed): %s", m.ReceivedFrom, string(m.Data))
//                     } else {
//                         addLine(messagesView, "[%s]: %s", cm.Sender, cm.Text)
//                     }
//                 }
//             }
//         }
//     }()

//     // Run the TUI main loop
//     if err := app.Run(); err != nil {
//         log.Printf("tview run error: %v\n", err)
//     }

//     return nil
// }

// func addLine(view *tview.TextView, format string, a ...interface{}) {
//     view.Write([]byte(fmt.Sprintf(format+"\n", a...)))
// }