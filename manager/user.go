package manager

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type User struct{}
type Usr struct {
	User     string
	Password string
	Id       string
	Uid      int
}

var path_disk string = ""
var path_sb string = ""
var path_tree string = ""

var logeado Usr
var estado bool = false

func (usr User) Login(tks []string, admin Disk) string {
	user := ""
	password := ""
	id := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "user" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				user = token[1 : len(token)-1]
			} else {
				user = token
			}
		} else if strings.ToLower(tk) == "pwd" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				password = token[1 : len(token)-1]
			} else {
				password = token
			}
		} else if strings.ToLower(tk) == "id" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				id = token[1 : len(token)-1]
			} else {
				id = token
			}
		} else {
			return "No se esperaba el parametro: " + tk
		}
	}

	disk = admin
	Superblock := SuperBlocks()
	var fileblock Fileblock
	particion := Partitions()

	if !estado {
		estado = true
	} else {
		return "Ya existe un usuario activo"
	}
	var paths string
	particion, err := disk.FindPartition(id, &paths)
	if err != nil {
		estado = false
		return "No se encontro la particion"
	}

	readfiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readfiles.Close()
	readfiles.Seek(int64(particion.Part_start), 0)
	binary.Read(readfiles, binary.LittleEndian, &Superblock)
	readfiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readfiles, binary.LittleEndian, &fileblock)

	var archivo string
	archivo += string(fileblock.B_content[:])
	list_users := usr.extraer(archivo, 10)
	var encontrado bool = false
	var correct_user bool = false
	var correct_password bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'U' || list_users[i][2] == 'u' {
			Users := usr.extraer(list_users[i], 44)
			for j := 0; j < len(Users); j++ {
				if Users[3] == user && Users[4] == password {
					encontrado = true
					logeado.User = Users[3]
					logeado.Password = Users[4]
					logeado.Id = id
					uid, _ := strconv.Atoi(string(Users[0]))
					logeado.Uid = uid
					break
				} else if Users[3] != user && Users[4] == password {
					correct_user = true
					break
				} else if Users[3] == user && Users[4] != password {
					correct_password = true
					break
				} else if Users[3] != user && Users[4] != password {
					correct_password = true
					correct_user = true
					break
				}

			}
		}
		if encontrado {
			break
		}

	}
	if !encontrado {
		estado = false
		if correct_user && !correct_password {
			return "No existe el usuario"
		} else if correct_password && !correct_user {
			return "Contraseña incorrecta"
		} else if correct_user && correct_password {
			return "Usuario y Contraseña incorrecta"
		}
	}

	return "Logueado con exito"

}

func (usr User) Logout() string {
	if logeado.User == "" {
		return "No existe un usuario logueado"
	}
	logeado = Usr{}
	estado = false
	return "Se ha cerrado sesion con exito"

}

func (usr User) Mkgrp(tks []string) string {
	name := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "name" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				name = token[1 : len(token)-1]
			} else {
				name = token
			}
		} else {
			return "No se esperaba el parametro: " + tk
		}
	}

	Superblock := SuperBlocks()
	var fileblock Fileblock
	particion := Partitions()
	if !(logeado.User == "root" && logeado.Password == "123") {
		return "Solo el usuario root puede crear grupos"
	}
	var paths string
	particion, err := disk.FindPartition(logeado.Id, &paths)
	if err != nil {
		return "No se encontro la particion"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.Part_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)

	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")
	list_users := usr.extraer(archivo, 10)
	var cont_grp int = 1
	var newcont_grp int = 0
	var encontrado bool = false
	var ya_esta bool = false
	var newarchivo string = ""
	var newecontrado bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'G' || list_users[i][2] == 'g' {
			Users := usr.extraer(list_users[i], 44)
			cont_grp++
			for j := 0; j < len(Users); j++ {
				if Users[0] != "0" && Users[2] == name {
					encontrado = true
					break
				} else if Users[0] == "0" && Users[2] == name {
					ya_esta = true
					newecontrado = true
					newcont_grp, _ = strconv.Atoi(string(list_users[i-1][0]))
					newcont_grp++
					newarchivo += strconv.Itoa(newcont_grp) + ",G," + name + "\n"
					break
				}
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if encontrado {
		return "El grupo ya existe"
	}
	if newecontrado {
		var bytes [64]byte
		copy(bytes[:], []byte(newarchivo))
		fileblock.B_content = bytes
		readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
		binary.Write(readFiles, binary.LittleEndian, &fileblock)
		return "Grupo creado con exito \n" + string(fileblock.B_content[:])
	}

	archivo += strconv.Itoa(cont_grp) + ",G," + name + "\n"
	var bytes [64]byte
	copy(bytes[:], []byte(archivo))
	fileblock.B_content = bytes
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)

	return "Grupo creado con exito \n" + string(fileblock.B_content[:])
}

func (usr User) extraer(txt string, tab byte) []string {
	var enviar []string = strings.Split(txt, string(tab))
	for _, v := range enviar {
		if v == "" {
			enviar = enviar[:len(enviar)-1]
		}
	}
	return enviar
}

func (usr User) Rmgrp(tks []string) string {

	name := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "name" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				name = token[1 : len(token)-1]
			} else {
				name = token
			}
		} else {
			return "No se esperaba el parametro: " + tk
		}
	}

	Superblock := SuperBlocks()
	var fileblock Fileblock
	particion := Partitions()
	if !(logeado.User == "root" && logeado.Password == "123") {
		return "Solo el usuario root puede eliminar grupos"
	}
	var paths string
	particion, err := disk.FindPartition(logeado.Id, &paths)
	if err != nil {
		return "No se encontro la particion"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.Part_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)

	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")
	list_users := usr.extraer(archivo, 10)
	var newarchivo string = ""
	var encontrado bool = false
	var ya_esta bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'G' || list_users[i][2] == 'g' {
			Users := usr.extraer(list_users[i], 44)
			for j := 0; j < len(Users); j++ {
				if Users[0] != "0" && Users[2] == name {
					encontrado = true
					ya_esta = true
					newarchivo += strconv.Itoa(0) + ",G," + name + "\n"
					break
				} else if Users[0] == "0" && Users[2] == name {
					return "El grupo ya esta eliminado"
				}
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if !encontrado {
		return "El grupo no existe"
	}

	var bytes [64]byte
	copy(bytes[:], []byte(newarchivo))
	fileblock.B_content = bytes
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)
	return "Grupo eliminado con exito \n" + string(fileblock.B_content[:])
}

func (usr User) Mkusr(tks []string) string {

	user := ""
	pwd := ""
	grp := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "user" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				user = token[1 : len(token)-1]
			} else {
				user = token
			}
		} else if strings.ToLower(tk) == "pwd" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				pwd = token[1 : len(token)-1]
			} else {
				pwd = token
			}
		} else if strings.ToLower(tk) == "grp" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				grp = token[1 : len(token)-1]
			} else {
				grp = token
			}
		} else {
			return "No se esperaba el parametro: " + tk
		}
	}

	Superblock := SuperBlocks()
	var fileblock Fileblock
	particion := Partitions()
	if !(logeado.User == "root" && logeado.Password == "123") {

		return "Solo el usuario root puede crear usuarios"
	}
	var paths string
	particion, err := disk.FindPartition(logeado.Id, &paths)
	if err != nil {
		return "No se encontro la particion"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.Part_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)
	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")
	list_users := usr.extraer(archivo, 10)
	var cont_user int = 0
	var ya_esta bool = false
	var validacion bool = false
	var newarchivo string = ""
	var newecontrado bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'G' {
			Users := usr.extraer(list_users[i], 44)
			if Users[0] != "0" && Users[2] == grp {
				validacion = true
				cont_user, _ = strconv.Atoi(Users[0])
			} else if Users[0] == "0" && Users[2] == grp {
				return "El grupo ya esta eliminado"
			}
		} else if list_users[i][2] == 'U' {
			Users := usr.extraer(list_users[i], 44)
			if Users[0] != "0" && Users[3] == user {
				return "El usuario ya existe"
			} else if Users[0] == "0" && Users[3] == user {
				ya_esta = true
				newecontrado = true
				newarchivo += strconv.Itoa(cont_user) + ",U," + grp + "," + user + "," + pwd + "\n"
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if !validacion {

		return "El grupo no existe"
	}
	if newecontrado {
		var bytes [64]byte
		copy(bytes[:], []byte(newarchivo))
		fileblock.B_content = bytes
		readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
		binary.Write(readFiles, binary.LittleEndian, &fileblock)
		return "Usuario creado con exito \n" + string(fileblock.B_content[:])
	}

	archivo += strconv.Itoa(cont_user) + ",U," + grp + "," + user + "," + pwd + "\n"
	var bytes [64]byte
	copy(bytes[:], []byte(archivo))
	fileblock.B_content = bytes
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)
	return "Usuario creado con exito \n" + string(fileblock.B_content[:])
}

func (usr User) Rmusr(tks []string) string {

	usuario := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "user" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				usuario = token[1 : len(token)-1]
			} else {
				usuario = token
			}
		} else {
			return "No se esperaba el parametro: " + tk
		}
	}

	Superblock := SuperBlocks()
	var fileblock Fileblock
	particion := Partitions()
	if !(logeado.User == "root" && logeado.Password == "123") {

		return "Solo el usuario root puede eliminar usuarios"
	}
	var paths string
	particion, err := disk.FindPartition(logeado.Id, &paths)
	if err != nil {
		return "No se encontro la particion"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.Part_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)

	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")

	list_users := usr.extraer(archivo, 10)
	var newarchivo string = ""
	var encontrado bool = false
	var ya_esta bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'U' {
			Users := usr.extraer(list_users[i], 44)
			for j := 0; j < len(Users); j++ {
				if Users[0] != "0" && Users[3] == usuario {
					encontrado = true
					ya_esta = true
					newarchivo += strconv.Itoa(0) + ",U," + Users[2] + "," + usuario + "," + Users[4] + "\n"
					break
				} else if Users[0] == "0" && Users[3] == usuario {
					return "El usuario ya esta eliminado"
				}
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if !encontrado {
		return "El usuario no existe"
	}

	var bytes [64]byte
	copy(bytes[:], []byte(newarchivo))
	fileblock.B_content = bytes
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)

	return "Usuario eliminado con exito \n" + string(fileblock.B_content[:])
}

func (usr User) CheckLogin(id string, name string, pass string) string {

	if id == logeado.Id && name == logeado.User && pass != logeado.Password {
		return "Password incorrecto"
	} else if id != logeado.Id && name == logeado.User && pass == logeado.Password {
		return "ID incorrecto"
	} else if id == logeado.Id && name != logeado.User && pass == logeado.Password {
		return "Usuario incorrecto"
	}

	return "Login exitoso"
}

func (usr User) MakeReport(tks []string) string {

	name := ""
	id := ""
	rute := ""
	repPath := ""
	fileDot := ""
	fileTxt := ""
	dirPath := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "name" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				name = token[1 : len(token)-1]
			} else {
				name = token
			}
		} else if strings.ToLower(tk) == "path" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				repPath = token[1 : len(token)-1]
			} else {
				repPath = token
			}

			//obtener el nombre del archivo .dot
			fileDot = repPath[0:strings.Index(repPath, ".")] + ".dot"

			//obtener el nombre del archivo .txt
			fileTxt = repPath[0:strings.Index(repPath, ".")] + ".txt"

			//obtener ruta de carpetas
			dirPath = repPath
			for i := len(dirPath) - 1; i >= 0; i-- {
				if repPath[i] == '/' {
					dirPath = dirPath[0:i]
					break
				}
			}
		} else if strings.ToLower(tk) == "id" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				id = token[1 : len(token)-1]
			} else {
				id = token
			}
		} else if strings.ToLower(tk) == "ruta" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				rute = token[1 : len(token)-1]
			} else {
				rute = token
			}
		} else {
			return "No se esperaba el parametro: " + tk
		}
	}

	//verificar si existe el directorio
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		cmds := "mkdir -p \"" + dirPath + "\""
		exec.Command("sh", "-c", cmds).Run()
	}

	var particiones Partition

	binPath := ""
	particiones, err := disk.FindPartition(id, &binPath)
	if err != nil {
		return "No se encontro la particion"
	}

	if strings.ToLower(name) == "disk" {
		user.DiskReport(repPath, fileDot, binPath)
		return "Reporte generado en: " + repPath
	} else if strings.ToLower(name) == "tree" {
		user.TreeReport(repPath, fileTxt, binPath, particiones)
		return "Reporte generado en: " + repPath
	} else if strings.ToLower(name) == "sb" {
		user.SBReport(repPath, fileTxt, binPath, particiones)
		return "Reporte generado en: " + repPath
	} else if strings.ToLower(name) == "file" {
		user.FileReport(repPath, fileTxt, binPath, rute, particiones)
		return "Reporte generado en: " + repPath
	} else {
		return "Tipo de reporte no valido"
	}
}

func (usr User) DiskReport(repPath string, fileDot string, binPath string) {

	archivo, _ := os.Open(binPath)
	defer archivo.Close()
	var read_MBR Mbr
	binary.Read(archivo, binary.LittleEndian, &read_MBR)

	var size_logics int32 = 0
	cont_logics := 0

	strGrafica := "digraph G{ \n graph [ dpi = \"800\" ]; \n node [shape = plaintext]; \n mbr [label = < \n "
	strGrafica += "<table  cellpadding='20' border='0' cellborder='1' cellspacing='0'>\n"
	strGrafica += "<tr>\n"
	strGrafica += "<td rowspan='2' height='200'><b>MBR</b></td> \n"

	strGrafica_aux := "<tr>\n"

	List_read := List_Partition(read_MBR)
	for i := 0; i < 4; i++ {
		if List_read[i].Part_status == '1' && List_read[i].Part_type == 'e' {
			list_ext := disk.getlogics(List_read[i], binPath)
			for _, ebr := range list_ext {
				size_logics += ebr.EBR_size
				if size_logics < List_read[i].Part_s {
					cont_logics += 2
					porcentaje := float64(ebr.EBR_size)/float64(read_MBR.MBR_size) - float64(unsafe.Sizeof(read_MBR))
					porcentaje = math.Round(porcentaje*10000.00) / 100.00
					strGrafica_aux += "<td><b>EBR</b></td> \n"
					strGrafica_aux += "<td><b>Logica</b> <br/>" + fmt.Sprintf("%v", porcentaje) + "% del disco</td>\n"
				}
			}
			if size_logics < List_read[i].Part_s {
				cont_logics += 1
				porcentaje := float64(list_ext[i].EBR_size-size_logics)/float64(read_MBR.MBR_size) - float64(unsafe.Sizeof(read_MBR))
				porcentaje = math.Round(porcentaje*10000.00) / 100.00
				strGrafica_aux += "<td><b>Libre</b> <br/>" + fmt.Sprintf("%v", porcentaje) + "% del disco</td>\n"
			}
		}
	}

	strGrafica_aux += "</tr>\n\n"

	var size_primaries int32 = 0

	for i := 0; i < 4; i++ {
		if List_read[i].Part_status == '1' {
			size_primaries += List_read[i].Part_s
		}
	}

	for i := 0; i < 4; i++ {
		if List_read[i].Part_status == '1' && List_read[i].Part_type == 'e' {
			cont_logic := strconv.Itoa(cont_logics)
			strGrafica += "<td colspan='" + cont_logic + "'> <b>Extendida</b> </td> \n"
		} else if List_read[i].Part_status == '1' && List_read[i].Part_type == 'p' {
			porcentaje := float64(List_read[i].Part_s)/float64(read_MBR.MBR_size) - float64(unsafe.Sizeof(read_MBR))
			porcentaje = math.Round(porcentaje*10000.00) / 100.00
			porcentajeInt := int(porcentaje * 100)
			porcentajeStr := strconv.Itoa(porcentajeInt)
			strGrafica += "<td rowspan='2'> <b>Primaria</b> <br/>" + porcentajeStr + "% del disco</td> \n"
		}
	}

	if size_primaries < read_MBR.MBR_size {
		libre := read_MBR.MBR_size - size_primaries
		resto := float64(libre)/float64(read_MBR.MBR_size) - float64(unsafe.Sizeof(read_MBR))
		resto = math.Round(resto*10000.00) / 100.00
		restoInt := int(resto * 100)
		restoStr := strconv.Itoa(restoInt)
		strGrafica += "<td rowspan='2'> <b>Libre</b> <br/>" + restoStr + "% del disco</td> \n"

	}

	strGrafica += "</tr>\n\n"
	strGrafica += strGrafica_aux
	strGrafica += "</table>>];\n}\n"

	// create and write to file
	file, err := os.Create(fileDot)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()
	_, err = file.WriteString(strGrafica)
	if err != nil {
		fmt.Println(err)
		return
	}

	// execute dot command
	dotCmd := exec.Command("dot", "-Tpdf", fileDot, "-o", repPath)
	err = dotCmd.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	// remove file
	err = os.Remove(fileDot)
	if err != nil {
		fmt.Println(err)
		return
	}

	path_disk = repPath

}

func (usr User) SBReport(repPath string, fileDot string, binPath string, partitions Partition) {

	archivo, _ := os.Open(binPath)
	defer archivo.Close()

	var super Superblock

	archivo.Seek(int64(partitions.Part_start), 0)
	binary.Read(archivo, binary.LittleEndian, &super)

	strGrafica := "digraph G{ \n graph [ dpi = \"800\" ]; \n node [shape = plaintext]; \n mbr [label = < \n"
	strGrafica += "<table border='0' cellborder='1' cellspacing='0'>\n"
	strGrafica += "<tr><td colspan = '2' ><b>Superblock</b></td></tr>\n"
	strGrafica += "<tr>\n <td><b>s_filesystem_type</b></td> <td><b>" + strconv.Itoa(int(super.S_filesystem_type)) + "</b></td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_inodes_count</td>\n<td>" + strconv.Itoa(int(super.S_inodes_count)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_free_inodes_count</td>\n<td>" + strconv.Itoa(int(super.S_free_inodes_count)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_blocks_count</td>\n<td>" + strconv.Itoa(int(super.S_blocks_count)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_free_blocks_count</td>\n<td>" + strconv.Itoa(int(super.S_free_blocks_count)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_mtime</td>\n<td>" + time.Unix(super.S_mtime, 0).Format("02/01/2006 15:04:05") + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_umtime</td>\n<td>" + time.Unix(super.S_umtime, 0).Format("02/01/2006 15:04:05") + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_mnt_count</td>\n<td>" + strconv.Itoa(int(super.S_mnt_count)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_magic</td>\n<td>" + strconv.Itoa(int(super.S_magic)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_inode_size</td>\n<td>" + strconv.Itoa(int(super.S_inode_size)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_block_size</td>\n<td>" + strconv.Itoa(int(super.S_block_size)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_first_ino</td>\n<td>" + strconv.Itoa(int(super.S_first_ino)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_first_blo</td>\n<td>" + strconv.Itoa(int(super.S_first_blo)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_bm_inode_start</td>\n<td>" + strconv.Itoa(int(super.S_bm_inode_start)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_bm_block_start</td>\n<td>" + strconv.Itoa(int(super.S_bm_block_start)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_inode_start</td>\n<td>" + strconv.Itoa(int(super.S_inode_start)) + "</td>\n </tr>\n"
	strGrafica += "<tr>\n <td>s_block_start</td>\n<td>" + strconv.Itoa(int(super.S_block_start)) + "</td>\n </tr>\n"
	strGrafica += "</table>>];"
	strGrafica += "\n\n}\n"

	// create and write to file
	file, err := os.Create(fileDot)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()
	_, err = file.WriteString(strGrafica)
	if err != nil {
		fmt.Println(err)
		return
	}

	// execute dot command
	dotCmd := exec.Command("dot", "-Tpdf", fileDot, "-o", repPath)
	err = dotCmd.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	// remove file
	err = os.Remove(fileDot)
	if err != nil {
		fmt.Println(err)
		return
	}

	path_sb = repPath

}

func (usr User) TreeReport(repPath string, fileDot string, binPath string, partitions Partition) {

	var super Superblock
	var inodes Inode = Inodes()
	aux, _ := os.OpenFile(binPath, os.O_RDWR, 0666)
	defer aux.Close()
	aux.Seek(int64(partitions.Part_start), 0)
	binary.Read(aux, binary.LittleEndian, &super)

	aux.Seek(int64(super.S_bm_inode_start), 0)
	bmInodo := make([]byte, super.S_inodes_count)
	binary.Read(aux, binary.LittleEndian, &bmInodo)

	aux.Seek(int64(super.S_bm_block_start), 0)
	bmBloque := make([]byte, super.S_blocks_count)
	binary.Read(aux, binary.LittleEndian, &bmBloque)

	aux.Seek(int64(super.S_inode_start), 0)
	binary.Read(aux, binary.LittleEndian, &inodes)
	var freeInodes int = GetInodes(super, binPath)

	strGrafica := "digraph G{ \n rankdir=LR; \n graph []; \n node [shape = plaintext]; \n"

	for i := 0; i < freeInodes; i++ {
		strGrafica += "inode" + strconv.Itoa(i) + " [label = <<table border='0' cellborder='1' cellspacing='0'>\n"
		strGrafica += "<tr><td colspan = '2' > <b> i-Nodo " + strconv.Itoa(i) + " </b></td></tr>\n"

		strGrafica += "<tr>\n <td><b>i_type</b></td> <td> <b>" + strconv.Itoa(int(inodes.I_type)) + "</b></td>\n </tr>\n"

		strGrafica += "<tr>\n <td>i_uid</td> <td>" + strconv.Itoa(int(inodes.I_uid)) + "</td>\n </tr>\n"
		strGrafica += "<tr>\n <td>i_gid</td> <td>" + strconv.Itoa(int(inodes.I_gid)) + "</td>\n </tr>\n"
		strGrafica += "<tr>\n <td>i_size</td> <td>" + strconv.Itoa(int(inodes.I_size)) + "</td>\n </tr>\n"
		strGrafica += "<tr>\n <td>i_atime</td> <td>" + time.Unix(inodes.I_atime, 0).Format("02/01/2006 15:04:05") + "</td>\n </tr>\n"
		strGrafica += "<tr>\n <td>i_ctime</td> <td>" + time.Unix(inodes.I_ctime, 0).Format("02/01/2006 15:04:05") + "</td>\n </tr>\n"
		strGrafica += "<tr>\n <td>i_mtime</td> <td>" + time.Unix(inodes.I_mtime, 0).Format("02/01/2006 15:04:05") + "</td>\n </tr>\n"

		for j := 0; j < 15; j++ {
			strGrafica += "<tr>\n <td>i_block_" + strconv.Itoa(j+1) + "</td> <td port=\"" + strconv.Itoa(j) + "\">" + strconv.Itoa(int(inodes.I_block[j])) + "</td>\n </tr>\n"
		}

		strGrafica += "<tr>\n <td><b>i_perm</b></td> <td><b>" + strconv.Itoa(int(inodes.I_perm)) + "</b></td>\n </tr>\n"
		strGrafica += "</table>>];\n"

		if inodes.I_type == 48 {
			for j := 0; j < 12; j++ {
				if inodes.I_block[j] != -1 {
					strGrafica += "inode" + strconv.Itoa(i) + ":" + strconv.Itoa(j) + "-> block" + strconv.Itoa(int(inodes.I_block[j])) + ":n\n"
					var foldertemp Folderblock
					aux.Seek(int64(super.S_block_start+(int32(unsafe.Sizeof(Folderblock{}))*inodes.I_block[j])), 0)
					binary.Read(aux, binary.LittleEndian, &foldertemp)
					strGrafica += "block" + strconv.Itoa(int(inodes.I_block[j])) + "  [label = <<table border='0' cellborder='1' cellspacing='0'>\n"
					strGrafica += "<tr><td colspan = '2' > <b> block " + strconv.Itoa(int(inodes.I_block[j])) + "</b></td></tr>\n"
					var aux string
					for k := 0; k < 4; k++ {
						aux += strings.TrimRight(string(foldertemp.B_content[k].B_name[:]), "\x00")
						strGrafica += "<tr>\n <td>" + aux + "</td>\n <td port=\"" + strconv.Itoa(k) + "\">" + strconv.Itoa(int(foldertemp.B_content[k].B_inodo)) + "</td>\n </tr>\n"
					}
					strGrafica += "</table>>];\n"

					for b := 0; b < 4; b++ {
						if foldertemp.B_content[b].B_inodo != -1 {
							es := strings.TrimRight(string(foldertemp.B_content[b].B_name[:]), "\x00")
							if !(es == "." || es == "..") {
								strGrafica += "block" + strconv.Itoa(int(inodes.I_block[j])) + ":" + strconv.Itoa(b) + " -> inode" + strconv.Itoa(int(foldertemp.B_content[b].B_inodo)) + ":n\n"
							}
						}
					}
				}
			}

		} else {
			for j := 0; j < 15; j++ {
				if inodes.I_block[j] != -1 {
					if i < 12 {
						strGrafica += "inode" + strconv.Itoa(i) + ":" + strconv.Itoa(j) + "-> block" + strconv.Itoa(int(inodes.I_block[j])) + ":n\n"
						var filetemp Fileblock
						aux.Seek(int64(super.S_block_start+(int32(unsafe.Sizeof(filetemp))*inodes.I_block[j])), 0)
						binary.Read(aux, binary.LittleEndian, &filetemp)
						strGrafica += "block" + strconv.Itoa(int(inodes.I_block[j])) + " [label = <<table border='0' cellborder='1' cellspacing='0'>\n"
						strGrafica += "<tr><td colspan = '2' ><b> block " + strconv.Itoa(int(inodes.I_block[j])) + "</b></td></tr>\n"
						strGrafica += "<tr><td colspan = '2'>"
						strGrafica += strings.TrimRight(string(filetemp.B_content[:]), "\x00")
						strGrafica += "</td></tr>\n"
						strGrafica += "</table>>];\n"

					}
				}
			}

		}

		inodes = Inodes()
		aux.Seek(int64(super.S_inode_start+(int32(unsafe.Sizeof(inodes))*int32(i+1))), 0)
		binary.Read(aux, binary.LittleEndian, &inodes)
	}

	strGrafica += "\n\n}\n"

	// create and write to file
	file, err := os.Create(fileDot)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()
	_, err = file.WriteString(strGrafica)
	if err != nil {
		fmt.Println(err)
		return
	}

	// execute dot command
	dotCmd := exec.Command("dot", "-Tpdf", fileDot, "-o", repPath)
	err = dotCmd.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	//remove file
	err = os.Remove(fileDot)
	if err != nil {
		fmt.Println(err)
		return
	}

	path_tree = repPath

}

func (usr User) FileReport(repPath string, fileDot string, binPath string, rute string, partitions Partition) {

}

func (usr User) return_disk() string {
	return path_disk
}

func (usr User) return_sb() string {
	return path_sb
}

func (usr User) return_tree() string {
	return path_tree
}

func GetInodes(superbloque Superblock, paths string) int {

	aux, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer aux.Close()
	BMInode := make([]byte, superbloque.S_inodes_count)
	aux.Seek(int64(superbloque.S_bm_inode_start), 0)
	binary.Read(aux, binary.LittleEndian, &BMInode)
	for i := 0; i < int(superbloque.S_inodes_count); i++ {
		if BMInode[i] == 48 {
			return i
		}
	}
	return -1
}
