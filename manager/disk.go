package manager

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Disk struct{}

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

type Ebr struct {
	EBR_status byte
	EBR_fit    byte
	EBR_start  int32
	EBR_size   int32
	EBR_next   int32
	EBR_name   [16]byte
}

type Mounted struct {
	NameP    string
	Id       string
	Namedisk string
	No       int
}
type Mount struct {
	Disco string
	Path  string
	Cont  int
	ids   []Mounted
}

type Transition struct {
	partition int32
	start     int32
	end       int32
	before    int32
	after     int32
}

type Inode struct {
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

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

type Folderblock struct {
	B_content [4]Content
}

type Fileblock struct {
	B_content [64]byte
}

func (disk Disk) Mkdisk(tks []string) string {
	//inicializar variables
	size := 0
	paths := ""
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
				return "Parametro fit no valido"
			}
		} else if strings.ToLower(tk) == "unit" {
			if strings.ToLower(token) == "k" || strings.ToLower(token) == "m" {
				unit = strings.ToLower(token)
			} else {
				return "Parametro unit no valido"
			}
		} else if strings.ToLower(tk) == "size" {

			sizes, err := strconv.Atoi(token)
			if err != nil || sizes <= 0 {
				return "Parametro size no valido"
			}
			size = sizes

		} else if strings.ToLower(tk) == "path" {

			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				paths = token[1 : len(token)-1]
			} else {
				paths = token
			}

			//obtener ruta de carpetas
			aux_path = paths
			for i := len(aux_path) - 1; i >= 0; i-- {
				if paths[i] == '/' {
					aux_path = aux_path[:i]
					break
				}
			}
		}
	}

	if FileExist(paths) {
		disk.LeerMMR(paths)
		return "El disco ya existe"
	}

	//verificar si existe el directorio
	if _, err := os.Stat(aux_path); os.IsNotExist(err) {
		cmds := "mkdir -p \"" + aux_path + "\""
		exec.Command("sh", "-c", cmds).Run()
	}

	// inicializar disco
	fit = fit[:2]

	if unit == "m" {
		fmt.Println("size")
		size = 1024 * 1024 * size
	} else if unit == "k" {
		size = 1024 * size
	}

	//crear archivo
	archivo, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0666)

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
	MBR.MBR_Part_1 = Partitions()
	MBR.MBR_Part_2 = Partitions()
	MBR.MBR_Part_3 = Partitions()
	MBR.MBR_Part_4 = Partitions()

	if err := binary.Write(archivo, binary.LittleEndian, &MBR); err != nil {
		panic(err)
	}

	return "Disco creado exitosamente: " + paths

}

func (disk Disk) Rmdisk(tks []string) string {
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

		}
	}

	if !FileExist(path) {
		return "El disco no existe"
	}
	os.Remove(path)

	return "Disco eliminado con exito"

}

func FileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, _ := os.Create(path)
		defer file.Close()
	} else {
		return true
	}
	return false
}

var startValue int

func (disk Disk) Fdisk(tks []string) string {
	startValue = 0

	//inicializar variables
	size := 0
	paths := ""
	fit := "wf"
	unit := "k"
	typed := "p"
	name := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "fit" {
			if strings.ToLower(token) == "bf" || strings.ToLower(token) == "ff" || strings.ToLower(token) == "wf" {
				fit = strings.ToLower(token)
			} else {
				return "Parametro fit no valido"
			}
		} else if strings.ToLower(tk) == "unit" {
			if strings.ToLower(token) == "k" || strings.ToLower(token) == "m" {
				unit = strings.ToLower(token)
			} else {
				return "Parametro unit no valido"
			}
		} else if strings.ToLower(tk) == "size" {

			sizes, err := strconv.Atoi(token)
			if err != nil || sizes <= 0 {
				fmt.Println("Parametro size no valido")
				return "Parametro size no valido"
			}
			size = sizes

		} else if strings.ToLower(tk) == "path" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				paths = token[1 : len(token)-1]
			} else {
				paths = token
			}

		} else if strings.ToLower(tk) == "type" {
			if strings.ToLower(token) == "e" || strings.ToLower(token) == "l" || strings.ToLower(token) == "p" {
				typed = strings.ToLower(token)
			} else {
				fmt.Println("Parametro type no valido")
			}
		} else if strings.ToLower(tk) == "name" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				name = token[1 : len(token)-1]
			} else {
				name = token
			}
		}
	}

	var tipo_part byte
	is_type := false
	if typed == "e" {
		tipo_part = 'e'
	} else if typed == "l" {
		is_type = true
		tipo_part = 'l'
	} else if typed == "p" {
		tipo_part = 'p'
	}
	var ajust byte
	if fit == "ff" {
		ajust = 'f'
	} else if fit == "wf" {
		ajust = 'w'
	} else if fit == "bf" {
		ajust = 'b'
	}
	if unit == "m" {
		size = size * 1024 * 1024
	} else if unit == "k" {
		size = size * 1024
	} else if unit == "" {
		size = size * 1024
	}

	var Disco Mbr
	archivo, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0755)
	archivo.Seek(0, 0)
	binary.Read(archivo, binary.LittleEndian, &Disco)
	defer archivo.Close()
	partitions := List_Partition(Disco)
	between := []Transition{}
	used := 0
	ext := 0
	c := int32(0)
	base := int32(unsafe.Sizeof(Disco))
	var extended Partition
	for _, prttn := range partitions {
		if prttn.PART_status == '1' {
			var trn Transition
			trn.partition = c
			trn.start = prttn.PART_start
			trn.end = prttn.PART_start + prttn.PART_size
			trn.before = trn.start - base
			base = trn.end
			if used != 0 {
				between[used-1].after = trn.start - (between[used-1].end)
			}
			between = append(between, trn)
			used++

			if prttn.PART_type == 'e' {
				ext++
				extended = prttn
			}
		}
		if used == 4 && !is_type {
			return "No se puede crear mas particiones"
		} else if ext == 1 && tipo_part == 'e' {
			return "Ya existe una particion extendida"
		}
		c++
	}
	if ext == 0 && tipo_part == 'l' {
		return "No existe una particion extendida"
	}

	if used != 0 {
		between[len(between)-1].after = Disco.MBR_size - (between[len(between)-1].end)
	}

	_, err := disk.findby(Disco, name, paths)
	if err == nil {
		return "Ya existe una particion con ese nombre"
	}
	var transitions Partition
	transitions.PART_status = '1'
	transitions.PART_fit = ajust
	copy(transitions.PART_name[:], name)
	transitions.PART_size = int32(size)
	transitions.PART_type = tipo_part

	if is_type {
		return disk.logic(transitions, extended, paths)
	}

	Disco, _ = disk.adjust(Disco, transitions, between, partitions, used)
	bfile, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0755)
	defer bfile.Close()
	binary.Write(bfile, binary.LittleEndian, &Disco)
	if tipo_part == 'p' {
		return "Particion primaria creada con exito"
	}
	if tipo_part == 'e' {
		ebr := Ebrs()
		ebr.EBR_start = int32(startValue)
		bfile.Seek(int64(startValue), 0)
		binary.Write(bfile, binary.LittleEndian, &ebr)
		return "Particion extendida creada con exito"
	}

	return ""

}

func List_Partition(mbr Mbr) []Partition {
	List := []Partition{}
	List = append(List, mbr.MBR_Part_1)
	List = append(List, mbr.MBR_Part_2)
	List = append(List, mbr.MBR_Part_3)
	List = append(List, mbr.MBR_Part_4)
	return List
}

func (disk Disk) findby(mbr Mbr, name string, path string) (Partition, error) {
	var partitions [4]Partition
	partitions[0] = mbr.MBR_Part_1
	partitions[1] = mbr.MBR_Part_2
	partitions[2] = mbr.MBR_Part_3
	partitions[3] = mbr.MBR_Part_4

	ext := false
	var extended Partition
	var bytes [16]byte
	copy(bytes[:], []byte(name))
	for _, partition1 := range partitions {
		if partition1.PART_status == '1' {
			if partition1.PART_name == bytes {
				return partition1, nil
			} else if partition1.PART_type == 'e' {
				ext = true
				extended = partition1
			}
		}
	}
	if ext {
		var ebrs []Ebr = disk.getlogics(extended, path)
		for _, ebr := range ebrs {
			if ebr.EBR_status == '1' {
				if ebr.EBR_name == bytes {
					var tmp Partition
					tmp.PART_status = '1'
					tmp.PART_type = 'l'
					tmp.PART_fit = ebr.EBR_fit
					tmp.PART_start = ebr.EBR_start
					tmp.PART_size = ebr.EBR_size
					tmp.PART_name = ebr.EBR_name
					return tmp, nil
				}
			}
		}

	}

	return Partition{}, fmt.Errorf("la partición no existe")
}

func (disk *Disk) getlogics(partition Partition, path string) []Ebr {
	var ebrs []Ebr
	archivo, _ := os.OpenFile(path, os.O_RDWR, 0666)
	defer archivo.Close()
	archivo.Seek(0, 0)
	var tmp = Ebrs()
	archivo.Seek(int64(partition.PART_start), 0)
	binary.Read(archivo, binary.LittleEndian, &tmp)

	for {
		if !(tmp.EBR_status == '0' && tmp.EBR_next == -1) {
			if tmp.EBR_status != '0' {
				ebrs = append(ebrs, tmp)
			}
			archivo.Seek(int64(tmp.EBR_next), 0)
			binary.Read(archivo, binary.LittleEndian, &tmp)

		} else {
			break
		}

	}
	return ebrs
}

func (disk Disk) logic(partition Partition, ep Partition, p string) string {
	var nlogic Ebr
	nlogic.EBR_status = '1'
	nlogic.EBR_fit = partition.PART_fit
	nlogic.EBR_size = partition.PART_size
	copy(nlogic.EBR_name[:], partition.PART_name[:])
	nlogic.EBR_next = -1

	archivo, _ := os.OpenFile(p, os.O_RDWR, 0666)
	defer archivo.Close()
	archivo.Seek(0, 0)

	var tmp Ebr
	archivo.Seek(int64(ep.PART_start), 0)
	binary.Read(archivo, binary.LittleEndian, &tmp)
	fmt.Println("temp", tmp.EBR_start)
	size := 0
	for {
		size += int(tmp.EBR_size) + binary.Size(Ebr{})
		if tmp.EBR_status == '0' && tmp.EBR_next == -1 {
			nlogic.EBR_start = tmp.EBR_start
			nlogic.EBR_next = nlogic.EBR_start + nlogic.EBR_size + int32(binary.Size(Ebr{}))
			if (ep.PART_size - int32(size)) <= nlogic.EBR_size {
				return "almacenamiento al maximo nose puede crear la particion logica"
			}
			archivo.Seek(int64(nlogic.EBR_start), 0)
			binary.Write(archivo, binary.LittleEndian, &nlogic)
			archivo.Seek(int64(nlogic.EBR_next), 0)
			var addLogic Ebr
			addLogic.EBR_status = '0'
			addLogic.EBR_next = -1
			addLogic.EBR_start = nlogic.EBR_next
			archivo.Seek(int64(addLogic.EBR_start), 0)
			binary.Write(archivo, binary.LittleEndian, &addLogic)
			return "partición creada correctamente "
		}
		archivo.Seek(int64(tmp.EBR_next), 0)
		binary.Read(archivo, binary.LittleEndian, &tmp)

	}
}

func (disk *Disk) adjust(mbr Mbr, p Partition, t []Transition, ps []Partition, u int) (Mbr, error) {
	if u == 0 {
		p.PART_start = int32(unsafe.Sizeof(mbr))
		startValue = int(p.PART_start)
		mbr.MBR_Part_1 = p
		return mbr, nil
	} else {
		var toUse Transition
		var c int = 0
		for _, tr := range t {
			if c == 0 {
				toUse = tr
				c++
				continue
			}
			if mbr.MBR_fit == 'f' {
				if toUse.before >= p.PART_size || toUse.after >= p.PART_size {
					break
				}
				toUse = tr
			} else if mbr.MBR_fit == 'b' {
				if toUse.before < p.PART_size || toUse.after <= p.PART_size {
					toUse = tr
				} else {
					if tr.before >= p.PART_size || tr.after >= p.PART_size {
						b1 := toUse.before - p.PART_size
						a1 := toUse.after - p.PART_size
						b2 := tr.before - p.PART_size
						a2 := tr.after - p.PART_size

						if (b1 < b2 && b1 < a2) || (a1 < b2 && a1 < a2) {
							c++
							continue
						}
						toUse = tr
					}
				}

			} else if mbr.MBR_fit == 'w' {

				if !(toUse.before >= p.PART_size) || !(toUse.after >= p.PART_size) {
					toUse = tr
				} else {
					if tr.before >= p.PART_size || tr.after >= p.PART_size {
						b1 := toUse.before - p.PART_size
						a1 := toUse.after - p.PART_size
						b2 := tr.before - p.PART_size
						a2 := tr.after - p.PART_size

						if (b1 > b2 && b1 > a2) || (a1 > b2 && a1 > a2) {
							c++
							continue
						}
						toUse = tr
					}
				}
			}
			c++
		}

		if toUse.before >= p.PART_size || toUse.after >= p.PART_size {
			if mbr.MBR_fit == 'f' {
				if toUse.before >= p.PART_size {
					p.PART_start = toUse.start - toUse.before
					startValue = int(p.PART_start)
				} else {
					p.PART_start = toUse.end
					startValue = int(p.PART_start)
				}
			} else if mbr.MBR_fit == 'b' {
				b1 := toUse.before - p.PART_size
				a1 := toUse.after - p.PART_size
				if (toUse.before >= p.PART_size && b1 < a1) || !(toUse.after >= p.PART_start) {
					p.PART_start = toUse.start - toUse.before
					startValue = int(p.PART_start)
				} else {
					p.PART_start = toUse.end
					startValue = int(p.PART_start)
				}
			} else if mbr.MBR_fit == 'w' {
				b1 := toUse.before - p.PART_size
				a1 := toUse.after - p.PART_size
				if (toUse.before >= p.PART_size && b1 > a1) || !(toUse.after >= p.PART_start) {
					p.PART_start = toUse.start - toUse.before
					startValue = int(p.PART_start)
				} else {
					p.PART_start = toUse.end
					startValue = int(p.PART_start)
				}
			}

			var partitions [4]Partition
			for i := 0; i < len(ps); i++ {
				copy(partitions[:], ps[:])
			}

			for i, partition := range partitions {
				if partition.PART_status == '0' {
					partitions[i] = p
					break
				}
			}

			var aux Partition
			for i := 3; i >= 0; i-- {
				for j := 0; j < i; j++ {
					if partitions[j].PART_start > partitions[j+1].PART_start {
						aux = partitions[j+1]
						partitions[j+1] = partitions[j]
						partitions[j] = aux
					}
				}
			}

			for i := 3; i >= 0; i-- {
				for j := 0; j < i; j++ {
					if partitions[j].PART_status == '0' {
						aux = partitions[j]
						partitions[j] = partitions[j+1]
						partitions[j+1] = aux
					}
				}
			}

			mbr.MBR_Part_1 = partitions[0]
			mbr.MBR_Part_2 = partitions[1]
			mbr.MBR_Part_3 = partitions[2]
			mbr.MBR_Part_4 = partitions[3]
			return mbr, nil
		} else {
			return Mbr{}, errors.New("no hay suficiente espacio para realizar la particion")
		}
	}
}

func (disk Disk) LeerMMR(paths string) {

	archivo, _ := os.Open(paths)
	defer archivo.Close()
	var read_MBR Mbr
	binary.Read(archivo, binary.LittleEndian, &read_MBR)
	fmt.Println("---------------------MBR---------------------")
	fmt.Println("mbr_size: ", read_MBR.MBR_size)
	fmt.Println("mbr_fit: ", string(read_MBR.MBR_fit))
	fmt.Println("mbr_time: ", read_MBR.MBR_time)
	fmt.Println("mbr_asigndisk: ", read_MBR.MBR_asigndisk)
	fmt.Println("tamano del mbr", unsafe.Sizeof(read_MBR))

	List_read := List_Partition(read_MBR)
	for i := 0; i < 4; i++ {
		if List_read[i].PART_status == '1' {
			if List_read[i].PART_type == 'p' {
				fmt.Println("--------------PARTICION PRIMARIA --------------")
				fmt.Println("	part_status: ", string(List_read[i].PART_status))
				fmt.Println("	part_type:   ", string(List_read[i].PART_type))
				fmt.Println("	part_fit: ", string(List_read[i].PART_fit))
				fmt.Println("	part_start: ", List_read[i].PART_start)
				fmt.Println("	part_size: ", List_read[i].PART_size)
				fmt.Println("	part_name: ", string(List_read[i].PART_name[:]))
			} else if List_read[i].PART_type == 'e' {
				fmt.Println("--------------PARTICION EXTENDIDA --------------")
				fmt.Println("	part_status: ", string(List_read[i].PART_status))
				fmt.Println("	part_type:   ", string(List_read[i].PART_type))
				fmt.Println("	part_fit: ", string(List_read[i].PART_fit))
				fmt.Println("	part_start: ", List_read[i].PART_start)
				fmt.Println("	part_size: ", List_read[i].PART_size)
				fmt.Println("	part_name: ", string(List_read[i].PART_name[:]))
				list_ext := disk.getlogics(List_read[i], paths)
				for _, ebr := range list_ext {
					fmt.Println("--------------PARTICION LOGICA --------------")
					fmt.Println("	part_status: ", string(ebr.EBR_status))
					fmt.Println("	part_fit: ", string(ebr.EBR_fit))
					fmt.Println("	part_start: ", ebr.EBR_start)
					fmt.Println("	part_size: ", ebr.EBR_size)
					fmt.Println("	part_next: ", ebr.EBR_next)
					fmt.Println("	part_name: ", string(ebr.EBR_name[:]))
				}
			}
		}
	}

}

var List_mount []Mount
var aumento int = 1

func (m *Mount) AddId(id string, namedisk string, no int, namep string) {
	m.ids = append(m.ids, Mounted{Id: id, Namedisk: namedisk, No: no, NameP: namep})
}

func (disk Disk) Mount(tks []string) string {

	paths := ""
	name := ""
	id := "53"

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "path" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				paths = token[1 : len(token)-1]
			} else {
				paths = token
			}
		} else if strings.ToLower(tk) == "name" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				name = token[1 : len(token)-1]
			} else {
				name = token
			}
		}
	}

	IdLIst := []byte{'1', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

	namedisk := strings.Replace(path.Base(paths), ".dsk", "", -1)

	var Disco Mbr
	file, _ := os.Open(paths)
	defer file.Close()
	file.Seek(0, 0)
	binary.Read(file, binary.LittleEndian, &Disco)

	fmt.Println(Disco.MBR_size, Disco.MBR_fit, Disco.MBR_time, Disco.MBR_asigndisk)

	var partitions [4]Partition
	partitions[0] = Disco.MBR_Part_1
	partitions[1] = Disco.MBR_Part_2
	partitions[2] = Disco.MBR_Part_3
	partitions[3] = Disco.MBR_Part_4
	encontrado_P := false

	for _, buscadoPart := range partitions {

		if buscadoPart.PART_type == 'p' {
			var bytes [16]byte
			copy(bytes[:], []byte(name))
			if buscadoPart.PART_name == bytes {
				encontrado_P = true
				break
			}
		} else if buscadoPart.PART_type == 'e' {
			var ebrs []Ebr = disk.getlogics(buscadoPart, paths)
			for _, buscadoLog := range ebrs {
				var bytes [16]byte
				copy(bytes[:], []byte(name))
				if buscadoLog.EBR_name == bytes {
					encontrado_P = true
					break
				}
			}
		}
	}

	if encontrado_P {
		es_mount := 0
		repetido := false
		var is_L byte
		var cont_L int
		for i := 0; i < len(List_mount); i++ {
			if List_mount[i].Disco == namedisk {
				repetido = true
				id += strconv.Itoa(List_mount[i].Cont)
				terminar := false

				for n := 1; n < len(IdLIst); n++ {
					for y := 0; y < len(List_mount[i].ids); y++ {
						es_mount = List_mount[i].ids[y].No
						es_mount++
						if n == es_mount {
							id += string(IdLIst[n])
							is_L = IdLIst[n]
							terminar = true
							List_mount[i].ids[y].No = es_mount
							break
						}
					}
					if terminar {
						break
					}
				}

				List_mount[i].AddId(id, string(is_L), cont_L, name)
				break
			}
		}

		if !repetido {
			var is_L byte
			List_mount = append(List_mount, Mount{})
			List_mount[len(List_mount)-1].Disco = namedisk
			List_mount[len(List_mount)-1].Path = paths
			List_mount[len(List_mount)-1].Cont = aumento

			id += strconv.Itoa(List_mount[len(List_mount)-1].Cont)
			for i := 1; i < len(IdLIst); i++ {
				if i == 1 {
					id += string(IdLIst[i])
					is_L = IdLIst[i]
					break
				}
			}
			aumento++
			List_mount[len(List_mount)-1].AddId(id, string(is_L), 1, name)

		}

	} else {
		return "no se encontro la particion"
	}
	disk.verVector()

	return "Particion montada con exito"

}

func (disk Disk) verVector() {
	for i := 0; i < len(List_mount); i++ {
		fmt.Println("Disco: ", List_mount[i].Disco)
		fmt.Println("Path: ", List_mount[i].Path)
		for j := 0; j < len(List_mount[i].ids); j++ {
			fmt.Println("Id: ", List_mount[i].ids[j].Id)
			fmt.Println("Name: ", List_mount[i].ids[j].NameP)

		}
	}
}

func (disk Disk) Mkfs(tks []string) string {

	types := ""
	id := ""

	//extraer parametros
	for _, token := range tks {
		tk := token[:strings.Index(token, "=")]
		token = token[len(tk)+1:]
		if strings.ToLower(tk) == "type" {
			//si trae comillas extraerlas
			if strings.HasPrefix(token, "\"") {
				types = token[1 : len(token)-1]
			} else {
				types = token
			}
			if strings.ToLower(types) != "full" {
				return "El parametro type no es valido"
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

	fmt.Println("id-->:", id)

	if types == "full" {
		fmt.Println("se realizara un formateo completo")
	}
	paths := ""
	var particion Partition
	particion, err := disk.EncontrarParticion(id, &paths)
	if err != nil {
		return "No hay discos montados"
	}

	// if disk.EstaFormateado(particion, paths) {
	// 	fmt.Println("YA ESTA FORMATEADO")
	// 	for {
	// 		fmt.Println("Desea formatear de nuevo? s/n")
	// 		return ""
	// 		var respuesta string
	// 		fmt.Scanln(&respuesta)
	// 		if respuesta == "s" || respuesta == "S" {
	// 			break
	// 		}
	// 		if respuesta == "n" || respuesta == "N" {
	// 			return ""
	// 		}
	// 	}
	// }

	ext2 := (particion.PART_size - int32(unsafe.Sizeof(Superblock{}))) / (4 + int32(unsafe.Sizeof(Inode{})) + 3*int32(unsafe.Sizeof(Fileblock{})))

	var superbloque Superblock
	superbloque.S_mtime = time.Now().Unix()
	superbloque.S_umtime = time.Now().Unix()
	superbloque.S_mnt_count = 1
	superbloque.S_filesystem_type = 2
	superbloque.S_inodes_count = ext2
	superbloque.S_blocks_count = ext2 * 3
	superbloque.S_free_blocks_count = ext2 * 3
	superbloque.S_free_inodes_count = ext2
	disk.Format_ext2(superbloque, particion, int(ext2), paths)

	return "Fromateo completo exitoso"

}

func (disk Disk) Format_ext2(superbloque Superblock, particion Partition, bloques int, paths string) {
	superbloque.S_bm_inode_start = particion.PART_start + int32(unsafe.Sizeof(Superblock{}))
	superbloque.S_bm_block_start = superbloque.S_bm_inode_start + int32(bloques)
	superbloque.S_inode_start = superbloque.S_bm_block_start + (3 * int32(bloques))
	superbloque.S_block_start = superbloque.S_inode_start + (int32(unsafe.Sizeof(Inode{})) * int32(bloques))
	var tmp byte = '0'
	leer, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0666)
	defer leer.Close()
	leer.Seek(int64(particion.PART_start), 0)
	binary.Write(leer, binary.LittleEndian, &superbloque)

	leer.Seek(int64(superbloque.S_bm_inode_start), 0)
	for i := 0; i < bloques; i++ {
		binary.Write(leer, binary.LittleEndian, &tmp)
	}
	leer.Seek(int64(superbloque.S_bm_block_start), 0)
	for i := 0; i < (3 * bloques); i++ {
		binary.Write(leer, binary.LittleEndian, &tmp)
	}

	var inodo Inode = Inodes()
	leer.Seek(int64(superbloque.S_inode_start), 0)
	for i := 0; i < bloques; i++ {
		binary.Write(leer, binary.LittleEndian, &inodo)
	}
	var bloqueCarpetas Folderblock = FolderBlocks()
	leer.Seek(int64(superbloque.S_block_start), 0)
	for i := 0; i < (3 * bloques); i++ {
		binary.Write(leer, binary.LittleEndian, &bloqueCarpetas)
	}
	readsuper := SuperBlocks()
	supblock, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer supblock.Close()
	supblock.Seek(int64(particion.PART_start), 0)
	binary.Read(supblock, binary.LittleEndian, &readsuper)

	inodo.I_uid = 1
	inodo.I_gid = 1
	inodo.I_size = 0
	inodo.I_atime = superbloque.S_umtime
	inodo.I_ctime = superbloque.S_umtime
	inodo.I_mtime = superbloque.S_umtime
	inodo.I_block[0] = 0
	inodo.I_type = '0'
	inodo.I_perm = 664

	bloke := FolderBlocks()
	copy(bloke.B_content[0].B_name[:], []byte("."))
	bloke.B_content[0].B_inodo = 0
	copy(bloke.B_content[1].B_name[:], []byte(".."))
	bloke.B_content[1].B_inodo = 0
	copy(bloke.B_content[2].B_name[:], []byte("users.txt"))
	bloke.B_content[2].B_inodo = 1
	copy(bloke.B_content[3].B_name[:], []byte("-"))
	bloke.B_content[3].B_inodo = -1

	data := "1,G,root\n1,U,root,root,123\n"
	inodotemp := Inodes()
	inodotemp.I_uid = 1
	inodotemp.I_gid = 1
	inodotemp.I_size = int32(len(data)) + int32(unsafe.Sizeof(Folderblock{}))
	inodotemp.I_atime = superbloque.S_umtime
	inodotemp.I_ctime = superbloque.S_umtime
	inodotemp.I_mtime = superbloque.S_umtime
	inodotemp.I_block[0] = 1
	inodotemp.I_type = '1'
	inodotemp.I_perm = 664

	inodo.I_size = inodotemp.I_size + int32(unsafe.Sizeof(Folderblock{})) + int32(unsafe.Sizeof(Inode{}))

	var fileb Fileblock
	copy(fileb.B_content[:], []byte(data))

	bfiles, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0666)
	defer bfiles.Close()
	caracter := '1'
	bfiles.Seek(int64(superbloque.S_bm_inode_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &caracter)
	binary.Write(bfiles, binary.LittleEndian, &caracter)

	bfiles.Seek(int64(superbloque.S_bm_block_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &caracter)
	binary.Write(bfiles, binary.LittleEndian, &caracter)

	bfiles.Seek(int64(superbloque.S_inode_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &inodo)
	binary.Write(bfiles, binary.LittleEndian, &inodotemp)

	bfiles.Seek(int64(superbloque.S_block_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &bloke)
	binary.Write(bfiles, binary.LittleEndian, &fileb)

}

func (disk Disk) EncontrarParticion(id string, p *string) (Partition, error) {

	nombreParticion := ""
	paths := ""

	for i := 0; i < len(List_mount); i++ {
		for j := 0; j < len(List_mount[i].ids); j++ {
			if List_mount[i].ids[j].Id == id {
				nombreParticion = List_mount[i].ids[j].NameP
				paths = List_mount[i].Path

				break
			}
		}
	}
	*p = paths
	var mbr Mbr
	file, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer file.Close()
	file.Seek(0, 0)
	binary.Read(file, binary.LittleEndian, &mbr)
	return disk.findby(mbr, nombreParticion, paths)
}

func (disk Disk) EstaFormateado(partition Partition, paths string) bool {
	var super Superblock = SuperBlocks()
	file, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer file.Close()
	file.Seek(0, 0)
	file.Seek(int64(partition.PART_start), 0)
	binary.Read(file, binary.LittleEndian, &super)

	return super.S_filesystem_type == int32(2)

}

func FolderBlocks() Folderblock {
	return Folderblock{B_content: [4]Content{Contents(), Contents(), Contents(), Contents()}}
}
func Contents() Content {
	return Content{B_name: [12]byte{}, B_inodo: -1}
}

func SuperBlocks() Superblock {
	return Superblock{
		S_filesystem_type:   0,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_blocks_count: 0,
		S_free_inodes_count: 0,
		S_magic:             0xEF53,
		S_inode_size:        int32(unsafe.Sizeof(Inode{})),
		S_block_size:        int32(unsafe.Sizeof(Folderblock{})),
		S_first_ino:         0,
		S_first_blo:         0,
		S_bm_inode_start:    0,
		S_bm_block_start:    0,
		S_inode_start:       0,
		S_block_start:       0,
	}
}

func Inodes() Inode {
	return Inode{
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

func Partitions() Partition {
	return Partition{
		PART_status: '0',
		PART_type:   '-',
		PART_fit:    '-',
		PART_start:  -1,
		PART_size:   0,
		PART_name:   [16]byte{},
	}
}

func Ebrs() Ebr {
	return Ebr{
		EBR_status: '0',
		EBR_fit:    '-',
		EBR_start:  -1,
		EBR_size:   0,
		EBR_next:   -1,
		EBR_name:   [16]byte{},
	}
}
