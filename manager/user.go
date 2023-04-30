package manager

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

type User struct{}
type Usr struct {
	User     string
	Password string
	Id       string
	Uid      int
}

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
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	disk = admin
	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()

	if !estado {
		estado = true
	} else {
		return "ERROR [LOGIN] YA HAY UNA SESION INICIADA"
	}
	var paths string
	particion, err := disk.EncontrarParticion(id, &paths)
	if err != nil {
		estado = false
		return "ERROR [LOGIN] PARA LOGEARSE NECESITA UN DISCO MONTADO"
	}

	readfiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readfiles.Close()
	readfiles.Seek(int64(particion.PART_start), 0)
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
			return "~ ERROR [LOGIN] EL USUARIO NO EXISTE"
		} else if correct_password && !correct_user {
			return "~ ERROR [LOGIN] CONTRASEÑA INCORRECTA"
		} else if correct_user && correct_password {
			return "~ ERROR [LOGIN] EL USUARIO Y LA CONTRASEÑA SON INCORRECTOS"
		}
	}

	return "[LOGIN] --- SESION INICIADA CON EXITO"

}

func (usr User) Logout() string {
	if logeado.User == "" {
		return "~ ERROR [LOGOUT] NO HAY UNA SESION INICIADA"
	}
	logeado = Usr{}
	estado = false
	return "[LOGOUT] --- SESION CERRADA CON EXITO"

}

func (usr User) MKGRP(tks []string) string {
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
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(logeado.User == "root" && logeado.Password == "123") {
		return "~ ERROR [MKGRP] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := disk.EncontrarParticion(logeado.Id, &paths)
	if err != nil {
		return "~ ERROR [MKGRP] PARA CREAR UN GRUPO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
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
		fmt.Println("ARCHIVO: ", archivo)
		return "~ ERROR [MKGRP] EL GRUPO YA EXISTE"
	}
	if newecontrado {
		var bytes [64]byte
		copy(bytes[:], []byte(newarchivo))
		fileblock.B_content = bytes
		fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
		readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
		binary.Write(readFiles, binary.LittleEndian, &fileblock)
		return "[MKGRP] --- GRUPO CREADO CON EXITO"
	}

	archivo += strconv.Itoa(cont_grp) + ",G," + name + "\n"
	var bytes [64]byte
	copy(bytes[:], []byte(archivo))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)

	return "[MKGRP] --- GRUPO CREADO CON EXITO"
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

func (usr User) RMGRP(tks []string) string {

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
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(logeado.User == "root" && logeado.Password == "123") {
		return "~ ERROR [RMGRP] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := disk.EncontrarParticion(logeado.Id, &paths)
	if err != nil {
		return "~ ERROR [RMGRP] PARA ELIMINAR UN GRUPO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)

	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")
	fmt.Println("ARCHIVO: ", archivo)
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
					fmt.Println("ya esta eliminado el grupo")
					return "~ ERROR [RMGRP] EL GRUPO YA ESTA ELIMINADO"
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
		fmt.Println("El grupo no existe")
		return "~ ERROR [RMGRP] EL GRUPO NO EXISTE"
	}

	var bytes [64]byte
	copy(bytes[:], []byte(newarchivo))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)
	return "[RMGRP] --- GRUPO ELIMINADO CON EXITO"
}

func (usr User) MKUSR(tks []string) string {

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
		} else if strings.ToLower(tk) == "id" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				grp = token[1 : len(token)-1]
			} else {
				grp = token
			}
		} else {
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(logeado.User == "root" && logeado.Password == "123") {

		return "~ ERROR [MKUSR] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := disk.EncontrarParticion(logeado.Id, &paths)
	if err != nil {
		return "~ ERROR [MKUSR] PARA CREAR UN USUARIO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
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
				return "~ ERROR [MKUSR] EL GRUPO ESTA ELIMINADO"
			}
		} else if list_users[i][2] == 'U' {
			Users := usr.extraer(list_users[i], 44)
			if Users[0] != "0" && Users[3] == user {
				return "~ ERROR [MKUSR] EL USUARIO YA EXISTE"
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

		return "~ ERROR [MKUSR] EL GRUPO NO EXISTE"
	}
	if newecontrado {
		var bytes [64]byte
		copy(bytes[:], []byte(newarchivo))
		fileblock.B_content = bytes
		fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
		readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
		binary.Write(readFiles, binary.LittleEndian, &fileblock)
		return "[MKUSR] --- USUARIO CREADO CON EXITO"
	}

	archivo += strconv.Itoa(cont_user) + ",U," + grp + "," + user + "," + pwd + "\n"
	var bytes [64]byte
	copy(bytes[:], []byte(archivo))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)
	return "[MKUSR] --- USUARIO CREADO CON EXITO"
}

func (usr User) RMUSER(tks []string) string {

	usuario := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "name" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				usuario = token[1 : len(token)-1]
			} else {
				usuario = token
			}
		} else {
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(logeado.User == "root" && logeado.Password == "123") {

		return "~ ERROR [RMUSER] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := disk.EncontrarParticion(logeado.Id, &paths)
	if err != nil {
		return "~ ERROR [RMUSER] PARA ELIMINAR UN USUARIO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
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
					return "~ ERROR [RMUSR] EL USUARIO YA ESTA ELIMINADO"
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
		return "~ ERROR [RMUSR] EL USUARIO NO EXISTE"
	}

	var bytes [64]byte
	copy(bytes[:], []byte(newarchivo))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)

	return "[RMUSR] --- USUARIO ELIMINADO CON EXITO"
}

func (usr User) REP(tks []string) string {

	name := ""
	path := ""
	id := ""
	rute := ""

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
				path = token[1 : len(token)-1]
			} else {
				path = token
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
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	pathdisco := ""
	_, err := disk.EncontrarParticion(id, &pathdisco)
	if err != nil {
		return "~ ERROR [REP] PARA EL REPORTE NECESITA UN DISCO MONTADO"
	}

	fmt.Println("PATH: ", path)
	fmt.Println("NAME: ", name)
	fmt.Println("ID: ", id)
	fmt.Println("RUTA: ", rute)

	// if name == "disk" {
	// 	Disk(paths, pathdisco)
	// } else {
	// 	return "~ ERROR [REP] NO EXISTE ESE TIPO DE REPORTE"
	// }

	return "[REP] --- REPORTE GENERADO CON EXITO"
}