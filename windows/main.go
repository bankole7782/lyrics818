package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	g143 "github.com/bankole7782/graphics143"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/saenuma/lyrics818/internal"
	"github.com/sqweek/dialog"
)

var colorObjCoords map[int]g143.Rect

func main() {
	rootPath, err := internal.GetRootPath()
	if err != nil {
		panic(err)
	}

	sampleLyricsPath := filepath.Join(rootPath, "bmtf.txt")
	os.WriteFile(sampleLyricsPath, internal.SampleLyricsFile, 0777)

	runtime.LockOSThread()

	internal.ObjCoords = make(map[int]g143.Rect)
	colorObjCoords = make(map[int]g143.Rect)
	internal.InputsStore = make(map[string]string)
	internal.InChannel = make(chan bool)

	window := g143.NewWindow(1000, 800, "lyrics818: a more comfortable lyrics video generator", false)
	internal.AllDraws(window)

	go func() {
		for {
			<-internal.InChannel
			ffPath := GetFFMPEGCommand()
			_, err := internal.MakeVideo(internal.InputsStore, ffPath)
			if err != nil {
				log.Println(err)
				return
			}

			internal.ClearAfterRender = true
		}
	}()

	// respond to the mouse
	window.SetMouseButtonCallback(mouseBtnCallback)
	// respond to mouse movement
	window.SetCursorPosCallback(internal.CursorPosCB)

	for !window.ShouldClose() {
		t := time.Now()
		glfw.PollEvents()

		if internal.ClearAfterRender {
			// clear the UI and redraw
			internal.InputsStore = make(map[string]string)
			internal.AllDraws(window)
			internal.DrawEndRenderView(window, internal.EmptyFrameNoInputs)
			time.Sleep(5 * time.Second)
			internal.AllDraws(window)

			// respond to the mouse
			window.SetMouseButtonCallback(mouseBtnCallback)
			// respond to mouse movement
			window.SetCursorPosCallback(internal.CursorPosCB)

			internal.ClearAfterRender = false
		}

		time.Sleep(time.Second/time.Duration(internal.FPS) - time.Since(t))
	}

}

func mouseBtnCallback(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Release {
		return
	}

	xPos, yPos := window.GetCursorPos()
	xPosInt := int(xPos)
	yPosInt := int(yPos)

	wWidth, wHeight := window.GetSize()

	// var widgetRS g143.Rect
	var widgetCode int

	for code, RS := range internal.ObjCoords {
		if g143.InRect(RS, xPosInt, yPosInt) {
			// widgetRS = RS
			widgetCode = code
			break
		}
	}

	if widgetCode == 0 {
		return
	}

	rootPath, _ := internal.GetRootPath()

	switch widgetCode {
	case internal.OpenWDBtn:
		internal.ExternalLaunch(rootPath)

	case internal.ViewLyricsBtn:
		sampleLyricsPath := filepath.Join(rootPath, "bmtf.txt")
		internal.ExternalLaunch(sampleLyricsPath)

	case internal.SelectLyricsBtn:
		filename, err := dialog.File().Filter("Lyrics File", "txt").Load()
		if filename == "" || err != nil {
			return
		}

		internal.InputsStore["lyrics_file"] = filename

		currentFrame := internal.RefreshInputsOnWindow(window, internal.EmptyFrameNoInputs)
		// send the frame to glfw window
		windowRS := g143.Rect{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, currentFrame, windowRS)
		window.SwapBuffers()

	case internal.FontFileBtn:
		filename, err := dialog.File().Filter("Font file", "ttf").Load()
		if filename == "" || err != nil {
			return
		}
		internal.InputsStore["font_file"] = filename
		currentFrame := internal.RefreshInputsOnWindow(window, internal.EmptyFrameNoInputs)
		// send the frame to glfw window
		windowRS := g143.Rect{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, currentFrame, windowRS)
		window.SwapBuffers()

	case internal.BgFileBtn:
		filename, err := dialog.File().Filter("PNG Image", "png").Load()
		if filename == "" || err != nil {
			return
		}

		internal.InputsStore["background_file"] = filename

		currentFrame := internal.RefreshInputsOnWindow(window, internal.EmptyFrameNoInputs)
		// send the frame to glfw window
		windowRS := g143.Rect{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, currentFrame, windowRS)
		window.SwapBuffers()

	case internal.MusicFileBtn:
		filename, err := dialog.File().Filter("MP3 Audio", "mp3").Load()
		if filename == "" || err != nil {
			return
		}
		internal.InputsStore["music_file"] = filename

		currentFrame := internal.RefreshInputsOnWindow(window, internal.EmptyFrameNoInputs)
		// send the frame to glfw window
		windowRS := g143.Rect{Width: wWidth, Height: wHeight, OriginX: 0, OriginY: 0}
		g143.DrawImage(wWidth, wHeight, currentFrame, windowRS)
		window.SwapBuffers()

	case internal.LyricsColorBtn:
		drawPickColors(window)
		window.SetMouseButtonCallback(pickColorsMouseCallback)
		window.SetCursorPosCallback(nil)

	case internal.OurSite:
		internal.ExternalLaunch("https://sae.ng")

	case internal.RenderBtn:
		if len(internal.InputsStore) != 5 {
			return
		}

		currentFrame := internal.RefreshInputsOnWindow(window, internal.EmptyFrameNoInputs)
		window.SetMouseButtonCallback(nil)
		window.SetKeyCallback(nil)
		window.SetCursorPosCallback(nil)
		internal.DrawRenderView(window, currentFrame)
		internal.InChannel <- true
	}

}
