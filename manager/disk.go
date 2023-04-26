package manager

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Mbr struct {
	MBR_size      int32
	MBR_time      int64
	MBR_asigndisk int32
	MBR_fit       byte
	MBR_Part_1    Partition
	MBR_Part_2    Partition
	MBR_Part_3    Partition
	MBR_Part_4    Partition
}

type Partition struct {
	PART_status byte
	PART_type   byte
	PART_fit    byte
	PART_start  int32
	PART_size   int32
	PART_name   [16]byte
}

func NewPartition() Partition {
	return Partition{
		PART_status: '0',
		PART_type:   '-',
		PART_fit:    '-',
		PART_start:  -1,
		PART_size:   0,
		PART_name:   [16]byte{},
	}
}

type EBR struct {
	EBR_status byte
	EBR_fit    byte
	EBR_start  int32
	EBR_size   int32
	EBR_next   int32
	EBR_name   [16]byte
}

func NewEBR() EBR {
	return EBR{
		EBR_status: '0',
		EBR_fit:    '-',
		EBR_start:  -1,
		EBR_size:   0,
		EBR_next:   -1,
		EBR_name:   [16]byte{},
	}
}

type Mount_id struct {
	Id       string
	Namedisk string
	No       int
}
type Mount struct {
	Disco string
	Path  string
	Cont  int
	ids   []Mount_id
}

func (m *Mount) AddId(id string, namedisk string, no int) {
	m.ids = append(m.ids, Mount_id{Id: id, Namedisk: namedisk, No: no})
}

type Inodes struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime int64
	I_ctime int64
	I_mtime int64
	I_block [16]int32
	I_type  byte
	I_perm  int32
}

func NewInodes() Inodes {
	return Inodes{
		I_uid:   -1,
		I_gid:   -1,
		I_size:  0,
		I_atime: 0,
		I_ctime: 0,
		I_mtime: 0,
		I_block: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  '-',
		I_perm:  -1,
	}
}

type Superblock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             int64
	S_umtime            int64
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

func NewSuperblock() Superblock {
	return Superblock{
		S_filesystem_type:   0,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_blocks_count: 0,
		S_free_inodes_count: 0,
		S_magic:             0xEF53,
		S_inode_size:        int32(unsafe.Sizeof(Inodes{})),
		S_block_size:        int32(unsafe.Sizeof(Folderblock{})),
		S_first_ino:         0,
		S_first_blo:         0,
		S_bm_inode_start:    0,
		S_bm_block_start:    0,
		S_inode_start:       0,
		S_block_start:       0,
	}
}

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

func NewFolder() Folderblock {
	return Folderblock{B_content: [4]Content{NewContent(), NewContent(), NewContent(), NewContent()}}
}
func NewContent() Content {
	return Content{B_name: [12]byte{}, B_inodo: -1}
}

type Folderblock struct {
	B_content [4]Content
}

type Fileblock struct {
	B_content [64]byte
}

func Mkdisk(tks []string) {
	//inicializar variables
	size := 0
	path := ""
	aux_path := ""
	fit := "ff"
	unit := "m"

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "fit" {
			if strings.ToLower(token) == "bf" || strings.ToLower(token) == "ff" || strings.ToLower(token) == "wf" {
				fit = strings.ToLower(token)
			} else {
				fmt.Println("Parametro fit no valido")
			}
		} else if strings.ToLower(tk) == "unit" {
			if strings.ToLower(token) == "k" || strings.ToLower(token) == "m" {
				unit = strings.ToLower(token)
			} else {
				fmt.Println("Parametro unit no valido")
			}
		} else if strings.ToLower(tk) == "size" {
			sizes, err := strconv.Atoi(token)
			if err != nil || sizes <= 0 {
				fmt.Println("Parametro size no valido")
				return
			}
			size = sizes

		} else if strings.ToLower(tk) == "path" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				path = token[1 : len(token)-1]
			} else {
				path = token
			}

			//obtener ruta de carpetas
			aux_path = path
			for i := len(aux_path) - 1; i >= 0; i-- {
				if path[i] == '/' {
					aux_path = aux_path[:i]
					break
				}
			}
		} else {
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	//verificar si existe el directorio
	if _, err := os.Stat(aux_path); os.IsNotExist(err) {
		cmds := "mkdir -p \"" + aux_path + "\""
		exec.Command("sh", "-c", cmds).Run()
	}

	// inicializar disco
	fit = fit[:2]

	if unit == "m" {
		size = 1024 * 1024 * size
	} else if unit == "k" {
		size = 1024 * size
	}

	//crear archivo
	archivo, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)

	defer archivo.Close()

	if _, err := archivo.Write([]byte{0}); err != nil {
		panic(err)
	}
	if _, err := archivo.Seek(int64(size-1), 0); err != nil {
		panic(err)
	}
	if _, err := archivo.Write([]byte{0}); err != nil {
		panic(err)
	}

	if _, err := archivo.Seek(0, 0); err != nil {
		panic(err)
	}

	MBR := Mbr{}
	MBR.MBR_size = int32(size)
	MBR.MBR_fit = fit[0]
	MBR.MBR_time = time.Now().Unix()
	MBR.MBR_asigndisk = int32(rand.Intn(501))
	MBR.MBR_Part_1 = NewPartition()
	MBR.MBR_Part_2 = NewPartition()
	MBR.MBR_Part_3 = NewPartition()
	MBR.MBR_Part_4 = NewPartition()

	if err := binary.Write(archivo, binary.LittleEndian, &MBR); err != nil {
		panic(err)
	}

	fmt.Println("Disco creado exitosamente: ", path)

}

func Rmdisk(tks []string) {

	//inicializar variables
	path := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "path" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				path = token[1 : len(token)-1]
			} else {
				path = token
			}

		} else {
			fmt.Println("No se esperaba el parametro: ", tk)
			break
		}
	}

	if !ExisteArchivo(path) {
		fmt.Println("El disco no existe")
		return
	}
	os.Remove(path)
	fmt.Println("Disco eliminado con exito")

}

func ExisteArchivo(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, _ := os.Create(path)
		defer file.Close()
	} else {
		return true
	}
	return false
}
