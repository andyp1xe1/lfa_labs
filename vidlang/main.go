package main

import (
	"fmt"
	"os"
)

func main() {
	testScript := `
#!/bin/venv vidlang

videoTrack, audioTrack := open "video.mp4"
introVid, introAud := open "intro.mp4"
outro, outroAud := open "outro.mp4"

audioTrack |> volume 1.5

# stream is a global variable representing the latest pipeline result

audioTrack = [stream, *]
    |> crossfade 0.5
    |> pitch 1.5

videoTrack =*
    |> brightness 1.3
    |> contrast 1.1

sequence := [introAud, audioTrack, outroAud]
    |> map [i, el] ( el |> volume 0.5*i+1 )


trackline [intoVid, videoTrack, outro] sequence
export "final.mp4"
`

	l := lex(testScript)

	for {
		item := <-l.items
		if item.typ == itemEOF {
			break
		}
		fmt.Printf("%#v\n", item)
		if item.typ == itemError {
			os.Exit(1)
		}
	}
}
