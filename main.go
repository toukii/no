package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
	"github.com/everfore/exc"
	"github.com/toukii/bytes"
	"github.com/toukii/closestmatch"
	"github.com/toukii/goutils"
)

func init() {
	dic = make(map[string]*Note)
	bs := goutils.ReadFile(notesFile, false)
	if len(bs) <= 0 {
		notesFile = os.Getenv("HOME") + "/.notes.toml"
		bs = goutils.ReadFile(notesFile, false)
	}
	err := toml.Unmarshal(bs, &dic)
	if err != nil {
		log.Fatal(err)
	}

}

type Note struct {
	Val   string
	Exced bool
}

func (n *Note) String() string {
	return fmt.Sprintf("%s [exced: %+v]", n.Val, n.Exced)
}

var (
	dic map[string]*Note

	notesFile = ".notes.toml"
	noteFile  = os.Getenv("HOME") + "/.note"
)

func refresh() {
	wr := bytes.NewWriter(make([]byte, 0, 1024))
	err := toml.NewEncoder(wr).Encode(dic)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(goutils.ToString(wr.Bytes()))

	goutils.WriteFile(notesFile, wr.Bytes())
}

func main() {
	args := os.Args
	size := len(args)
	switch size {
	case 1:
		ListKeys()
	case 2:
		GetNote(args[1])
	default:
		SetNote(args[1:])
	}
}

func SetNote(args []string) {
	size := len(args)
	exced := false
	if args[1] == "-e" {
		exced = true
	}
	if size == 2 && exced || size == 1 {
		if _, ex := dic[args[0]]; !ex {
			return
		}
		delete(dic, args[0])
		refresh()
		return
	}
	var vals []string
	w := bytes.NewWriter(make([]byte, 0, 1024))
	if exced {
		vals = args[2:]
	} else {
		vals = args[1:]
	}
	space := ""
	for i, it := range vals {
		if i > 0 {
			space = " "
		}
		if it[0] != "-"[0] {
			if i == 0 || it[0] == "|"[0] || it[0] == "&"[0] || vals[i-1][0] != "-"[0] {
				w.Write(goutils.ToByte(fmt.Sprintf("%s%s", space, it)))
			} else {
				w.Write(goutils.ToByte(fmt.Sprintf(`%s'%s'`, space, it)))
			}
		} else {
			w.Write(goutils.ToByte(fmt.Sprintf("%s%s", space, it)))
		}
	}

	dic[args[0]] = &Note{
		Val:   goutils.ToString(w.Bytes()),
		Exced: exced,
	}
	fmt.Printf("[%s] ==> %s\n", args[0], dic[args[0]])
	refresh()
}

func ListKeys() {
	fmt.Println("**** keys ***** exced *******")
	for k, n := range dic {
		fmt.Printf("%-15s %t\n", k, n.Exced)
	}
}

func GetNote(key string) {
	keys := make([]string, 0, len(dic))
	for k, _ := range dic {
		keys = append(keys, k)
	}
	cm := closestmatch.New(keys, []int{1})
	note, ex := dic[key]
	if !ex {
		k2 := cm.Closest(key)
		fmt.Printf("%s ≈≈> %s\n", key, k2)
		note, ex = dic[k2]
	}
	if !ex {
		fmt.Println("note nil")
		return
	}
	if !note.Exced {
		switch runtime.GOOS {
		case "darwin":
			// 			exc.Bash(fmt.Sprintf(`cat>%s<<-EOF
			// %s`, noteFile, note.Val)).Debug(false).Execute()
			err := goutils.WriteFile(noteFile, []byte(note.Val))
			if err != nil {
				log.Fatal(err)
			}
			exc.Bash(fmt.Sprintf(`cat %s | pbcopy`, noteFile)).Debug(false).Execute()
			fmt.Println("command is in the clipbroad.")
		default:
			fmt.Printf("%s\n", note.Val)
		}

		return
	}

	exc.Bash(note.Val).Debug(true).Execute()
}
