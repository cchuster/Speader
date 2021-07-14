package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"image/color"
	"io/ioutil"
	"log"
	"strings"
	"time"
	"unicode"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var splitString []string

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Box Layout")
	myApp.Settings().SetTheme(theme.DarkTheme())

	sentenceEntry := widget.NewMultiLineEntry()
	sentenceEntry.SetPlaceHolder("Type sentence here...")
	// When sentence entry changes
	sentenceEntry.OnChanged = func(s string) {
		// Split string by seperators (space or "-"), and set splitString
		splitString = splitSeparators(s)
	}

	var fileContent string

	fileBtn := widget.NewButton("Speed Read From File", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			if r == nil {
				return
			}
			bytes, err := ioutil.ReadAll(r)
			if err != nil {
				log.Fatalln(err)
			}
			fileContent = string(bytes)
			sentenceEntry.SetText(fileContent)
		}, myWindow)
	})

	wpmLabel := widget.NewLabel("1 word per minute")
	wpmSlider := widget.NewSlider(1, 1500)
	wpmSlider.Step = 1
	wpmSlider.OnChanged = func(value float64) {
		if value == 1 {
			wpmLabel.SetText("1 word per minute")
		} else {
			wpmLabel.SetText(fmt.Sprintf("%d words per minute", int(value)))
		}
	}

	wordCountLabel := widget.NewLabel("1 word at a time")
	wordCountSlider := widget.NewSlider(1, 15)
	wordCountSlider.Step = 1
	wordCountSlider.OnChanged = func(value float64) {
		if value == 1 {
			wordCountLabel.SetText("1 word at a time")
		} else {
			wordCountLabel.SetText(fmt.Sprintf("%d words at a time", int(value)))
		}
	}

	backBtn := widget.NewButton("Back", nil)

	// Move here as it is not needed earlier
	content := widget.NewLabel("")

	sentenceView := container.New(layout.NewVBoxLayout(),
		layout.NewSpacer(),
		canvas.NewLine(color.White),
		container.New(layout.NewCenterLayout(), content),
		canvas.NewLine(color.White),
		layout.NewSpacer(),
		backBtn,
	)
	sentenceView.Hide()

	sentenceBtn := widget.NewButton("Speed Read", nil)

	settingsView := container.New(layout.NewVBoxLayout(),
		sentenceEntry,
		wpmLabel,
		wpmSlider,
		wordCountLabel,
		wordCountSlider,
		layout.NewSpacer(),
		sentenceBtn,
		fileBtn,
	)

	quit := make(chan bool)
	sentenceBtn.OnTapped = func() {
		sentenceView.Show()
		settingsView.Hide()

		go func() {
			delay := time.Duration(1. / wpmSlider.Value * float64(time.Minute))
			for i := int(wordCountSlider.Value); i < len(splitString); i += int(wordCountSlider.Value) {
				select {
				case <-quit:
					return
				default:
					startPos := int(wordCountSlider.Value)*(i/int(wordCountSlider.Value)) - int(wordCountSlider.Value)
					fmt.Println(i+int(wordCountSlider.Value), len(splitString))
					// If not enough words for another loop
					if i+int(wordCountSlider.Value) >= len(splitString) {
						// Display 4 words
						content.SetText(strings.Join(splitString[startPos:i], " "))
						// Wait required delay
						time.Sleep(delay)
						// Display everything after previous 4 words
						content.SetText(strings.Join(splitString[i:], " "))
					} else {
						// Otherwise just display 4 words
						content.SetText(strings.Join(splitString[startPos:i], " "))
					}
					time.Sleep(delay)
				}
			}
			// Wait for quit signal in case user wants to go back after loop completes
			<-quit
		}()
	}

	backBtn.OnTapped = func() {
		quit <- true
		sentenceView.Hide()
		settingsView.Show()
	}

	myWindow.SetContent(container.New(layout.NewMaxLayout(),
		sentenceView,
		settingsView,
	))
	myWindow.ShowAndRun()

}

// Move this here so it can be reused
func splitSeparators(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		if unicode.IsSpace(r) {
			return true
		} else if r == '-' {
			return true
		}
		return false
	})
}
