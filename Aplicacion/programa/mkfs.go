package programa

import(
	"regexp"
	"strings"
	//"fmt"
	"unicode"
	"os"
	"encoding/binary"
	"math"
	"strconv"
	"time"
	"bytes"
)

func Mkfs(parametros *[]string, discos *[]Disco, salidas *[6]string){
    // VARIABLES
    var paramFlag bool = true  //Indica si se cumplen con los parametros del comando
    var required bool = true   //Indica si vienen los parametros obligatorios
    var valid bool = true      //Verifica que los valores de los parametros sean correctos
    var vacio byte = '0'     //Usado para escribir el archivo binario
    var tipo string = ""      //Atrubuto -type
    var id string = ""        //Atributo -id
    var fs string = ""        //Atributo -fs
    var diskName string = ""  //Nombre del disco
    var posDisco int = -1     //Posicion del disco dentro del vector
    var posParticion int = -1 //Posicion de la particion dentro del vector del disco
    var posInicio int         //Posicion donde inicia la particion
    var tamaño int            //Tamaño de la particion que se va a formatear
    var nuevo Sbloque         //Super Bloque nuevo que se va a escribir
    var ninodo Inodo          //Para el manejo de los inodos
    var ncarpeta Bcarpetas    //Para el manejo de bloques de carpetas
    var narchivo Barchivos    //Para el manejo de bloques de archivo
    var posLectura int = -1   //Para posiciones en el disco

	//COMPROBACIÓN DE PARAMETROS
	for i := 1; i < len(*parametros); i++ {
        temp := (*parametros)[i]
        salida := regexp.MustCompile(`=`).Split(temp, -1)

        tag := salida[0]
        value := salida[1]

        // Pasar a minusculas
        tag = strings.ToLower(tag)

        if tag == "id" {
			id = value
		} else if tag == "type" {
			tipo = value
		} else {
			(*salidas)[0] += "ERROR: El parametro" + tag + "no es valido.\n"
			//fmt.Printf("ERROR: El parametro %s no es valido.\n", tag)
			paramFlag = false
			break
		}
    }

    if !paramFlag {
        return
    }

	//COMPROBAR PARAMETROS OBLIGATORIOS
	if id == "" {
		required = false
	}
	
	if !required {
		(*salidas)[0] += "ERROR: La instrucción mount carece de todos los parametros obligatorios.\n"
		//fmt.Println("ERROR: La instrucción mount carece de todos los parametros obligatorios.")
		return
	}

	//VALIDACION DE PARAMETROS
	tipo = strings.ToLower(tipo)
	fs = strings.ToLower(fs)

	if fs == "2fs" || fs == "3fs" || fs == "" {
		// Valid file system type
	} else {
		(*salidas)[0] += "ERROR: Tipo de Sistema de Archivos Invalido.\n"
		//fmt.Println("ERROR: Tipo de Sistema de Archivos Invalido.")
		valid = false
	}
	
	if tipo == "full" || tipo == "" {
		// Valid formatting type
	} else {
		(*salidas)[0] += "ERROR: Tipo de Formateo Invalido.\n"
		//fmt.Println("ERROR: Tipo de Formateo Invalido.")
		valid = false
	}
	
	if !valid {
		return
	}
	
	//PREPARACION DE PARAMETROS
	if fs == "" {
		fs = "2fs"
	}

	//REMOVER LOS NUMEROS DEL ID
    posicion := 0
    for i := 0; i < len(id); i++ {
        if unicode.IsDigit(rune(id[i])) {
            posicion++
        } else {
            break
        }
    }
    diskName = id[posicion:]

    //CONVERTIR LA LETRA A BYTE
    posDisco = 65 - int(byte(diskName[0]))

	//EXTRAER LA POSICION DE LA PARTICION EN EL DISCO
	posParticion, err := strconv.Atoi(string(id[2]))	
	posParticion -= 1
    if posParticion < 0{
        posParticion = 0
    }

    //BUSCAR LA PARTICION DENTRO DEL DISCO MONTADO
	if posDisco > len(*discos){
		(*salidas)[0] += "ERROR: El disco no se encuentra montado.\n"
		//fmt.Println("ERROR: El disco no se encuentra montado.")
		return
	}
    tempD := (*discos)[posDisco]

	if posParticion > len(tempD.particiones){
		(*salidas)[0] += "ERROR: La partición no se encuentra montado.\n"
 		//fmt.Println("ERROR: La partición no se encuentra montado.")
		return
	}
    
	//VERIFICAR QUE EXISTA EL ARCHIVO
	formatear := tempD.particiones[posParticion]
    archivo, err := os.OpenFile(tempD.ruta, os.O_RDWR, 0644)
    if err != nil {
		(*salidas)[0] += "ERROR: No se encontro el disco.\n"
        //fmt.Println("ERROR: No se encontro el disco.")
        return
    }
    defer archivo.Close()

	//DETERMINAR LA POSICION DE INICIO PARA LEER LA PARTICION
    if formatear.posMBR != -1 {
		var mbr MBR
        archivo.Seek(0, 0)
        binary.Read(archivo, binary.LittleEndian, &mbr)
        posInicio = ToInt(mbr.Mbr_partition[formatear.posMBR].Part_start[:])
		tamaño = ToInt(mbr.Mbr_partition[formatear.posMBR].Part_s[:])
    } else {
		var ebr EBR
        archivo.Seek(int64(formatear.posEBR), 0)
        binary.Read(archivo, binary.LittleEndian, &ebr)
        posInicio = ToInt(ebr.Part_start[:])
		tamaño = ToInt(ebr.Part_s[:]);
    }

	//EXCRIBIR EL SUPERBLOQUE
	intTemp := 0
	n := int(math.Floor(float64((tamaño-int(binary.Size(Sbloque{}))) / (196 + int(binary.Size(Inodo{}))))))
	nuevo.S_filesystem_type[0] = byte('2')
	copy(nuevo.S_inodes_count[:], strconv.Itoa(n))
	copy(nuevo.S_blocks_count[:], strconv.Itoa(n * 3))
	copy(nuevo.S_free_blocks_count[:], strconv.Itoa((n * 3) - 2))
	copy(nuevo.S_free_inodes_count[:], strconv.Itoa(n - 2))
	copy(nuevo.S_umtime[:], []byte(time.Now().String()))
	copy(nuevo.S_mtime[:], []byte(time.Now().String()))
	copy(nuevo.S_magic[:], strconv.Itoa(61267))
	copy(nuevo.S_mnt_count[:], strconv.Itoa(1))
	copy(nuevo.S_inode_s[:], strconv.Itoa(int(binary.Size(Inodo{}))))
	copy(nuevo.S_block_s[:], strconv.Itoa(int(binary.Size(Barchivos{}))))
	intTemp = posInicio + int(binary.Size(Sbloque{})) + (int(binary.Size(byte(0))) * n) + (int(binary.Size(byte(0))) * n * 3) + (int(binary.Size(Inodo{})) * 2)
	copy(nuevo.S_firts_ino[:], strconv.Itoa(intTemp))
	intTemp = posInicio + int(binary.Size(Sbloque{})) + (int(binary.Size(byte(0))) * n) + (int(binary.Size(byte(0))) * n * 3) + (int(binary.Size(Inodo{})) * 2) + (int((int(binary.Size(Barchivos{})) * 2)))
	copy(nuevo.S_first_blo[:], strconv.Itoa(intTemp))
	intTemp = posInicio + int(binary.Size(Sbloque{}))
	copy(nuevo.S_bm_inode_start[:], strconv.Itoa(intTemp))
	intTemp = posInicio + int(binary.Size(Sbloque{})) + (int(binary.Size(byte(0))) * n)
	copy(nuevo.S_bm_block_start[:], strconv.Itoa(intTemp))
	intTemp = posInicio + int(binary.Size(Sbloque{})) + (int(binary.Size(byte(0))) * n) + (int(binary.Size(byte(0))) * 3)
	copy(nuevo.S_inode_start[:], strconv.Itoa(intTemp))
	intTemp = posInicio + int(binary.Size(Sbloque{})) + (int(binary.Size(byte(0))) * n) + (int(binary.Size(byte(0))) * 3) + (int(binary.Size(Inodo{})) * n)
	copy(nuevo.S_block_start[:], strconv.Itoa(intTemp))
	
	archivo.Seek(int64(posInicio), 0)
	binary.Write(archivo, binary.LittleEndian, &nuevo)

	// LLENAR CON 0s EL BITMAP DE INODOS
	archivo.Seek(int64(ToInt(nuevo.S_bm_inode_start[:])), 0)
	buffer := bytes.Repeat([]byte{vacio}, int(ToInt(nuevo.S_inodes_count[:])))
	archivo.Write(buffer)

	// ESCRIBIR CON 0s EL BITMAP DE BLOQUES
	archivo.Seek(int64(ToInt(nuevo.S_bm_block_start[:])), 0)
	buffer = bytes.Repeat([]byte{vacio}, int(ToInt(nuevo.S_blocks_count[:])))
	archivo.Write(buffer)

	// LLENAR LOS BLOQUES CON ESPACIOS VACIOS
	archivo.Seek(int64(ToInt(nuevo.S_block_start[:])), 0)
	buffer = bytes.Repeat([]byte{vacio}, int(ToInt(nuevo.S_blocks_count[:])) * int(binary.Size(Barchivos{})))
	archivo.Write(buffer)

	// LLENAR LOS INODOS CON ESPACIOS VACIOS
	archivo.Seek(int64(ToInt(nuevo.S_inode_start[:])), 0)
	buffer = bytes.Repeat([]byte{vacio}, int(ToInt(nuevo.S_inodes_count[:]))*int(binary.Size(Inodo{})))
	archivo.Write(buffer)

	 // MARCAR EL PRIMER INODO
	vacio = '1'
	archivo.Seek(int64(ToInt(nuevo.S_bm_inode_start[:])), 0)
	archivo.Write([]byte{vacio})

	// MARCAR EL PRIMER BLOQUE
	var c byte = 'c'
	archivo.Seek(int64(ToInt(nuevo.S_bm_block_start[:])), 0)
	archivo.Write([]byte{c})

	// CREAR Y ESCRIBIR EL INODO
	copy(ninodo.I_uid[:], strconv.Itoa(1))
	copy(ninodo.I_gid[:], strconv.Itoa(1))
	copy(ninodo.I_s[:], strconv.Itoa(0))
	copy(ninodo.I_atime[:], []byte(time.Now().String()))
	copy(ninodo.I_ctime[:], []byte(time.Now().String()))
	copy(ninodo.I_mtime[:], []byte(time.Now().String()))
	var bloques [16]string
	sliceTemp := []byte{}
	for i := 0; i < 16; i++ {
		bloques[i] = "-1"
	}
	bloques[0] = "0"
	for _, i := range bloques {
        sliceTemp = append(sliceTemp, []byte(i)...)
		sliceTemp = append(sliceTemp, '!')
    }
	copy(ninodo.I_block[:], sliceTemp)
	ninodo.I_type[0] = byte('0')
	copy(ninodo.I_perm[:], "777")
	archivo.Seek(int64(ToInt(nuevo.S_inode_start[:])), 0)
	binary.Write(archivo, binary.LittleEndian, ninodo)

	// CREAR E INICIAR EL BLOQUE DE CARPETAS
	for i := 0; i < 4; i++ {
		copy(ncarpeta.B_content[i].B_name[:], "-")
		copy(ncarpeta.B_content[i].B_inodo [:], []byte(strconv.Itoa(-1)))
	}

	// REGISTRAR EL INODO ACTUAL Y EL DEL PADRE
	ncarpeta.B_content[0].B_inodo = [4]byte{}
	copy(ncarpeta.B_content[0].B_name[:], ".")
	copy(ncarpeta.B_content[0].B_inodo[:], []byte(strconv.Itoa(0)))

	ncarpeta.B_content[1].B_inodo = [4]byte{}
	copy(ncarpeta.B_content[1].B_name[:], "..")
	copy(ncarpeta.B_content[1].B_inodo[:], []byte(strconv.Itoa(0)))
	
	// REGISTRAR EL ARCHIVO DE USUARIOS
	ncarpeta.B_content[2].B_inodo = [4]byte{}
	copy(ncarpeta.B_content[2].B_name[:], []byte("users.txt"))
	copy(ncarpeta.B_content[2].B_inodo[:], []byte(strconv.Itoa(1)))
	archivo.Seek(int64(ToInt(nuevo.S_block_start[:])), 0)
	binary.Write(archivo, binary.LittleEndian, &ncarpeta)
	
	// MARCAR UN NUEVO INODO PARA EL ARCHIVO
	posLectura = ToInt(nuevo.S_bm_inode_start[:]) + binary.Size(vacio)
	archivo.Seek(int64(posLectura), 0)
	binary.Write(archivo, binary.LittleEndian, &vacio)

	// MARCAR UN NUEVO BLOQUE PARA EL ARCHIVO
	var a byte = 'a'
	posLectura = ToInt(nuevo.S_bm_block_start[:]) + binary.Size(a)
	archivo.Seek(int64(posLectura), 0)
	binary.Write(archivo, binary.LittleEndian, &a)
	
	// LLENAR EL INODO DEL ARCHIVO
	ninodo = Inodo{}
	contenido := "1,G,root\n1,U,root,root,123\n"
	copy(ninodo.I_uid[:], strconv.Itoa(1))
	copy(ninodo.I_gid[:], strconv.Itoa(1))
	copy(ninodo.I_s[:], strconv.Itoa(len(contenido)))
	copy(ninodo.I_atime[:], []byte(time.Now().String()))
	copy(ninodo.I_ctime[:], []byte(time.Now().String()))
	copy(ninodo.I_mtime[:], []byte(time.Now().String()))
	sliceTemp = []byte{}
	for i := 0; i < 16; i++ {
		bloques[i] = "-1"
	}
	bloques[0] = "1"
	for _, i := range bloques {
        sliceTemp = append(sliceTemp, []byte(i)...)
		sliceTemp = append(sliceTemp, '!')
    }
	copy(ninodo.I_block[:], sliceTemp)
	ninodo.I_type[0] = byte('1')
	copy(ninodo.I_perm[:], "777")
	posLectura = ToInt(nuevo.S_inode_start[:]) + binary.Size(Inodo{})
	archivo.Seek(int64(posLectura), 0)
	binary.Write(archivo, binary.LittleEndian, &ninodo)
	
	// LLENAR Y ESCRIBIR EL BLOQUE DE ARCHIVOS
	copy(narchivo.B_content[:], contenido)
	posLectura = ToInt(nuevo.S_block_start[:]) + binary.Size(Barchivos{	})
	archivo.Seek(int64(posLectura), 0)
	binary.Write(archivo, binary.LittleEndian, &narchivo)
	
	(*salidas)[0] += "MENSAJE: Particion formateada correctamente.\n"
	//fmt.Println("MENSAJE: Particion formateada correctamente.")
}
