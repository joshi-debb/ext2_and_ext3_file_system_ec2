package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var disk Disk
var user User

type Comand struct {
	Parametro string `json:"comando"`
}

func Cmd() {

	router := mux.NewRouter()
	enableCORS(router)

	router.HandleFunc("/Comands", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var comando Comand
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&comando)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tk := Token(comando.Parametro)
		tokens := SplitTokens(comando.Parametro)

		Search(tk, tokens, w, r)

	}).Methods("POST")

	http.ListenAndServe(":8080", router)
}

func enableCORS(router *mux.Router) {
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}).Methods(http.MethodOptions)
	router.Use(middlewareCors)
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization,Access-Control-Allow-Origin")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
			next.ServeHTTP(w, req)
		})
}

// función para buscar comando
func Search(tk string, tks []string, w http.ResponseWriter, r *http.Request) {
	switch strings.ToLower(tk) {
	case "mkdisk":
		hola := disk.Mkdisk(tks)
		respuesta := Comand{
			Parametro: hola,
		}
		jsonBytes, _ := json.Marshal(respuesta)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	case "rmdisk":
		disk.Rmdisk(tks)
	case "fdisk":
		disk.Fdisk(tks)
	case "mount":
		disk.Mount(tks)
	case "mkfs":
		disk.Mkfs(tks)
	case "login":
		user.Login(tks, disk)

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
				character := ""

				for i := 0; i < len(token); i++ {
					if token[i] != 10 {
						character += string(token[i])
					}
				}
				tokens = append(tokens, character)
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
