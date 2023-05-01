package manager

import (
	"encoding/json"
	"net/http"
	"strings"

	"bufio"
	"log"
	"os"

	"github.com/gorilla/mux"
)

var disk Disk
var user User

type Cmds struct {
	Params string `json:"cmds"`
}

type Login struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Pass string `json:"pass"`
}

func Cmd() {

	router := mux.NewRouter()
	enableCORS(router)

	router.HandleFunc("/cmds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cmdo Cmds
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&cmdo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tk := Token(cmdo.Params)
		tokens := SplitTokens(cmdo.Params)

		Search(tk, tokens, w, r)

	}).Methods("POST")

	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var logged Login
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&logged)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		responde := ""

		responde = user.CheckLogin(logged.Id, logged.Name, logged.Pass)

		respuesta := Cmds{
			Params: responde,
		}
		jsonBytes, _ := json.Marshal(respuesta)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)

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

// función para buscar cmdo
func Search(tk string, tks []string, w http.ResponseWriter, r *http.Request) {

	responde := ""

	if strings.HasPrefix(tk, "#") {
		responde = "comentario: " + tk
	} else {

		switch strings.ToLower(tk) {

		case "mkdisk":
			responde = disk.Mkdisk(tks)
		case "rmdisk":
			responde = disk.Rmdisk(tks)
		case "fdisk":
			responde = disk.Fdisk(tks)
		case "mount":
			responde = disk.Mount(tks)
		case "mkfs":
			responde = disk.Mkfs(tks)
		case "login":
			responde = user.Login(tks, disk)
		case "logout":
			responde = user.Logout()
		case "mkgrp":
			responde = user.Mkgrp(tks)
		case "rmgrp":
			responde = user.Rmgrp(tks)
		case "mkusr":
			responde = user.Mkusr(tks)
		case "rmusr":
			responde = user.Rmusr(tks)
		case "rep":
			responde = user.MakeReport(tks)
		case "pause":
			responde = "pause"
		case "execute":
			responde = execute(tks)

		default:
			responde = "Comando no encontrado"
		}
	}

	respuesta := Cmds{
		Params: responde,
	}
	jsonBytes, _ := json.Marshal(respuesta)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)

}

// función para ejecutar script
func execute(tks []string) string {

	var txt string

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "path" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				txt = token[1 : len(token)-1]
			} else {
				txt = token
			}
		}
	}

	filename := txt
	var lines []string
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	enableCORS(router)

	router.HandleFunc("/cmds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cmdo Cmds
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&cmdo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, line := range lines {
			txt := line
			tk := Token(txt)
			if txt != "" {
				txt = txt[len(tk)+1:]
				tks := SplitTokens(txt)
				Search(tk, tks, w, r)
			}
		}

	}).Methods("POST")

	return "Script ejecutado"

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
