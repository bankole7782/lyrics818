package main

import (
  "os"
	color2 "github.com/gookit/color"
	"github.com/bankole7782/zazabul"
	"fmt"
  "time"
  "path/filepath"
  "image"
  "image/draw"
  "github.com/golang/freetype"
  "golang.org/x/image/font"
  "github.com/golang/freetype/truetype"
  "golang.org/x/image/math/fixed"
  "github.com/lucasb-eyer/go-colorful"
  "strconv"
  "strings"
  "os/exec"
  "github.com/otiai10/copy"
  "github.com/disintegration/imaging"

)

const (
  DPI = 72.0
  SIZE = 80.0
  SPACING = 1.1
)

// 1366 - 130

func main() {
  rootPath, err := GetRootPath()
  if err != nil {
    panic(err)
    os.Exit(1)
  }

  if len(os.Args) < 2 {
		color2.Red.Println("Expecting a command. Run with help subcommand to view help.")
		os.Exit(1)
	}


	switch os.Args[1] {
	case "--help", "help", "h":
  		fmt.Println(`lyrics818 is a terminal program that creates lyrics videos.
It uses a constant picture for the background.

Directory Commands:
  pwd     Print working directory. This is the directory where the files needed by any command
          in this cli program must reside.

Main Commands:
  init    Creates a config file describing your video. Edit to your own requirements.
          The file from 'init' is expected for the 'run' command.

  run     Renders a project with the config created above. It expects a a config file generated from
          'init' command above.
          All files must be placed in the working directory.

  			`)

  	case "pwd":
  		fmt.Println(rootPath)

    case "init":
      var	tmplOfMethod1 = `// lyrics_file is the file that contains timestamps and lyrics chunks seperated by newlines.
// a sample can be found at https://sae.ng/static/bmtf.txt
lyrics_file:


// the font_file is the file of a ttf font that the text would be printed with.
// you could find a font on https://fonts.google.com
font_file:

// lyrics_color is the color of the rendered lyric. Example is #af1382
lyrics_color: #666666


// background_file is the background that would be used for this lyric video.
// the background_file must be a png or an mp4
// the background_file must be of dimensions (1366px x 768px)
// the framerate must be 60fps
// you can generate an mp4 from videos229
background_file:

// total_length: The duration of the songs in this format (mm:ss)
total_length:

// music_file is the song to add its audio to the video.
music_file:

  	`
  		configFileName := "s" + time.Now().Format("20060102T150405") + ".zconf"
  		writePath := filepath.Join(rootPath, configFileName)

  		conf, err := zazabul.ParseConfig(tmplOfMethod1)
      if err != nil {
      	panic(err)
      }

      err = conf.Write(writePath)
      if err != nil {
        panic(err)
      }

      fmt.Printf("Edit the file at '%s' before launching.\n", writePath)


    case "run":
    	if len(os.Args) != 3 {
    		color2.Red.Println("The run command expects a file created by the init command")
    		os.Exit(1)
    	}

    	confPath := filepath.Join(rootPath, os.Args[2])

    	conf, err := zazabul.LoadConfigFile(confPath)
    	if err != nil {
    		panic(err)
    		os.Exit(1)
    	}

    	for _, item := range conf.Items {
    		if item.Value == "" {
    			color2.Red.Println("Every field in the launch file is compulsory.")
    			os.Exit(1)
    		}
    	}


      outName := "s" + time.Now().Format("20060102T150405")
      totalSeconds := timeFormatToSeconds(conf.Get("total_length"))
      lyricsObject := parseLyricsFile(filepath.Join(rootPath, conf.Get("lyrics_file")))
      renderPath := filepath.Join(rootPath, outName)
      os.MkdirAll(renderPath, 0777)

      // get the right ffmpeg command
      begin := os.Getenv("SNAP")
      command := "ffmpeg"
      if begin != "" && ! strings.HasPrefix(begin, "/snap/go/") {
        command = filepath.Join(begin, "bin", "ffmpeg")
      }



      if filepath.Ext(conf.Get("background_file")) == ".png" {

        var lastSeconds int
        startedPrinting := false
        firstFrame := false

        for seconds := 0; seconds < totalSeconds; seconds++ {

          if startedPrinting == false {
            _, ok := lyricsObject[seconds]
            if ! ok {
              fileHandle, err := os.Open(filepath.Join(rootPath, conf.Get("background_file")))
              if err != nil {
                panic(err)
              }
              img, _, err := image.Decode(fileHandle)
              if err != nil {
                panic(err)
              }
              writeManyImagesToDisk(img, renderPath, seconds)
            } else {
              startedPrinting = true
              firstFrame = true
              lastSeconds = seconds
            }

          } else {

            img := writeLyricsToImage(conf, lyricsObject[lastSeconds])

            if firstFrame == true {
              writeManyImagesToDisk(img, renderPath, lastSeconds )
              firstFrame = false
            }

            writeManyImagesToDisk(img, renderPath, seconds)
            _, ok := lyricsObject[seconds]
            if ok {
              firstFrame = true
              lastSeconds = seconds
            }
          }

        }

        color2.Green.Println("Completed generating frames of your lyrics video")

        out, err := exec.Command(command, "-framerate", "24", "-i", filepath.Join(renderPath, "%d.png"),
          filepath.Join(renderPath, "tmp_" + outName + ".mp4")).CombinedOutput()
        if err != nil {
          fmt.Println(string(out))
          panic(err)
        }


      } else if filepath.Ext(conf.Get("background_file")) == ".mp4" {

        framesPath := filepath.Join(rootPath, "frames_" + outName)
        os.MkdirAll(framesPath, 0777)
        out, err := exec.Command(command, "-i", filepath.Join(rootPath, conf.Get("background_file")),
          "-r", "60/1", filepath.Join(framesPath, "%d.png")).CombinedOutput()
        if err != nil {
          fmt.Println(string(out))
          panic(err)
        }

        color2.Green.Println("Finished getting frames from your video")

        var lastSeconds int
        startedPrinting := false
        firstFrame := false

        for frameCount := 0; frameCount < (totalSeconds * 60); frameCount++ {

          seconds := frameCount / 60
          videoFramePath := getNextVideoFrame(framesPath)

          if startedPrinting == false {
            _, ok := lyricsObject[seconds]
            if ! ok {
              newPath := filepath.Join(renderPath, filepath.Base(videoFramePath) )
              copy.Copy(videoFramePath, newPath)
            } else {
              startedPrinting = true
              firstFrame = true
              lastSeconds = seconds
            }

          } else {
            img := writeLyricsToVideoFrame(conf, lyricsObject[lastSeconds], videoFramePath)

            if firstFrame == true {
              imaging.Save(img, filepath.Join(renderPath, strconv.Itoa(frameCount - 1) + ".png"))
              firstFrame = false
            }

            imaging.Save(img, filepath.Join(renderPath, strconv.Itoa(frameCount) + ".png"))
            _, ok := lyricsObject[seconds]
            if ok {
              firstFrame = true
              lastSeconds = seconds
            }
          }


        }


        color2.Green.Println("Completed generating frames of your lyrics video")

        out, err = exec.Command(command, "-framerate", "60", "-i", filepath.Join(renderPath, "%d.png"),
          filepath.Join(renderPath, "tmp_" + outName + ".mp4")).CombinedOutput()
        if err != nil {
          fmt.Println(string(out))
          panic(err)
        }

        os.RemoveAll(framesPath)

      } else {
        color2.Red.Println("Unsupported backround_file format: must be .png or .mp4")
        os.Exit(1)
      }


      out, err := exec.Command(command, "-i", filepath.Join(renderPath, "tmp_" + outName + ".mp4"),
        "-i", filepath.Join(rootPath, conf.Get("music_file")),
        filepath.Join(rootPath, outName + ".mp4") ).CombinedOutput()
      if err != nil {
        fmt.Println(string(out))
        panic(err)
      }

      // clearing temporary files
      os.RemoveAll(renderPath)

      color2.Green.Println("The video has been generated into: ", filepath.Join(rootPath, outName + ".mp4") )

  	default:
  		color2.Red.Println("Unexpected command. Run the cli with --help to find out the supported commands.")
  		os.Exit(1)
  	}

}


func writeManyImagesToDisk(img image.Image, renderPath string, seconds int) {
  for i := 1; i <= 24; i++ {
    out := (24 * seconds) + i
    outPath := filepath.Join(renderPath, strconv.Itoa(out) + ".png")
    imaging.Save(img, outPath)
  }
}


func writeLyricsToImage(conf zazabul.Config, text string) image.Image {
  rootPath, _ := GetRootPath()

  fileHandle, err := os.Open(filepath.Join(rootPath, conf.Get("background_file")))
  if err != nil {
    panic(err)
  }
  pngData, _, err := image.Decode(fileHandle)
  if err != nil {
    panic(err)
  }
  b := pngData.Bounds()
  img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
  draw.Draw(img, img.Bounds(), pngData, b.Min, draw.Src)

  lyricsColor, _ := colorful.Hex(conf.Get("lyrics_color"))
  fg := image.NewUniform(lyricsColor)

  fontBytes, err := os.ReadFile(filepath.Join(rootPath, conf.Get("font_file")))
  if err != nil {
    panic(err)
  }
  fontParsed, err := freetype.ParseFont(fontBytes)
  if err != nil {
    panic(err)
  }

  c := freetype.NewContext()
  c.SetDPI(DPI)
  c.SetFont(fontParsed)
  c.SetFontSize(SIZE)
  c.SetClip(img.Bounds())
  c.SetDst(img)
  c.SetSrc(fg)
  c.SetHinting(font.HintingNone)

  texts := strings.Split(text, "\n")

  finalTexts := make([]string, 0)
  for _, txt := range texts {
    wrappedTxts := wordWrap(conf, txt, 1366 - 130)
    finalTexts = append(finalTexts, wrappedTxts...)
  }

  if len(finalTexts) > 7 {
    color2.Red.Println("Shorten the following text for it to fit this video:")
    color2.Red.Println()
    for _, t := range strings.Split(text, "\n") {
      color2.Red.Println("    ", t)
    }

    os.Exit(1)
  }

  // Draw the text.
  pt := freetype.Pt(80, 50+int(c.PointToFixed(SIZE)>>6))
  for _, s := range finalTexts {
    _, err = c.DrawString(s, pt)
    if err != nil {
      panic(err)
    }
    pt.Y += c.PointToFixed(SIZE * SPACING)
  }

  return img
}


func wordWrap(conf zazabul.Config, text string, writeWidth int) []string {
  rootPath, _ := GetRootPath()

  rgba := image.NewRGBA(image.Rect(0, 0, 1366, 768))

  fontBytes, err := os.ReadFile(filepath.Join(rootPath, conf.Get("font_file")))
  if err != nil {
    panic(err)
  }
  fontParsed, err := freetype.ParseFont(fontBytes)
  if err != nil {
    panic(err)
  }


  fontDrawer := &font.Drawer{
    Dst: rgba,
    Src: image.Black,
    Face: truetype.NewFace(fontParsed, &truetype.Options{
      Size: SIZE,
      DPI: DPI,
      Hinting: font.HintingNone,
    }),
  }

  widthFixed := fixed.I(writeWidth)

  strs := strings.Fields(text)
  outStrs := make([]string, 0)
  var tmpStr string
  for i, oneStr := range strs {
    var aStr string
    if i == 0 {
      aStr = oneStr
    } else {
      aStr += " " + oneStr
    }

    tmpStr += aStr
    if fontDrawer.MeasureString(tmpStr) >= widthFixed {
      outStr := tmpStr[ : len(tmpStr) - len(aStr) ]
      tmpStr = oneStr
      outStrs = append(outStrs, outStr)
    }
  }
  outStrs = append(outStrs, tmpStr)

  return outStrs
}


var currentFrame int
func getNextVideoFrame(framesPath string) string {
  currentFrame += 1
  currentFramePath := filepath.Join(framesPath, strconv.Itoa(currentFrame) + ".png")
  if DoesPathExists(currentFramePath) {
    return currentFramePath
  } else {
    currentFrame = 1
    return filepath.Join(framesPath, strconv.Itoa(currentFrame) + ".png")
  }
}


func writeLyricsToVideoFrame(conf zazabul.Config, text, videoFramePath string) image.Image {
  rootPath, _ := GetRootPath()

  pngData, err := imaging.Open(videoFramePath)

  b := pngData.Bounds()
  img := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
  draw.Draw(img, img.Bounds(), pngData, b.Min, draw.Src)

  lyricsColor, _ := colorful.Hex(conf.Get("lyrics_color"))
  fg := image.NewUniform(lyricsColor)

  fontBytes, err := os.ReadFile(filepath.Join(rootPath, conf.Get("font_file")))
  if err != nil {
    panic(err)
  }
  fontParsed, err := freetype.ParseFont(fontBytes)
  if err != nil {
    panic(err)
  }

  c := freetype.NewContext()
  c.SetDPI(DPI)
  c.SetFont(fontParsed)
  c.SetFontSize(SIZE)
  c.SetClip(img.Bounds())
  c.SetDst(img)
  c.SetSrc(fg)
  c.SetHinting(font.HintingNone)

  texts := strings.Split(text, "\n")

  finalTexts := make([]string, 0)
  for _, txt := range texts {
    wrappedTxts := wordWrap(conf, txt, 1366 - 130)
    finalTexts = append(finalTexts, wrappedTxts...)
  }

  if len(finalTexts) > 7 {
    color2.Red.Println("Shorten the following text for it to fit this video:")
    color2.Red.Println()
    for _, t := range strings.Split(text, "\n") {
      color2.Red.Println("    ", t)
    }

    os.Exit(1)
  }

  // Draw the text.
  pt := freetype.Pt(80, 50+int(c.PointToFixed(SIZE)>>6))
  for _, s := range finalTexts {
    _, err = c.DrawString(s, pt)
    if err != nil {
      panic(err)
    }
    pt.Y += c.PointToFixed(SIZE * SPACING)
  }

  return img
}
