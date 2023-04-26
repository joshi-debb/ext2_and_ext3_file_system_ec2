package manager

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Cmd() {

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("[MIA]-Terminal:~$ ")
		cadena, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ocurrió un error:", err)
			return
		}

		tk := Token(cadena)
		tokens := SplitTokens(cadena)

		Search(tk, tokens)

	}
}

// función para buscar comando
func Search(tk string, tks []string) {
	switch strings.ToLower(tk) {
	case "mkdisk":
		Mkdisk(tks)
	case "rmdisk":
		Rmdisk(tks)

	default:
		fmt.Println("Comando no encontrado")
	}
}

func SplitTokens(txt string) []string {
	var tokens []string
	txt += " "
	token := ""
	state := 0
	for _, caracter := range txt {
		if state == 0 && caracter == '>' {
			state = 1
		} else if state == 0 && caracter == '#' {
			continue
		} else if state != 0 {
			if state == 1 {
				if caracter == '=' {
					state = 2
				} else if caracter == ' ' {
					continue
				}
			} else if state == 2 {
				if caracter == '"' {
					state = 3
				} else {
					state = 4
				}
			} else if state == 3 {
				if caracter == '"' {
					state = 4
				}
			} else if state == 4 && caracter == '"' {
				tokens = []string{}
				continue
			} else if state == 4 && caracter == ' ' {
				state = 0
				tokens = append(tokens, token)
				token = ""
				continue
			}
			token += string(caracter)
		}
	}

	return tokens
}

func Token(txt string) string {
	tkn := ""
	flag := false
	for _, caracter := range txt {
		if flag {
			if caracter == ' ' || caracter == '>' {
				break
			}
			tkn += string(caracter)
		} else if caracter != ' ' && !flag {
			if caracter == '#' {
				tkn = txt
				break
			} else {
				tkn += string(caracter)
				flag = true
			}
		}
	}
	return tkn
}
