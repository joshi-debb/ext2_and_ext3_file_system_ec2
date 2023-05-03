package manager

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"bufio"
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

		respuesta := Cmds{
			Params: Search(tk, tokens, w, r),
		}
		jsonBytes, _ := json.Marshal(respuesta)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)

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

	router.HandleFunc("/reporte1", func(w http.ResponseWriter, r *http.Request) {
		pdfFile, err := os.Open(user.return_disk())
		if err != nil {
			http.Error(w, "No existe el archivo", 404)
			return
		}
		defer pdfFile.Close()
		stat, err := pdfFile.Stat()
		if err != nil {
			http.Error(w, "Error", 500)
			return
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=archivo.pdf")
		w.Header().Set("Content-Length", fileSize)

		io.Copy(w, pdfFile)

	}).Methods("POST")

	router.HandleFunc("/reporte2", func(w http.ResponseWriter, r *http.Request) {
		pdfFile, err := os.Open(user.return_sb())
		if err != nil {
			http.Error(w, "No existe el archivo", 404)
			return
		}
		defer pdfFile.Close()
		stat, err := pdfFile.Stat()
		if err != nil {
			http.Error(w, "Error", 500)
			return
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=archivo.pdf")
		w.Header().Set("Content-Length", fileSize)

		io.Copy(w, pdfFile)

	}).Methods("POST")

	router.HandleFunc("/reporte3", func(w http.ResponseWriter, r *http.Request) {
		pdfFile, err := os.Open(user.return_tree())
		if err != nil {
			http.Error(w, "No existe el archivo", 404)
			return
		}
		defer pdfFile.Close()
		stat, err := pdfFile.Stat()
		if err != nil {
			http.Error(w, "Error", 500)
			return
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=archivo.pdf")
		w.Header().Set("Content-Length", fileSize)

		io.Copy(w, pdfFile)

	}).Methods("POST")

	router.HandleFunc("/scripts", func(w http.ResponseWriter, r *http.Request) {
		file, _, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		defer file.Close()

		respuestas := ""
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			tk := Token(scanner.Text())
			tokens := SplitTokens(scanner.Text())
			respuestas += Search(tk, tokens, w, r)
			respuestas += "\n"
		}

		respuesta := Cmds{
			Params: respuestas,
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
func Search(tk string, tks []string, w http.ResponseWriter, r *http.Request) string {

	if strings.HasPrefix(tk, "#") {
		return "comentario: " + tk
	} else {

		switch strings.ToLower(tk) {

		case "mkdisk":
			return disk.Mkdisk(tks)
		case "rmdisk":
			return disk.Rmdisk(tks)
		case "fdisk":
			return disk.Fdisk(tks)
		case "mount":
			return disk.Mount(tks)
		case "mkfs":
			return disk.Mkfs(tks)
		case "login":
			return user.Login(tks, disk)
		case "logout":
			return user.Logout()
		case "mkgrp":
			return user.Mkgrp(tks)
		case "rmgrp":
			return user.Rmgrp(tks)
		case "mkusr":
			return user.Mkusr(tks)
		case "rmusr":
			return user.Rmusr(tks)
		case "rep":
			return user.MakeReport(tks)
		case "pause":
			return "pause"
		default:
			return ""
		}
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
