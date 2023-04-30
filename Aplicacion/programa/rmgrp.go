package programa

import(
	"regexp"
	"strings"
	"unicode"
	"strconv"
	"os"
	"encoding/binary"
	"time"
)

func Rmgrp(parametros *[]string, discos *[]Disco, sesion *Usuario, salidas *[6]string) {
	//VERIFICAR QUE EL USUARIO ROOT ESTÉ LOGUEADO
	if sesion.user != "root" {
		(*salidas)[0] += "ERROR: Este comando solo funciona con el usuario root."
		return
	}

	//VARIABLES
	var paramFlag bool = true //Indica si se cumplen con los parametros del comando
	var required bool = true //Indica si vienen los parametros obligatorios
	var nombre string = "" //Atributo name
	var diskName string = "" //Nombre del disco
	var posDisco int = -1 //Posicion del disco dentro del vector
	var posParticion int = -1 //Posicion de la particion dentro del vector del disco
	var posInicio int //Posicion donde inicia la particion
	var posLectura int //Para determinar la posicion de lectura en disco
	var inodo_buscado int = -1 //Numero de Inodo del archivo users.txt
	var sblock Sbloque //Para leer el superbloque
	var linodo Inodo //Para el manejo de los inodos
	var lcarpeta Bcarpetas //Para el manejo de bloques de carpetas
	var larchivo Barchivos //Para el manejo de bloques de archivo
	var texto string = "" //Para almacenar el contenido del archivo de usuarios
	var existe_grupo bool = false //Indica si se encontró el grupo
	var bloque_inicial int //Numero de bloque que contiene el inicio del archivo
	var escribir string //Variable para almacenar los cortes del archivo

	//COMPROBACIÓN DE PARAMETROS
	for i := 1; i < len(*parametros); i++ {
		temp := (*parametros)[i]
		salida := regexp.MustCompile(`=`).Split(temp, -1)

		tag := salida[0]
		value := salida[1]

		// Pasar a minusculas
		tag = strings.ToLower(tag)

		if tag == "name" {
			nombre = value
		}else {
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
	if nombre == "" {
		required = false
	}

	if !required {
		(*salidas)[0] += "ERROR: La instrucción rmgrp carece de todos los parametros obligatorios.\n"
		//fmt.Println("ERROR: La instrucción login carece de todos los parametros obligatorios.")
		return
	}

	//REMOVER LOS NUMEROS DEL ID
	posicion := 0
	for i := 0; i < len(sesion.disco); i++ {
		if unicode.IsDigit(rune(sesion.disco[i])) {
			posicion++
		} else {
			break
		}
	}
	diskName = sesion.disco[posicion:]

	//CONVERTIR LA LETRA A BYTE
	posDisco = 65 - int(byte(diskName[0]))

	//EXTRAER LA POSICION DE LA PARTICION EN EL DISCO
	posParticion, err := strconv.Atoi(string(sesion.disco[2]))
	posParticion -= 1
	if posParticion < 0 {
		posParticion = 0
	}

	//BUSCAR LA PARTICION DENTRO DEL DISCO MONTADO
	if posDisco > len(*discos) {
		(*salidas)[0] += "ERROR: El disco no se encuentra montado.\n"
		//fmt.Println("ERROR: El disco no se encuentra montado.")
		return
	}
	tempD := (*discos)[posDisco]

	if posParticion > len(tempD.particiones) {
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
	} else {
		var ebr EBR
		archivo.Seek(int64(formatear.posEBR), 0)
		binary.Read(archivo, binary.LittleEndian, &ebr)
		posInicio = ToInt(ebr.Part_start[:])
	}

	//LEER EL SUPERBLOQUE
	archivo.Seek(int64(posInicio), 0)
	binary.Read(archivo, binary.LittleEndian, &sblock)

	//LEER EL INODO RAIZ
	posLectura = ToInt(sblock.S_inode_start[:])
	archivo.Seek(int64(posLectura), 0)
	binary.Read(archivo, binary.LittleEndian, &linodo)

	// BUSCAR EL ARCHIVO DE USUARIOS
	recorrer := ToStringArray(linodo.I_block[:])
	for i := 0; i < 16; i++ {
		if inodo_buscado != -1 {
			break
		}

		if recorrer[i] == -1 {
			continue
		}

		posLectura = ToInt(sblock.S_block_start[:]) + (int(binary.Size(Bcarpetas{})) * recorrer[i])
		archivo.Seek(int64(posLectura), 0)
		binary.Read(archivo, binary.LittleEndian, &lcarpeta)

		for j := 0; j < 4; j++ {
			carpeta := ToString(lcarpeta.B_content[j].B_name[:])

			if carpeta == "users.txt" {
				inodo_buscado = ToInt(lcarpeta.B_content[j].B_inodo[:])
				break
			}
		}
	}

	if inodo_buscado == -1 {
		(*salidas)[0] += "ERROR: No se encontró el archivo de usuarios.\n"
		//fmt.Println("ERROR: No se encontró el archivo de usuarios.")
		return
	}

	// LEER EL INODO DEL ARCHIVO
	posLectura = ToInt(sblock.S_inode_start[:]) + (int(binary.Size(Inodo{})) * inodo_buscado)
	archivo.Seek(int64(posLectura), 0)
	binary.Read(archivo, binary.LittleEndian, &linodo)

	// LEER EL ARCHIVO DE USUARIOS
	recorrer = ToStringArray(linodo.I_block[:])
	for i := 0; i < 16; i++ {
		if recorrer[i] == -1 {
			continue
		}

		posLectura = ToInt(sblock.S_block_start[:]) + (int(binary.Size(Barchivos{})) * recorrer[i])
		archivo.Seek(int64(posLectura), 0)
		binary.Read(archivo, binary.LittleEndian, &larchivo)

		temp := ToString(larchivo.B_content[:])
		texto += temp
	}

	//SEPARAR EL ARCHIVO POR LINEAS
	lineas := strings.Split(texto, "\n")

	// VERIFICAR SI EL GRUPO EXISTE
	for i, linea := range lineas {
		atributos := strings.Split(linea, ",")
		if len(atributos) == 3 { //Los grupos tienen tres parametros
			if atributos[0] != "0" {
				if atributos[2] == nombre {
					editar := "0," + atributos[1] + "," + atributos[2]
					existe_grupo = true
					lineas[i] = editar
				}
			}
		}

		if existe_grupo {
			break
		}
	}

	if !existe_grupo {
		(*salidas)[0] += "ERROR: El grupo que desea eliminar no existe. \n"
		return
	}

	//RECONSTRUIR EL ARCHIVO EDITADO
	texto = strings.Join(lineas, "\n") + "\n"

	//REINICIAR TODOS LOS ESPACIOS DEL INODO
	for i := 0; i < 16; i++ {
		recorrer[i] = -1
	}

	// REESCRIBIR EL ARCHIVO DE USUARIOS
	continuar := true
	posicion = 0
	var c byte
	for continuar {
		revisar := true
		var earchivo Barchivos
	
		if len(texto) > 63 {
			escribir = texto[0:63]
			texto = texto[63:len(texto)]
		} else {
			escribir = texto
			continuar = false
		}
	
		for revisar {
			// Escribir en el bitmap el bloque
			c = 'a'
			posLectura := ToInt(sblock.S_bm_block_start[:]) + ((bloque_inicial) * int(binary.Size(c)))
			archivo.Seek(int64(posLectura), 0)
			binary.Write(archivo, binary.LittleEndian, &c)
	
			// Crear y escribir el bloque
			copy(earchivo.B_content[:], []byte(escribir))
			posLectura = ToInt(sblock.S_block_start[:]) + ((bloque_inicial) * int(binary.Size(Barchivos{})))
			archivo.Seek(int64(posLectura), 0)
			binary.Write(archivo, binary.LittleEndian, &earchivo)
	
			// Actualizar el inodo
			recorrer[posicion] = bloque_inicial
			posicion += 1
			bloque_inicial += 1
			revisar = false
		}
	}

	// REESCRIBIR EL INODO CON TODOS LOS CAMBIOS
	linodo.I_mtime = [30]byte{}
	copy(linodo.I_mtime[:], []byte(time.Now().String()))
	sliceTemp := ToByteArray(recorrer)
	copy(linodo.I_block[:], sliceTemp)

	posLectura = ToInt(sblock.S_inode_start[:]) + (int(binary.Size(Inodo{})) * int(inodo_buscado))
	archivo.Seek(int64(posLectura), 0)
	binary.Write(archivo, binary.LittleEndian, linodo)

	(*salidas)[0] += "MENSAJE: Grupo eliminado correctamente.\n"
}