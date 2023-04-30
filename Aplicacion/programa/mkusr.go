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

func Mkusr(parametros *[]string, discos *[]Disco, sesion *Usuario, salidas *[6]string) {
	//VERIFICAR QUE EL USUARIO ROOT ESTÉ LOGUEADO
	if sesion.user != "root" {
		(*salidas)[0] += "ERROR: Este comando solo funciona con el usuario root."
		return
	}

	//VARIABLES
	paramFlag := true                             //Indica si se cumplen con los parametros del comando
	required := true                              //Indica si vienen los parametros obligatorios
	usuario := ""                                  //Atributo usr
	pass := ""                                     //Atributo pass
	grupo := ""                                    //Atributo grupo
	diskName := ""                                 //Nombre del disco
	posDisco := -1                                //Posicion del disco dentro del vector
	posParticion := -1                            //Posicion de la particion dentro del vector del disco
	posInicio := 0                                //Posicion donde inicia la particion
	posLectura := 0                               //Para determinar la posicion de lectura en disco
	inodo_buscado := -1                           //Numero de Inodo del archivo users.txt
	var sblock Sbloque                            //Para leer el superbloque
	var linodo Inodo                              //Para el manejo de los inodos
	var lcarpeta Bcarpetas                        //Para el manejo de bloques de carpetas
	var larchivo Barchivos                        //Para el manejo de bloques de archivo
	texto := ""                                    //Para almacenar el contenido del archivo de usuarios
	existe_grupo := false                          //Indica si se encontró el grupo
	existe_usuario := false                        //Indica si se encontró el usuario
	contador_usuarios := 0                         //Numero de usuarios registrados en el archivo
	bloques_iniciales := 0                         //Numero de bloques que usaba al inicio el archivo
	bloques_finales := 0                           //Cantidad de bloques que el archivo usa al final
	comprobar_bloques := false                     //Indica si se va a buscar espacios
	nuevos_bloques := 0                            //Indica la cantidad de bloques que van buscar
	bloque_inicial := 0                            //Numero de bloque que contiene el inicio del archivo
	espacios_vacios := 0                           //Bloques vacios contiguos en el bitmap
	escribir := ""                                 //Variable para almacenar los cortes del archivo

	//COMPROBACIÓN DE PARAMETROS
	for i := 1; i < len(*parametros); i++ {
		temp := (*parametros)[i]
		salida := regexp.MustCompile(`=`).Split(temp, -1)

		tag := salida[0]
		value := salida[1]

		// Pasar a minusculas
		tag = strings.ToLower(tag)

		if tag == "user" {
			usuario = value
		} else if tag == "pass" {
			pass = value
		}else if tag == "grp" {
			grupo = value
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
	if usuario == "" || grupo == "" || pass == "" {
		required = false
	}

	if !required {
		(*salidas)[0] += "ERROR: La instrucción mkusr carece de todos los parametros obligatorios.\n"
		//fmt.Println("ERROR: La instrucción login carece de todos los parametros obligatorios.")
		return
	}

	//VERIFICACION DE PARAMETROS
    if len(usuario) > 10 || len(grupo) > 10 || len(pass) > 10 {
        (*salidas)[0] += "ERROR: El tamaño máximo para cualquiera de los parametros es de diez caracteres.\n" 
        return;
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

	for i := 0; i < len(lineas); i++ {
		atributos := strings.Split(lineas[i], ",")
		if len(atributos) == 3 {
			if atributos[0] != "0" {
				if atributos[2] == grupo {
					existe_grupo = true
				}
			}
		} else {
			if atributos[0] != "0" {
				if atributos[3] == usuario {
					existe_usuario = true
				}
				contador_usuarios++
			}
		}
	}

	if !existe_grupo {
		(*salidas)[0] +=  "ERROR: El grupo al que desea añadir al usuario no existe.\n"
		return
	}
	
	if existe_usuario {
		(*salidas)[0] +=  "ERROR: El usuario ya existe.\n"
		return
	}

	//NUEVA LINEA PARA EL ARCHIVO
	nuevo := strconv.Itoa(contador_usuarios+1) + ",U," + grupo + "," + usuario + "," + pass + "\n"

	//DETERMINAR EL NUMERO DE BLOQUES USADOS INICIALMENTE
	if len(texto) % 63 == 0 {
		bloques_iniciales = len(texto) / 63
	} else {
		bloques_iniciales = (len(texto) / 63) + 1
	}

	//AÑADIR LA NUEVA LINEA Y DETERMINAR DE NUEVO NUMERO DE BLOQUES
	texto += nuevo
	if len(texto) % 63 == 0 {
		bloques_finales = len(texto) / 63
	} else {
		bloques_finales = (len(texto) / 63) + 1
	}

	//DECIDIR SI SE VA A AÑADIR UN NUEVO BLOQUE
	if bloques_finales != bloques_iniciales {
		comprobar_bloques = true
	}

	// ACTUALIZAR EL SUPERBLOQUE
	enteros := ToInt(sblock.S_free_blocks_count[:]) - (bloques_finales - bloques_iniciales)
	sblock.S_free_blocks_count = [40]byte{}
	copy(sblock.S_free_blocks_count[:], strconv.Itoa(enteros))
	archivo.Seek(int64(posInicio), 0)
	binary.Write(archivo, binary.LittleEndian, &sblock)

	// DETERMINAR EL NUMERO DE BLOQUES NECESARIOS - DIRECTOS
	buscados := bloques_finales - bloques_iniciales
	recorrer = ToStringArray(linodo.I_block[:])
	if comprobar_bloques {
		directos := 0
		for i := 0; i < 16; i++ {
			if recorrer[i] == -1 {
				directos += 1
			}
		}
		if buscados <= directos {
			comprobar_bloques = false
			nuevos_bloques += buscados
		} else {
			buscados -= directos
			nuevos_bloques += directos
		}
	}

	//Mandar error si aún faltan espacios
	if comprobar_bloques {
		(*salidas)[0] += "ERROR: Ya no hay bloques disponibles en el inodo para añadir más información.\n"
		return
	}

	//DETERMINAR SI HAY ESPACIO EN CASO DE NECESITAR MAS BLOQUES (LOS BLOQUES SON CONSECUTIVOS)
	bloqueInicial := recorrer[0]
	tamaño := len(texto)
	if nuevos_bloques != 0 {
		posLectura := ToInt(sblock.S_bm_block_start[:]) + ((bloqueInicial+bloques_iniciales) * int(binary.Size(byte(0))))

		// Contar el número de espacios vacíos
		continuar := true
		var c byte
		for continuar {
			archivo.Seek(int64(posLectura), 0)
			binary.Read(archivo, binary.LittleEndian, &c)

			if c == 'p' || c == 'a' || c == 'c' {
				continuar = false
			} else {
				espacios_vacios += 1
				posLectura += int(binary.Size(byte(0)))
			}
		}

		if espacios_vacios < nuevos_bloques {
			(*salidas)[0] += "ERROR: No hay espacio disponible para añadir un nuevo bloque.\n"
			return
		}
	}

	// REINICIAR TODOS LOS ESPACIOS DEL INODO
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
	linodo.I_s = [40]byte{}
	copy(linodo.I_s[:],  strconv.Itoa(tamaño))
	sliceTemp := ToByteArray(recorrer)
	copy(linodo.I_block[:], sliceTemp)

	posLectura = ToInt(sblock.S_inode_start[:]) + (int(binary.Size(Inodo{})) * int(inodo_buscado))
	archivo.Seek(int64(posLectura), 0)
	binary.Write(archivo, binary.LittleEndian, linodo)

	(*salidas)[0] += "MENSAJE: Usuario añadido correctamente.\n"
}	