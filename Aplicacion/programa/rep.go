package programa

import (
	//"fmt"
	"encoding/binary"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func Rep(parametros *[]string, discos *[]Disco, salidas *[6]string) {
	var paramFlag bool = true // Indica si se cumplen con los parametros del comando
	var required bool = true  // Indica si vienen los parametros obligatorios
	var ruta string = ""      // Atributo path
	var nombre string = ""    // Atributo name
	var id string = ""        // Atributo ID
	var rutaS string = ""     // Atributo ruta
	var diskName string       // Nombre del disco sin los numeros del ID
	var posDisco int = -1     // Posicion del disco en la lista
	var posParticion int = -1 // Posicion de la particion dentro del vector del disco

	//COMPROBACIÓN DE PARAMETROS
	for i := 1; i < len(*parametros); i++ {
		temp := (*parametros)[i]
		salida := regexp.MustCompile(`=`).Split(temp, -1)

		tag := salida[0]
		value := salida[1]

		// Pasar a minusculas
		tag = strings.ToLower(tag)

		if tag == "path" {
			ruta = value
		} else if tag == "id" {
			id = value
		} else if tag == "name" {
			nombre = value
		} else if tag == "ruta" {
			rutaS = value
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
	if nombre == "" || ruta == "" || id == "" {
		required = false
	}

	if !required {
		(*salidas)[0] += "ERROR: La instrucción rep carece de todos los parametros obligatorios.\n"
		//fmt.Println("ERROR: La instrucción rep carece de todos los parametros obligatorios.")
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
	diskName = id[posicion:len(id)]

	//CONVERTIR LA LETRA A BYTE
	posDisco = 65 - int(byte(diskName[0]))

	//EXTRAER LA POSICION DE LA PARTICION EN EL DISCO
	posParticion, err := strconv.Atoi(string(id[2]))
	posParticion -= 1
	if posParticion < 0 {
		posParticion = 0
	}
	if err != nil {

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

	// CREAR DIRECTORIOS EN CASO NO EXISTAN
	if err := os.MkdirAll(filepath.Dir(ruta), os.ModePerm); err != nil {
		(*salidas)[0] += "Error creando directorios.\n"
		//fmt.Println("Error creando directorios:", err)
		return
	}

	// BORRAR EL ARCHIVO EN CASO YA EXISTA
	if err := os.Remove(ruta); err != nil && !os.IsNotExist(err) {
		(*salidas)[0] += "Error borrando archivo.\n"
		//fmt.Println("Error borrando archivo:", err)
		return
	}

	//SEPARAR TIPO DE INSTRUCIION Y EJECUTARLA
	nombre = strings.ToLower(nombre)

	switch nombre {
	case "mbr":
		Mbr(discos, posDisco, &ruta, salidas)
	case "disk":
		Disk(discos, posDisco, &ruta, salidas)
	case "tree":
		Tree(discos, posDisco, posParticion, &ruta, salidas)
	case "sb":
		Sb(discos, posDisco, posParticion, &ruta, salidas)
	case "file":
		File(discos, posDisco, posParticion, &ruta, &rutaS, salidas)
	default:
		(*salidas)[0] += "ERROR: Tipo de reporte invalido.\n"
		//fmt.Println("ERROR: Tipo de reporte invalido.")
	}
}

func Mbr(discos *[]Disco, posDisco int, ruta *string, salidas *[6]string) {
	var codigo string          //Contenedor del codigo del dot
	uso := (*discos)[posDisco] //Disco en uso
	var mbr MBR                //Para leer el mbr
	var posExtendida int       //Posicion para leer la extendida
	var ebr EBR                //Para leer los ebr de las particiones logicas
	var comando string         //Instruccion a mandar a la consola para generar el comando

	//VERIFICAR QUE EXISTA EL ARCHIVO
	archivo, err := os.OpenFile(uso.ruta, os.O_RDWR, 0644)
	if err != nil {
		(*salidas)[0] += "ERROR: No se encontro el disco.\n"
		//fmt.Println("ERROR: No se encontro el disco.")
		return
	}
	defer archivo.Close()

	//LEER EL MBR
	archivo.Seek(0, 0)
	binary.Read(archivo, binary.LittleEndian, &mbr)

	//ESCRIBIR EL DOT
	codigo = "digraph mbr {node [shape=plaintext] struct1 [label= <<TABLE BORDER='2' CELLBORDER='0' CELLSPACING='0'>"

	codigo += "<TR>"
	codigo += "<TD BGCOLOR='#cd6155' WIDTH='300'>REPORTE DE MBR</TD>"
	codigo += "<TD WIDTH='300' BGCOLOR='#cd6155'></TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Tamaño</TD>"
	codigo += "<TD>"
	codigo += ToString(mbr.Mbr_tamano[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Fit</TD>"
	codigo += "<TD>"
	codigo += ToString(mbr.Dsk_fit[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>DSK Signature</TD>"
	codigo += "<TD>"
	codigo += ToString(mbr.Mbr_dsk_signature[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Fecha Creacion</TD>"
	codigo += "<TD>"
	codigo += ToString(mbr.Mbr_fecha_creacion[:])
	codigo += "</TD>"
	codigo += "</TR>"

	//Particiones
	for i := 0; i < 4; i++ {
		if len(strings.Trim(string(mbr.Mbr_partition[i].Part_name[:]), "\x00")) != 0 {
			if mbr.Mbr_partition[i].Part_type[0] == 'p' {
				codigo += "<TR>"
				codigo += "<TD BGCOLOR='#e67e22' WIDTH='300'>PARTICION</TD>"
				codigo += "<TD WIDTH='300' BGCOLOR='#e67e22'></TD>"
				codigo += "</TR>"
				codigo += "<TR>"
				codigo += "<TD>Nombre</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_name[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Tamaño</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_s[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Fit</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_fit[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Inicio</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_start[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Status</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_status[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Tipo</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_type[:])
				codigo += "</TD>"
				codigo += "</TR>"
			} else {
				codigo += "<TR>"
				codigo += "<TD BGCOLOR='#e67e22' WIDTH='300'>PARTICION</TD>"
				codigo += "<TD WIDTH='300' BGCOLOR='#e67e22'></TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Nombre</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_name[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Tamaño</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_s[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Fit</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_fit[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Inicio</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_start[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Status</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_status[:])
				codigo += "</TD>"
				codigo += "</TR>"

				codigo += "<TR>"
				codigo += "<TD>Tipo</TD>"
				codigo += "<TD>"
				codigo += ToString(mbr.Mbr_partition[i].Part_type[:])
				codigo += "</TD>"
				codigo += "</TR>"

				//Recorrer Logicas
				posExtendida = ToInt(mbr.Mbr_partition[i].Part_start[:])
				archivo.Seek(int64(posExtendida), 0)
				binary.Read(archivo, binary.LittleEndian, &ebr)
				continuar := true
				for continuar {
					codigo += "<TR>"
					codigo += "<TD BGCOLOR='#f1c40f' WIDTH='300'>PARTICION LOGICA</TD>"
					codigo += "<TD WIDTH='300' BGCOLOR='#f1c40f'></TD>"
					codigo += "</TR>"

					codigo += "<TR>"
					codigo += "<TD>Nombre</TD>"
					codigo += "<TD>"
					codigo += ToString(ebr.Part_name[:])
					codigo += "</TD>"
					codigo += "</TR>"

					codigo += "<TR>"
					codigo += "<TD>Tamaño</TD>"
					codigo += "<TD>"
					codigo += ToString(ebr.Part_s[:])
					codigo += "</TD>"
					codigo += "</TR>"

					codigo += "<TR>"
					codigo += "<TD>Fit</TD>"
					codigo += "<TD>"
					codigo += ToString(ebr.Part_fit[:])
					codigo += "</TD>"
					codigo += "</TR>"

					codigo += "<TR>"
					codigo += "<TD>Inicio</TD>"
					codigo += "<TD>"
					codigo += ToString(ebr.Part_start[:])
					codigo += "</TD>"
					codigo += "</TR>"

					codigo += "<TR>"
					codigo += "<TD>Status</TD>"
					codigo += "<TD>"
					codigo += ToString(ebr.Part_status[:])
					codigo += "</TD>"
					codigo += "</TR>"

					codigo += "<TR>"
					codigo += "<TD>Next</TD>"
					codigo += "<TD>"
					codigo += ToString(ebr.Part_next[:])
					codigo += "</TD>"
					codigo += "</TR>"

					if ToInt(ebr.Part_next[:]) == -1 {
						continuar = false
					} else {
						posExtendida = ToInt(ebr.Part_next[:])
						archivo.Seek(int64(posExtendida), 0)
						binary.Read(archivo, binary.LittleEndian, &ebr)
					}
				}

			}
		}
	}

	codigo += "</TABLE>>];}"

	//GENERAR EL DOT
	salida, err := os.Create("grafo.dot")
	defer salida.Close()
	_, err = salida.WriteString(codigo + "\n")

	//EXTRAER EL TIPO DE FORMATO
	pos := strings.LastIndex(*ruta, ".")
	extension := (*ruta)[pos+1:]

	//CREAR EL COMANDO DOT
	comando = "dot -T" + extension + " grafo.dot -o '" + *ruta + "'"

	//GENERAR EL GRAFO
	cmd := exec.Command("sh", "-c", comando)
	err = cmd.Run()
	if err != nil {

	}
	(*salidas)[0] += "MENSAJE: Reporte MBR creado correctamente.\n"
	//fmt.Println("MENSAJE: Reporte MBR creado correctamente.")
}

func Disk(discos *[]Disco, posDisco int, ruta *string, salidas *[6]string) {
	//VARIABLES
	var codigo string = ""     //Contenedor del código del dot
	uso := (*discos)[posDisco] //Disco en uso
	var mbr MBR                //Para leer el mbr
	var ebr EBR                //Para leer los ebr de las particiones lógicas
	var comando string         //Instrucción a mandar a la consola para generar el comando
	var size float64           //Tamaño del disco
	finExtendida := -1
	posEBR := -1
	var posiciones OrdenarPosicion
	var porcentaje int //Maneja los porcentajes a escribir en el reporte

	//VERIFICAR QUE EXISTA EL ARCHIVO
	archivo, err := os.OpenFile(uso.ruta, os.O_RDWR, 0644)
	if err != nil {
		(*salidas)[0] += "ERROR: No se encontro el disco.\n"
		//fmt.Println("ERROR: No se encontro el disco.")
		return
	}
	defer archivo.Close()

	//LEER EL MBR
	archivo.Seek(0, 0)
	binary.Read(archivo, binary.LittleEndian, &mbr)

	// BUSCAR LA EXTENDIDA
	for i := 0; i < 4; i++ {
		if mbr.Mbr_partition[i].Part_type[0] == byte('e') {
			posEBR = ToInt(mbr.Mbr_partition[i].Part_start[:])
			finExtendida = ToInt(mbr.Mbr_partition[i].Part_start[:]) + ToInt(mbr.Mbr_partition[i].Part_s[:])
			break
		}
	}

	// ESCRIBIR EL DOT PARA PARTICIONES PRIMARIAS / EXTENDIDA
	codigo += "digraph mbr {node [shape=plaintext] struct1 [label= <<TABLE BORDER='2' CELLBORDER='1' CELLSPACING='0'>"
	size = float64(ToInt(mbr.Mbr_tamano[:]))

	codigo += "<TR>"
	codigo += "<TD ROWSPAN='3' BGCOLOR='#A10035' HEIGHT='100'>MBR</TD>"

	// Crear una lista de las particiones
	for i := 0; i < 4; i++ {
		if len(strings.Trim(string(mbr.Mbr_partition[i].Part_name[:]), "\x00")) != 0 {
			var temp Position
			temp.inicio = ToInt(mbr.Mbr_partition[i].Part_start[:])
			temp.fin = ToInt(mbr.Mbr_partition[i].Part_start[:]) + ToInt(mbr.Mbr_partition[i].Part_s[:]) - 1
			temp.nombre = ToString(mbr.Mbr_partition[i].Part_name[:])
			temp.tipo = mbr.Mbr_partition[i].Part_type[0]
			temp.tamaño = ToInt(mbr.Mbr_partition[i].Part_s[:])
			posiciones = append(posiciones, temp)
		}
	}

	if len(posiciones) != 0 {
		sort.Sort(OrdenarPosicion(posiciones))
	}

	//Añadir el codigo de las particiones y los espacios vacios
	if len(posiciones) == 0 {
		codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
		porcentaje := int((size / size) * 100)
		codigo += strconv.Itoa(porcentaje)
		codigo += "% del disco"
		codigo += "</TD>"
	} else {
		for i := 0; i < len(posiciones); i++ {
			x := &posiciones[i]
			free := 0

			if i == 0 && i != (len(posiciones)-1) {

				if x.tipo == 'p' {
					codigo += "<TD ROWSPAN='3' BGCOLOR='#355764' WIDTH='100'>PRIMARIA<BR/>"
					codigo += x.nombre
					codigo += "<br/>"
					porcentaje = int(math.Round(float64(x.tamaño) / size * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"
				} else {
					codigo += "<TD COLSPAN ='50' BGCOLOR='#FFA500' WIDTH='100'>EXTENDIDA<BR/>"
					codigo += "</TD>"
				}

				y := &posiciones[i+1]
				free = y.inicio - (x.fin + 1)

				if free > 0 {
					codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
					porcentaje = int(math.Round(float64(free) / size * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"
				}
			} else if i == 0 && i == (len(posiciones)-1) {
				// Espacio entre el inicio y la primera particion
				free := x.inicio - int(binary.Size(MBR{})) + 1
				if free > 0 {
					codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
					porcentaje := int(math.Round((float64(free) / float64(size)) * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"
				}

				// Añadir la particion
				if x.tipo == 'p' {
					codigo += "<TD ROWSPAN='3' BGCOLOR='#355764' WIDTH='100'>PRIMARIA<BR/>"
					codigo += x.nombre
					codigo += "<br/>"
					porcentaje := int(math.Round((float64(x.tamaño) / float64(size)) * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"

				} else {
					codigo += "<TD COLSPAN ='50' BGCOLOR='#FFA500' WIDTH='100'>EXTENDIDA<BR/>"
					codigo += "</TD>"
				}

				// Espacio entre la primera particion y el fin
				free = int(size) - (x.fin + 1)
				if free > 0 {
					codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
					porcentaje := int(math.Round((float64(free) / float64(size)) * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"
				}
			} else if i != len(posiciones)-1 {
				// Añadir la partición
				if x.tipo == 'p' {
					codigo += "<TD ROWSPAN='3' BGCOLOR='#355764' WIDTH='100'>PRIMARIA<BR/>" + x.nombre + "<br/>"
					porcentaje := int(math.Round((float64(x.tamaño) / size) * 100))
					codigo += strconv.Itoa(porcentaje) + "% del disco</TD>"
				} else {
					codigo += "<TD COLSPAN ='50' BGCOLOR='#FFA500' WIDTH='100'>EXTENDIDA<BR/></TD>"
				}

				// Espacio entre la partición actual y la siguiente
				y := posiciones[i+1]
				free := y.inicio - (x.fin + 1)
				if free > 0 {
					codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
					porcentaje := int(math.Round((float64(free) / size) * 100))
					codigo += strconv.Itoa(porcentaje) + "% del disco</TD>"
				}
			} else {
				// Añadir la particion
				if x.tipo == 'p' {
					codigo += "<TD ROWSPAN='3' BGCOLOR='#355764' WIDTH='100'>PRIMARIA<BR/>"
					codigo += x.nombre
					codigo += "<br/>"
					porcentaje := int(math.Round((float64(x.tamaño) / size) * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"

				} else {
					codigo += "<TD COLSPAN ='50' BGCOLOR='#FFA500' WIDTH='100'>EXTENDIDA<BR/>"
					codigo += "</TD>"
				}

				// Espacio entre la primera particion y el fin
				free := int(size) - (x.fin + 1)
				if free > 0 {
					codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
					porcentaje := int(math.Round((float64(free) / size) * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"
				}
			}
		}
	}
	codigo += "</TR>"

	//AÑADIR EL CODIGO DE LAS PARTICIONES LOGICAS
	if posEBR != -1 {
		codigo += "<TR>"
		archivo.Seek(int64(posEBR), 0)
		binary.Read(archivo, binary.LittleEndian, &ebr)
		cabecera_visitada := false //Indica si es la cabecera la revisada
		continuar := true          //Sirve para salir del while
		var free int

		for continuar {
			// Revisar primero la cabecera
			if !cabecera_visitada {
				if ToInt(ebr.Part_s[:]) == 0 {
					codigo += "<TD  BGCOLOR='#1F4690' HEIGHT='100'>EBR</TD>"
					if ToInt(ebr.Part_next[:]) == -1 {
						free = finExtendida - ToInt(ebr.Part_start[:])
					} else {
						free = ToInt(ebr.Part_next[:]) - ToInt(ebr.Part_start[:])
					}

					if free > 0 {
						codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
						porcentaje = int(math.Floor((float64(free) / size) * 100))
						codigo += strconv.Itoa(porcentaje)
						codigo += "% del disco"
						codigo += "</TD>"
					}
				} else {
					codigo += "<TD  BGCOLOR='#1F4690' HEIGHT='100'>EBR</TD>"
					codigo += "<TD  BGCOLOR='#3A5BA0' WIDTH='100'>LOGICA<BR/>"
					codigo += ToString(ebr.Part_name[:])
					codigo += "<br/>"
					porcentaje = int(math.Floor((float64(ToInt(ebr.Part_s[:])) / size) * 100))
					codigo += strconv.Itoa(porcentaje)
					codigo += "% del disco"
					codigo += "</TD>"
					if ToInt(ebr.Part_next[:]) == -1 {
						free = finExtendida - (ToInt(ebr.Part_start[:]) + ToInt(ebr.Part_s[:]))
					} else {
						free = ToInt(ebr.Part_next[:]) - (ToInt(ebr.Part_start[:]) + ToInt(ebr.Part_s[:]))
					}

					if free > 0 {
						codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
						porcentaje = int(math.Floor((float64(free) / size) * 100))
						codigo += strconv.Itoa(porcentaje)
						codigo += "% del disco"
						codigo += "</TD>"
					}
				}

				cabecera_visitada = true
				if ToInt(ebr.Part_next[:]) == -1 {
					continuar = false
				} else {
					posEBR = ToInt(ebr.Part_next[:])
					archivo.Seek(int64(posEBR), 0)
					binary.Read(archivo, binary.LittleEndian, &ebr)
				}
			} else {
				codigo += "<TD  BGCOLOR='#1F4690' HEIGHT='100'>EBR</TD>"
				codigo += "<TD  BGCOLOR='#3A5BA0' WIDTH='100'>LOGICA<BR/>"
				codigo += ToString(ebr.Part_name[:])
				codigo += "<br/>"
				porcentaje = int(math.Floor((float64(ToInt(ebr.Part_s[:])) / float64(size)) * 100))
				codigo += strconv.Itoa(porcentaje) + "% del disco"
				codigo += "</TD>"
				if ToInt(ebr.Part_next[:]) == -1 {
					free = finExtendida - (ToInt(ebr.Part_start[:]) + ToInt(ebr.Part_s[:]))
				} else {
					free = ToInt(ebr.Part_next[:]) - (ToInt(ebr.Part_start[:]) + ToInt(ebr.Part_s[:]))
				}
				if free > 0 {
					codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
					porcentaje = int(math.Floor((float64(free) / float64(size)) * 100))
					codigo += strconv.Itoa(porcentaje) + "% del disco"
					codigo += "</TD>"
				}
				if ToInt(ebr.Part_next[:]) == -1 {
					continuar = false
				} else {
					posEBR = ToInt(ebr.Part_next[:])
					archivo.Seek(int64(posEBR), 0)
					binary.Read(archivo, binary.LittleEndian, &ebr)
				}
			}
		}
		codigo += "</TR>"
	}

	codigo += "</TABLE>>];}"

	//ALMACENAR EL DOT
	(*salidas)[5] = codigo

	//GENERAR EL DOT
	salida, err := os.Create("grafo.dot")
	defer salida.Close()
	_, err = salida.WriteString(codigo + "\n")

	//EXTRAER EL TIPO DE FORMATO
	pos := strings.LastIndex(*ruta, ".")
	extension := (*ruta)[pos+1:]

	//CREAR EL COMANDO DOT
	comando = "dot -T" + extension + " grafo.dot -o '" + *ruta + "'"

	//GENERAR EL GRAFO
	cmd := exec.Command("sh", "-c", comando)
	err = cmd.Run()
	if err != nil {

	}
	(*salidas)[0] += "MENSAJE: Reporte DISKS creado correctamente.\n"
	//fmt.Println("MENSAJE: Reporte DISKS creado correctamente.")
}

func Tree(discos *[]Disco, posDisco int, posParticion int, ruta *string, salidas *[6]string) {
	//VARIABLES
	codigo := ""
	disc_uso := (*discos)[posDisco]                //Disco en uso
	part_uso := disc_uso.particiones[posParticion] //Particion Montada
	posInicio := -1                                //Posicion donde inicia la particion
	sblock := Sbloque{}                            //Para leer el superbloque
	comando := ""
	posInodos := -1
	posBloques := -1

	//VERIFICAR QUE EL ARCHIVO EXISTE
	archivo, err := os.OpenFile(disc_uso.ruta, os.O_RDWR, 0644) //Para leer el archivo
	if err != nil {
		(*salidas)[0] += "ERROR: No se encontro el disco.\n"
		//fmt.Println("ERROR: No se encontro el disco.")
		return
	}

	//DETERMINAR LA POSICION DE INICIO PARA LEER LA PARTICION
	if part_uso.posMBR != -1 {
		var mbr MBR
		archivo.Seek(0, 0)
		binary.Read(archivo, binary.LittleEndian, &mbr)
		posInicio = ToInt(mbr.Mbr_partition[part_uso.posMBR].Part_start[:])
	} else {
		var ebr EBR
		archivo.Seek(int64(part_uso.posEBR), 0)
		binary.Read(archivo, binary.LittleEndian, &ebr)
		posInicio = ToInt(ebr.Part_start[:])
	}

	//LEER EL SUPERBLOQUE
	archivo.Seek(int64(posInicio), 0)
	binary.Read(archivo, binary.LittleEndian, &sblock)

	//Definir posiciones
	posInodos = ToInt(sblock.S_inode_start[:])
	posBloques = ToInt(sblock.S_block_start[:])
	archivo.Close()

	// Escribir el DOT
	codigo = "digraph G { \n rankdir = LR; node[shape = plaintext];\n"

	// Leer el inodo raíz. Es el número 0.
	inodo_leido := 0
	padre := ""
	leer_inodo(disc_uso.ruta, posInodos, posBloques, inodo_leido, &codigo, padre, salidas)
	codigo += "}"

	//ASIGNAR EL DOT
	(*salidas)[3] = codigo

	//GENERAR EL DOT
	salida, err := os.Create("grafo.dot")
	defer salida.Close()
	_, err = salida.WriteString(codigo + "\n")

	//EXTRAER EL TIPO DE FORMATO
	pos := strings.LastIndex(*ruta, ".")
	extension := (*ruta)[pos+1:]

	//CREAR EL COMANDO DOT
	comando = "dot -T" + extension + " grafo.dot -o '" + *ruta + "'"

	//GENERAR EL GRAFO
	cmd := exec.Command("sh", "-c", comando)
	err = cmd.Run()
	if err != nil {

	}
	(*salidas)[0] += "MENSAJE: Reporte Tree creado correctamente.\n"
	//fmt.Println("MENSAJE: Reporte Tree creado correctamente.")
}

func leer_inodo(ruta string, posInodos int, posBloques int, no_inodo int, codigo *string, padre string, salidas *[6]string) {
	// VARIABLES
	var linodo Inodo   // Para leer inodos
	var posLectura int // Usado para las posiciones de lectura

	// ABRIR ARCHIVO
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0644)
	if err != nil {
		(*salidas)[0] += "Error al abrir archivo.\n"
		//fmt.Println("Error al abrir archivo: %s", err)
	}

	// DECLARAR EL INODO
	posLectura = posInodos + (int(binary.Size(linodo)) * no_inodo)
	archivo.Seek(int64(posLectura), 0)
	binary.Read(archivo, binary.LittleEndian, &linodo)

	if linodo.I_type[0] != '0' {
		if linodo.I_type[0] != '1' {
			(*salidas)[0] += "ERROR: No se encontró el inodo raiz.\n"
			//fmt.Println("ERROR: No se encontró el inodo raiz.")
			return
		}
	}

	nombre := "INODO"
	nombre += strconv.Itoa(no_inodo)
	*codigo += nombre
	nombre = "Inodo "
	nombre += strconv.Itoa(no_inodo)
	*codigo += "[ label = <<TABLE BORDER='2' CELLBORDER='0' CELLSPACING='5' BGCOLOR='#0f4c5c'>\n"
	*codigo += "<TR><TD colspan ='2' ><b>"
	*codigo += nombre
	*codigo += "</b></TD></TR>\n"

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "ID del Propietario:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_uid[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "ID del Grupo:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_gid[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "Tamaño del archivo:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_s[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "Ultima lectura:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_atime[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "Fecha de Creación:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_ctime[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "Ultima modificación:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_mtime[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	recorrer := ToStringArray(linodo.I_block[:])
	for j := 0; j < 16; j++ {
		*codigo += "<TR>"
		*codigo += "<TD Align='left'>"
		nombre := "Bloque " + strconv.Itoa(j) + ":"
		*codigo += nombre
		*codigo += "</TD>"
		*codigo += "<TD PORT='P" + strconv.Itoa(j) + "'>"
		*codigo += strconv.Itoa(recorrer[j])
		*codigo += "</TD>"
		*codigo += "</TR>"
	}

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "Tipo de Inodo:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_type[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	*codigo += "<TR>"
	*codigo += "<TD Align='left'>"
	*codigo += "Permisos:"
	*codigo += "</TD>"
	*codigo += "<TD>"
	*codigo += ToString(linodo.I_perm[:])
	*codigo += "</TD>"
	*codigo += "</TR>"

	*codigo += "</TABLE>>];\n"

	// CONECTAR CON EL PADRE
	nombre = "INODO" + strconv.Itoa(no_inodo)
	if padre != "" {
		*codigo += padre + "->" + nombre + "[minlen = 2];\n"
	}

	// RECORRER LA LISTA DE BLOQUES DEL INODO
	for i, direccion := range recorrer {
		if direccion == -1 {
			continue
		}

		nombre_nodo := nombre + ":P" + strconv.Itoa(i)
		if linodo.I_type[0] == '0' {
			leer_carpeta(ruta, posInodos, posBloques, direccion, codigo, nombre_nodo, salidas)
		} else {
			leer_archivo(ruta, posInodos, posBloques, direccion, codigo, nombre_nodo, salidas)
		}

	}
	archivo.Close()
}

func leer_carpeta(ruta string, posInodos int, posBloques int, no_bloque int, codigo *string, padre string, salidas *[6]string) {
	//VARIABLES
	var lcarpeta Bcarpetas //Para leer bloques de carpetas
	var posLectura int     //Usado para las posiciones de lectura

	//ABRIR ARCHIVO
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0644)
	if err != nil {

	}

	//DECLARAR BLOQUE DE CARPETAS
	posLectura = posBloques + (64 * no_bloque)
	archivo.Seek(int64(posLectura), 0)
	binary.Read(archivo, binary.LittleEndian, &lcarpeta)

	nombre := "BLOQUE"
	nombre += strconv.Itoa(no_bloque)
	*codigo += nombre

	nombre = "Bloque Carpetas "
	nombre += strconv.Itoa(no_bloque)
	*codigo += "[ label = <<TABLE BORDER='2' CELLBORDER='0' CELLSPACING='5' BGCOLOR='#8b8c89'>\n"
	*codigo += "<TR><TD colspan ='2' ><b>"
	*codigo += nombre
	*codigo += "</b></TD></TR>\n"
	*codigo += "<TR><TD><b>Nombre</b></TD><TD><b>Inodo</b></TD></TR>"

	for j := 0; j < 4; j++ {
		temp := lcarpeta.B_content[j]
		*codigo += "<TR>"
		*codigo += "<TD>"
		*codigo += ToString(temp.B_name[:])
		*codigo += "</TD>"
		*codigo += "<TD PORT='P"
		*codigo += strconv.Itoa(j)
		*codigo += "'>"
		*codigo += ToString(temp.B_inodo[:])
		*codigo += "</TD>"
		*codigo += "</TR>"
	}

	*codigo += "</TABLE>>];\n"

	//CONECTAR CON EL PADRE
	nombre = "BLOQUE"
	nombre += strconv.Itoa(no_bloque)

	*codigo += padre
	*codigo += "->"
	*codigo += nombre
	*codigo += "[minlen = 2];\n"

	// RECORRER LA LISTA DE CARPETAS DEL BLOQUE
	for i := 0; i < 4; i++ {
		direccion := ToInt(lcarpeta.B_content[i].B_inodo[:])
		carpeta := ToString(lcarpeta.B_content[i].B_name[:])

		if direccion == -1 || carpeta == "." || carpeta == ".." {
			continue
		}

		nombreNodo := nombre + ":P" + strconv.Itoa(i)
		leer_inodo(ruta, posInodos, posBloques, direccion, codigo, nombreNodo, salidas)
	}
	archivo.Close()
}
func leer_archivo(ruta string, posInodos int, posBloques int, no_bloque int, codigo *string, padre string, salidas *[6]string) {
	//VARIABLES
	var larchivo Barchivos // Para leer bloques de archivos
	var posLectura int     // Usado para las posiciones de lectura

	// ABRIR ARCHIVO
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0666)
	if err != nil {

	}

	// DECLARAR BLOQUE DE ARCHIVOS
	posLectura = posBloques + (64 * no_bloque)
	archivo.Seek(int64(posLectura), 0)
	binary.Read(archivo, binary.LittleEndian, &larchivo)

	nombre := "BLOQUE"
	nombre += strconv.Itoa(no_bloque)
	*codigo += nombre

	nombre = "Bloque Archivos "
	nombre += strconv.Itoa(no_bloque)
	*codigo += "[ label = <<TABLE BORDER='2' CELLBORDER='0' CELLSPACING='5' BGCOLOR='#fb8b24'>\n"
	*codigo += "<TR><TD><b>"
	*codigo += nombre
	*codigo += "</b></TD></TR>\n"

	*codigo += "<TR>"
	*codigo += "<TD>"
	*codigo += ToString(larchivo.B_content[:])
	*codigo += "</TD>"
	*codigo += "</TR>"
	*codigo += "</TABLE>>];\n"

	// CONECTAR CON EL PADRE
	nombre = "BLOQUE"
	nombre += strconv.Itoa(no_bloque)

	*codigo += padre
	*codigo += "->"
	*codigo += nombre
	*codigo += "[minlen = 2];"
	*codigo += "\n"

	archivo.Close()
}

func Sb(discos *[]Disco, posDisco int, posParticion int, ruta *string, salidas *[6]string) {
	//VARIABLES
	var codigo string                                //Contenedor del codigo del dot
	disco_uso := (*discos)[posDisco]                 //Disco en uso
	part_uso := &disco_uso.particiones[posParticion] //Particion Montada
	var archivo *os.File                             //Para leer el archivo
	var mbr MBR                                      //Para leer el mbr
	var ebr EBR                                      //Para leer los ebr de las particiones logicas
	var comando string                               //Instruccion a mandar a la consola para generar el comando
	var posInicio int                                //Posicion donde inicia la particion
	var sblock Sbloque                               //Para leer el superbloque

	//VERIFICAR QUE EXISTA EL ARCHIVO
	archivo, err := os.OpenFile(disco_uso.ruta, os.O_RDWR, 0644)
	if err != nil {
		(*salidas)[0] += "ERROR: No se encontro el disco.\n"
		//fmt.Println("ERROR: No se encontro el disco.")
		return
	}
	defer archivo.Close()

	//DETERMINAR LA POSICION DE INICIO PARA LEER LA PARTICION
	if part_uso.posMBR != -1 {
		archivo.Seek(0, 0)
		binary.Read(archivo, binary.LittleEndian, &mbr)
		posInicio = ToInt(mbr.Mbr_partition[part_uso.posMBR].Part_start[:])
	} else {
		archivo.Seek(int64(part_uso.posEBR), 0)
		binary.Read(archivo, binary.LittleEndian, &ebr)
		posInicio = ToInt(ebr.Part_start[:])
	}

	//LEER EL SUPERBLOQUE
	archivo.Seek(int64(posInicio), 0)
	binary.Read(archivo, binary.LittleEndian, &sblock)

	//ESCRIBIR EL DOT
	codigo = "digraph mbr {node [shape=plaintext] struct1 [label= <<TABLE BORDER='2' CELLBORDER='0' CELLSPACING='0'>"

	codigo += "<TR>"
	codigo += "<TD BGCOLOR='#cd6155' WIDTH='300'>REPORTE DE SUPERBLOQUE</TD>"
	codigo += "<TD WIDTH='300' BGCOLOR='#cd6155'></TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Tipo de Sistema</TD>"
	codigo += "<TD>"
	if sblock.S_filesystem_type[0] == '2' {
		codigo += "EXT2"
	} else {
		codigo += "EXT3"
	}
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Posición del Bitmap de Inodos</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_bm_inode_start[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Tamaño del Inodo</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_inode_s[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Inicio de los Inodos</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_inode_start[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Posición del Primer Inodo Libre</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_firts_ino[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Total de Inodos</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_inodes_count[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Inodos Libres</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_free_inodes_count[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Posición del Bitmap de Bloques</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_bm_block_start[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Tamaño del Bloque</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_block_s[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Inicio de los Bloques</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_block_start[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Posición del Primer Bloque Libre</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_first_blo[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Total de Bloques</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_blocks_count[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Bloques Libres</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_free_blocks_count[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Ultima fecha - Montado</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_mtime[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Ultima fecha - Desmontado</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_umtime[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Veces Montado</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_mnt_count[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "<TR>"
	codigo += "<TD>Magic Number</TD>"
	codigo += "<TD>"
	codigo += ToString(sblock.S_magic[:])
	codigo += "</TD>"
	codigo += "</TR>"

	codigo += "</TABLE>>];}"

	//ASIGNAR EL DOT
	(*salidas)[4] = codigo

	//GENERAR EL DOT
	salida, err := os.Create("grafo.dot")
	defer salida.Close()
	_, err = salida.WriteString(codigo + "\n")

	//EXTRAER EL TIPO DE FORMATO
	pos := strings.LastIndex(*ruta, ".")
	extension := (*ruta)[pos+1:]

	//CREAR EL COMANDO DOT
	comando = "dot -T" + extension + " grafo.dot -o '" + *ruta + "'"

	//GENERAR EL GRAFO
	cmd := exec.Command("sh", "-c", comando)
	err = cmd.Run()
	if err != nil {

	}
	(*salidas)[0] += "MENSAJE: Reporte Super Bloque creado correctamente.\n"
	//fmt.Println("MENSAJE: Reporte Super Bloque creado correctamente.")
}

func File(discos *[]Disco, posDisco int, posParticion int, ruta *string, ruta_contenido *string, salidas *[6]string) {
	//VARIABLES
	disc_uso := (*discos)[posDisco]                //Disco en uso
	part_uso := disc_uso.particiones[posParticion] //Particion Montada
	posInicio := 0                                 //Posicion donde inicia la particion
	sblock := Sbloque{}                            //Para leer el superbloque
	linodo := Inodo{}                              //Para leer los inodos
	larchivo := Barchivos{}                        //Para leer bloques de archivos
	lcarpeta := Bcarpetas{}                        //Para leer bloques de carpeta
	posLectura := 0                                //Usado para las posiciones de lectura
	inodo_leido := 0                               //Numero de inodo leido actualmente
	contenido := ""                                //Para almacenar el contenido del archivo

	//VERIFICAR QUE EL ARCHIVO EXISTE
	archivo, err := os.OpenFile(disc_uso.ruta, os.O_RDWR, 0644) //Para leer el archivo
	if err != nil {
		(*salidas)[0] += "ERROR: No se encontro el disco.\n"
		//fmt.Println("ERROR: No se encontro el disco.")
		return
	}
	defer archivo.Close()

	//VERIFICAR QUE VENGA LA RUTA DEL ARCHIVO A LEER
	if *ruta_contenido == "" {
		(*salidas)[0] += "ERROR: Se necesita la ruta del archivo a leer.\n"
		//fmt.Println("ERROR: Se necesita la ruta del archivo a leer.")
		return
	}

	//DETERMINAR LA POSICION DE INICIO PARA LEER LA PARTICION
	if part_uso.posMBR != -1 {
		var mbr MBR
		archivo.Seek(0, 0)
		binary.Read(archivo, binary.LittleEndian, &mbr)
		posInicio = ToInt(mbr.Mbr_partition[part_uso.posMBR].Part_start[:])
	} else {
		var ebr EBR
		archivo.Seek(int64(part_uso.posEBR), 0)
		binary.Read(archivo, binary.LittleEndian, &ebr)
		posInicio = ToInt(ebr.Part_start[:])
	}

	//LEER EL SUPERBLOQUE
	archivo.Seek(int64(posInicio), 0)
	binary.Read(archivo, binary.LittleEndian, &sblock)

	//LEER EL INODO RAIZ
	posLectura = ToInt(sblock.S_inode_start[:])
	inodo_leido = 0
	archivo.Seek(int64(posLectura), 0)
	binary.Read(archivo, binary.LittleEndian, &linodo)

	if linodo.I_type[0] != '0' {
		if linodo.I_type[0] != '1' {
			(*salidas)[0] += "ERROR: No se encontró el inodo raiz.\n"
			//fmt.Println("ERROR: No se encontró el inodo raiz.")
			return
		}
	}

	// SEPARAR LOS NOMBRES QUE VENGAN EN LA RUTA
	path_cont := strings.Split(*ruta_contenido, "/")

	// BUSCAR EL ARCHIVO
	posicion := 1
	continuar := true
	if *ruta_contenido == "/" {
		continuar = false
	} else if path_cont[0] != "" {
		continuar = false
	}

	for continuar {
		inodo_temporal := -1

		recorrer := ToStringArray(linodo.I_block[:])
		//Buscar si existe la carpeta
		for i := 0; i < 16; i++ {
			if inodo_temporal != -1 {
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

				if carpeta == path_cont[posicion] {
					copy(linodo.I_atime[:], []byte(time.Now().String()))
					posLectura = ToInt(sblock.S_inode_start[:]) + int(binary.Size(Inodo{}))*int(inodo_leido)
					archivo.Seek(int64(posLectura), 0)
					binary.Write(archivo, binary.LittleEndian, &linodo)

					inodo_temporal = ToInt(lcarpeta.B_content[j].B_inodo[:])
					inodo_leido = inodo_temporal
					posicion += 1
					posLectura = ToInt(sblock.S_inode_start[:]) + int(binary.Size(Inodo{}))*int(inodo_temporal)
					archivo.Seek(int64(posLectura), 0)
					binary.Read(archivo, binary.LittleEndian, &linodo)
					break
				}
			}
		}

		if inodo_temporal == -1 {
			continuar = false
			(*salidas)[0] += "ERROR: La ruta ingresada del contenido no existe.\n"
			//fmt.Println("ERROR: La ruta ingresada del contenido no existe.")
			inodo_leido = -1
		} else if posicion == len(path_cont) && linodo.I_type[0] == '1' {
			continuar = false
		} else if posicion == len(path_cont) && linodo.I_type[0] == '0' {
			continuar = false
			(*salidas)[0] += "ERROR: No se encontró el archivo con el contenido a leer.\n"
			//fmt.Println("ERROR: No se encontró el archivo con el contenido a leer.")
			inodo_leido = -1
		}
	}

	if inodo_leido == -1 {
		return
	}

	//Leer el contenido del archivo
	recorrer := ToStringArray(linodo.I_block[:])
	for i := 0; i < 16; i++ {
		if recorrer[i] == -1 {
			continue
		}

		posLectura := ToInt(sblock.S_block_start[:]) + (recorrer[i] * int(binary.Size(Barchivos{})))
		archivo.Seek(int64(posLectura), 0)
		binary.Read(archivo, binary.LittleEndian, &larchivo)

		temp := ToString(larchivo.B_content[:])
		contenido += temp
	}

	//ASIGNAR EL CONTENIDO
	(*salidas)[2] = contenido
	//GENERAR EL DOT
	salida, err := os.Create(*ruta)
	defer salida.Close()
	_, err = salida.WriteString(contenido + "\n")

	(*salidas)[0] += "MENSAJE: Reporte file creado correctamente.\n"
	//fmt.Println("MENSAJE: Reporte file creado correctamente.")

}
