package programa

import (
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func Login(parametros *[]string, discos *[]Disco, sesion *Usuario) {
	//VERIFICAR QUE NO EXISTA UNA SESIÓN
	if sesion.user != "" {
		fmt.Println("ERROR: Ya hay una sesión iniciada.")
		return
	}

	//VARIABLES
	var paramFlag bool = true       //Indica si se cumplen con los parametros del comando
	var required bool = true        //Indica si vienen los parametros obligatorios
	var user string                 //Atributo user
	var pass string                 //Atributo pass
	var id string                   //Atributo id
	var diskName string             //Nombre del disco
	var posDisco int                //Posicion del disco dentro del vector
	var posParticion int            //Posicion de la particion dentro del vector del disco
	var posInicio int               //Posicion donde inicia la particion
	var posLectura int              //Para determinar la posicion de lectura en disco
	var inodo_buscado int = -1      //Numero de Inodo del archivo users.txt
	var sblock Sbloque              //Para leer el superbloque
	var linodo Inodo                //Para el manejo de los inodos
	var lcarpeta Bcarpetas          //Para el manejo de bloques de carpetas
	var larchivo Barchivos          //Para el manejo de bloques de archivo
	var texto string                //Para almacenar el contenido del archivo de usuarios
	var existe_usuario bool = false //Indica si se encontró el usuario

	//COMPROBACIÓN DE PARAMETROS
	for i := 1; i < len(*parametros); i++ {
		temp := (*parametros)[i]
		salida := regexp.MustCompile(`=`).Split(temp, -1)

		tag := salida[0]
		value := salida[1]

		// Pasar a minusculas
		tag = strings.ToLower(tag)

		if tag == "user" {
			user = value
		} else if tag == "id" {
			id = value
		} else if tag == "pass" {
			pass = value
		} else {
			fmt.Printf("ERROR: El parametro %s no es valido.\n", tag)
			paramFlag = false
			break
		}
	}

	if !paramFlag {
		return
	}

	//COMPROBAR PARAMETROS OBLIGATORIOS
	if user == "" || pass == "" || id == "" {
		required = false
	}

	if !required {
		fmt.Println("ERROR: La instrucción login carece de todos los parametros obligatorios.")
		return
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
	if posParticion < 0 {
		posParticion = 0
	}

	//BUSCAR LA PARTICION DENTRO DEL DISCO MONTADO
	if posDisco > len(*discos) {
		fmt.Println("ERROR: El disco no se encuentra montado.")
		return
	}
	tempD := (*discos)[posDisco]

	if posParticion > len(tempD.particiones) {
		fmt.Println("ERROR: La partición no se encuentra montado.")
		return
	}

	//VERIFICAR QUE EXISTA EL ARCHIVO
	formatear := tempD.particiones[posParticion]
	archivo, err := os.OpenFile(tempD.ruta, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("ERROR: No se encontro el disco.")
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
			carpeta := string(lcarpeta.B_content[j].B_name[:])

			if carpeta == "users.txt" {
				inodo_buscado = ToInt(lcarpeta.B_content[j].B_inodo[:])
				break
			}
		}
	}

	if inodo_buscado == -1 {
		fmt.Println("ERROR: No se encontró el archivo de usuarios.")
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

		temp := string(larchivo.B_content[:])
		texto += temp
	}

	//SEPARAR EL ARCHIVO POR LINEAS
	lineas := strings.Split(texto, "\n")

	// BUSCAR EL USUARIO EN LAS LINEAS DEL ARCHIVO
	for i := 0; i < len(lineas); i++ {
		atributos := strings.Split(lineas[i], ",")

		if len(atributos) == 5 { // Los usuarios tienen cinco parametros
			if atributos[0] != "0" {
				if atributos[3] == user && atributos[4] == pass {
					sesion.user = user
					sesion.pass = pass
					sesion.disco = id
					sesion.grupo = atributos[2]
					sesion.id_user = atributos[0]
					existe_usuario = true
				}
			} else {
				if atributos[3] == user && atributos[4] == pass {
					fmt.Println("ERROR: El usuario que busca ha sido eliminado.")
					return
				}
			}
		}

		if existe_usuario {
			break
		}
	}

	// BUSCAR EL GRUPO
	for i := 0; i < len(lineas); i++ {
		// Separar por comas los atributos
		atributos := strings.Split(lineas[i], ",")
		if len(atributos) == 3 { // Los grupos tienen tres parametros
			if atributos[0] != "0" {
				if atributos[2] == sesion.grupo {
					sesion.id_grp = atributos[0]
					break
				}
			} else {
				if atributos[2] == sesion.grupo {
					fmt.Println("ERROR: El grupo al que el usuario pertenece fue eliminado.")
					return
				}
			}
		}
	}

	if !existe_usuario {
		fmt.Println("ERROR: Usuario Inexistente. No es posible iniciar sesión.")
	} else {
		fmt.Println("MENSAJE: Inicio de Sesión Exitoso.")
	}

}
