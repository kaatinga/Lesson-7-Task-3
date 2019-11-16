package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/marcusolsson/tui-go"
)

type post struct {
	username string
	message  string
	time     string
}

var posts = make([]post, 2, 100)
var name string

func main() {

	// Сообщение, пока ограничиваем 80 байтами
	nameMaxLenght := 20
	textMaxLenght := 60
	message := make([]byte,nameMaxLenght+textMaxLenght)

	// Переменная для хранения сообщений от сервера
	messages := make(chan string, 1)
	users := make(chan string, 1)

	// Добавляем дежурное сообщение в чат от системы
	posts[0] = post{username: "System", message: "Welcome to Kaatinga's chat! Press `Esc` button to exit", time: time.Now().Format("15:04")}
	posts[1] = post{username: "System", message: "You are connecting to the server...", time: time.Now().Format("15:04")}

	// Рисуем GUI. Левая панель с юзерами
	sidebar := tui.NewVBox()
	sidebar.SetTitle("Chat User List")
	sidebar.Insert(0, tui.NewLabel("There are no users yet"))
	sidebar.SetBorder(true)

	// Рисуем GUI. Правая панель, чат
	history := tui.NewVBox()

	for _, m := range posts {
		history.Append(tui.NewHBox(
			tui.NewLabel(m.time),
			tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", m.username))),
			tui.NewLabel(m.message),
			tui.NewSpacer(),
		))

	}

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)
	historyBox.SetTitle("Chat History")

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)
	inputBox.SetTitle("Enter you name and press `Enter` to enter to chat!")

	chat := tui.NewVBox(historyBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	root := tui.NewHBox(sidebar, chat)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

		// Ну и собственно пробуем соединиться
	go func() {
		time.Sleep(3*time.Second)

		conn, err := net.Dial("tcp", "localhost:9000")
		if err != nil {
			log.Fatalln(err)
		}
		defer conn.Close()

		// Добавляем дежурное сообщение в чат от системы (успех)
		history.Append(tui.NewHBox(
			tui.NewLabel(time.Now().Format("15:04")),
			tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", "System"))),
			tui.NewLabel("Connection is established!"),
			tui.NewSpacer(),
		))



		// Добавлялка сообщений на экран
		go func(out <-chan string, ui *tui.UI) {
			var takenName string
			var takenText string

			for {
				value := <-out
				takenName, takenText = DecodeByteSlice([]byte(value))
				// Refresh'илка чата
				(*ui).Update(func() {
					history.Append(tui.NewHBox(
						tui.NewLabel(time.Now().Format("15:04")),
						tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", takenName))),
						tui.NewLabel(takenText),
						tui.NewSpacer(),
					))
				})

			}
		}(messages, &ui)

		// Добавлялка-обновлялка юзеров на экране
		go func(out <-chan string, ui *tui.UI) {
			for {
				var readName string
				readName = <-users

				// Refresh'илка сайдбара
				(*ui).Update(func() {
					inputBox.SetTitle("Enter your message")
					sidebar.Remove(0)
					sidebar.Insert(0, tui.NewLabel(readName+"             "))
				})

			}
		}(users, &ui)

		// Запускаем читалку сообщений с сервера
		go func(in chan<- string, c net.Conn) {
			for {
				_, err = c.Read(message) // Ждём и вычитываем сообщение с сервера
				if err != nil {
					log.Println(err)
					return
				}
				in <- string(message)
			}
		}(messages, conn)

		input.OnSubmit(func(e *tui.Entry) {

			// Если поле ввода не пустое
			if e.Text() != "" {
				if name == "" {
					users <- e.Text()
					name = e.Text()
				} else {

					// Отправляем на сервер сообщение
					_, err := conn.Write(MakeByteSlice(name, e.Text(), nameMaxLenght, textMaxLenght))
					if err != nil {
						history.Append(tui.NewHBox(
							tui.NewLabel(time.Now().Format("15:04")),
							tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", "john"))),
							tui.NewLabel(err.Error()),
							tui.NewSpacer(),
						))
						time.Sleep(5 * time.Second)
						return
					}

				}
				input.SetText("") // Обнуляем поле ввода
			}

		})
		var stopper chan string
		wannaStopHere := <-stopper
		fmt.Println(wannaStopHere)
	}()

	// Выход из программы по клавише ESC
	ui.SetKeybinding("Esc", func() { ui.Quit() })

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}

func MakeByteSlice(string1, string2 string, maxLenString1, maxLenString2 int) (resultByteSlice []byte) {
	resultByteSlice = make([]byte, maxLenString1+maxLenString2)
	if len(string1) > maxLenString1 {
		string1 = string1[:maxLenString1]
	}
	if len(string2) > maxLenString2 {
		string2 = string2[:maxLenString2]
	}
	myBuffer := bytes.NewBuffer([]byte{byte(len(string1))})
	myBuffer.Write([]byte(string1))
	myBuffer.Write([]byte{byte(len(string2))})
	myBuffer.Write([]byte(string2))
	myBuffer.Read(resultByteSlice)
	return
}

func DecodeByteSlice(byteSlice []byte) (text1, text2 string) {
	nameLen1 := int(byteSlice[0])
	text1 = string(byteSlice[1 : nameLen1+1])
	nameLen2 := int(byteSlice[nameLen1+1])
	text2 = string(byteSlice[nameLen1+2 : nameLen1+2+nameLen2])
	return
}
