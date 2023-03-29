package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	// os.Setenv("FYNE_THEME", "dark")
	rootPath, err := GetRootPath()
	if err != nil {
		panic(err)
	}

	myApp := app.New()
	myWindow := myApp.NewWindow("hanan: a more comfortable shell / terminal")
	myWindow.SetOnClosed(func() {
	})

	openWDBtn := widget.NewButton("Open Working Directory", func() {
		exec.Command("cmd", "/C", "start", rootPath).Run()
	})

	viewSampleBtn := widget.NewButton("View Sample Lyrics File", func() {
		box := container.NewScroll(container.NewMax(widget.NewLabel(string(sampleLyricsFile))))
		dialog.ShowCustom("Sample Lyrics File", "Close", box, myWindow)
	})

	topBar := container.NewHBox(openWDBtn, viewSampleBtn)
	formBox := container.NewPadded()
	outputsBox := container.NewVBox()

	getLyricsForm := func() *widget.Form {
		dirFIs, err := os.ReadDir(rootPath)
		if err != nil {
			panic(err)
		}
		files := make([]string, 0)
		for _, dirFI := range dirFIs {
			if !dirFI.IsDir() && !strings.HasPrefix(dirFI.Name(), ".") {
				files = append(files, dirFI.Name())
			}
		}

		lyricsInputForm := widget.NewForm()
		lyricsInputForm.Append("lyrics_file", widget.NewSelect(files, nil))
		lyricsInputForm.Append("font_file", widget.NewSelect(files, nil))
		lyricsInputForm.Append("background_file", widget.NewSelect(files, nil))
		lyricsInputForm.Append("music_file", widget.NewSelect(files, nil))
		colorEntry := widget.NewEntry()
		colorEntry.SetText("#666666")
		lyricsInputForm.Append("lyrics_color", colorEntry)
		lyricsInputForm.SubmitText = "Make Lyrics Video"
		lyricsInputForm.CancelText = "Close"
		lyricsInputForm.OnCancel = func() {
			os.Exit(0)
		}
		lyricsInputForm.OnSubmit = func() {
			outputsBox.Add(widget.NewLabel("Beginning"))
			inputs := getFormInputs(lyricsInputForm.Items)
			outFileName, err := makeLyrics(inputs)
			if err != nil {
				outputsBox.Add(widget.NewLabel("Error occured: " + err.Error()))
				return
			}
			openOutputButton := widget.NewButton("Open Video", func() {
				exec.Command("cmd", "/C", "start", filepath.Join(rootPath, outFileName)).Run()
			})
			outputsBox.Add(openOutputButton)
			outputsBox.Refresh()
		}

		return lyricsInputForm
	}

	refreshBtn := widget.NewButton("Refresh Files List", func() {
		lyricsForm := getLyricsForm()
		formBox.Add(lyricsForm)
		formBox.Refresh()
	})

	topBar.Add(refreshBtn)

	windowBox := container.NewVBox(
		topBar,
		widget.NewSeparator(),
		formBox, outputsBox,
	)

	lyricsForm := getLyricsForm()
	formBox.Add(lyricsForm)
	formBox.Refresh()

	myWindow.SetContent(windowBox)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.SetFixedSize(true)
	myWindow.ShowAndRun()
}

func getFormInputs(content []*widget.FormItem) map[string]string {
	inputs := make(map[string]string)
	for _, formItem := range content {
		entryWidget, ok := formItem.Widget.(*widget.Entry)
		if ok {
			inputs[formItem.Text] = entryWidget.Text
			continue
		}

		selectWidget, ok := formItem.Widget.(*widget.Select)
		if ok {
			inputs[formItem.Text] = selectWidget.Selected
		}
	}

	return inputs
}
