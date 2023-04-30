package programa

import (
	"encoding/binary"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func Mkdir(parametros *[]string, discos *[]Disco, sesion *Usuario, salidas *[6]string) {
	//VERIFICAR QUE EL USUARIO ROOT ESTÉ LOGUEADO
	if sesion.user == "" {
		(*salidas)[0] += "ERROR: Debe de haber una sesión iniciada para usar este comando.\n"
		//fmt.Println("ERROR: Ya hay una sesión iniciada.")
		return
	}

	//VARIABLES
	var paramFlag bool = true       //Indica si se cumplen con los parametros del comando
	var required bool = true        //Indica si vienen los parametros obligatorios
	var ruta string = ""            //Atributo path
	var padre bool = false          //Atributo p
	var diskName string = ""        //Nombre del disco
	var posDisco int = -1           //Posicion del disco dentro del vector
	var posParticion int = -1       //Posicion de la particion dentro del vector del disco
	var posInicio int               //Posicion donde inicia la particion
	var posLectura int              //Para determinar la posicion de lectura en disco
	var sblock Sbloque              //Para leer el superbloque
	var linodo Inodo                //Para el manejo de los inodos
	var lcarpeta Bcarpetas          //Para el manejo de bloques de carpetas
	var continuar bool = true       //Usado como bandera al buscar la ruta
	var inodo_leido int = -1        //Inodo actual con el que se este trabajando
	var error_permisos bool = false //Bandera para indicar que no se pudo crear en una carpeta por su padre
	var error_padre bool = false    //Indica que una carpeta no existe y no se pudo crear
	var contador int = 0            //Para saber si debe de crearse el reporte

	//COMPROBACIÓN DE PARAMETROS
	for i := 1; i < len(*parametros); i++ {
		temp := (*parametros)[i]
		salida := regexp.MustCompile(`=`).Split(temp, -1)

		//Se separa en dos para manejar el atributo -r
		if len(salida) == 1 {
			tag := salida[0]

			// Pasar a minusculas
			tag = strings.ToLower(tag)

			if tag == "r" {
				padre = true
			} else {
				(*salidas)[0] += "ERROR: El parametro" + tag + "no es valido.\n"
				//fmt.Printf("ERROR: El parametro %s no es valido.\n", tag)
				paramFlag = false
				break
			}

		} else {
			tag := salida[0]
			value := salida[1]

			// Pasar a minusculas
			tag = strings.ToLower(tag)

			if tag == "path" {
				ruta = value
			} else if tag == "r" {
				(*salidas)[0] += "ERROR: El parametro 'r' no recibe ningún valor.\n"
				paramFlag = false
			} else {
				(*salidas)[0] += "ERROR: El parametro" + tag + "no es valido.\n"
				//fmt.Printf("ERROR: El parametro %s no es valido.\n", tag)
				paramFlag = false
				break
			}
		}
	}

	if !paramFlag {
		return
	}

	//COMPROBAR PARAMETROS OBLIGATORIOS
	if ruta == "" {
		required = false
	}

	if !required {
		(*salidas)[0] += "ERROR: La instrucción mkdir carece de todos los parametros obligatorios.\n"
		//fmt.Println("ERROR: La instrucción login carece de todos los parametros obligatorios.")
		return
	}

	//PREPARACIÓN DE PARAMETROS - Separar los nombres que vengan en la ruta.
	path := strings.Split(ruta, "/")

	//PREPARACIÓN DE PARAMETROS - Reducir el tamaño de los nombres al límite
	for i, nombre := range path {
		if len(nombre) >= 12 {
			path[i] = nombre[:11]
		}
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
	inodo_leido = 0
	archivo.Seek(int64(posLectura), 0)
	binary.Read(archivo, binary.LittleEndian, &linodo)

	// CREAR LAS CARPETAS QUE HAGAN FALTA EN LA RUTA
	posicion = 1
	if ruta == "/" {
		continuar = false
	} else if path[0] != "" {
		(*salidas)[0] += "ERROR: La ruta ingresada es erronea.\n"
		return
	}

	recorrer := ToStringArray(linodo.I_block[:])
	for continuar {
		crear := false
		inodoTemporal := -1

		// 1. Buscar si existe la carpeta
		for i := 0; i < 16; i++ {
			if inodoTemporal != -1 {
				break
			}

			if recorrer[i] == -1 {
				continue
			}

			posLectura = ToInt(sblock.S_block_start[:]) + (binary.Size(Bcarpetas{}) * recorrer[i])
			archivo.Seek(int64(posLectura), 0)
			binary.Read(archivo, binary.LittleEndian, &lcarpeta)

			for j := 0; j < 4; j++ {
				carpeta := ToString(lcarpeta.B_content[j].B_name[:])
				if carpeta == path[posicion] {
					linodo.I_atime = [30]byte{}
					copy(linodo.I_atime[:], []byte(time.Now().String()))
					posLectura = ToInt(sblock.S_inode_start[:]) + (binary.Size(Inodo{}) * inodo_leido)
					archivo.Seek(int64(posLectura), 0)
					binary.Write(archivo, binary.LittleEndian, &linodo)

					inodoTemporal = ToInt(lcarpeta.B_content[j].B_inodo[:])
					inodo_leido = inodoTemporal
					posicion += 1
					posLectura = ToInt(sblock.S_inode_start[:]) + (binary.Size(Inodo{}) * inodoTemporal)
					archivo.Seek(int64(posLectura), 0)
					binary.Read(archivo, binary.LittleEndian, &linodo)
					break
				}
			}
		}

		//Determinar si se puede crear
		if inodoTemporal == -1 {
			// Revisar que tenga permisos sobre la carpeta
			permisos := ToString(linodo.I_perm[:])
			ugo := 3 // 1 para dueño, 2 para grupo, 3 para otros
			acceso := false

			if sesion.id_user == ToString(linodo.I_uid[:]) {
				ugo = 1
			} else if sesion.id_grp == ToString(linodo.I_gid[:]) {
				ugo = 2
			}

			if ugo == 1 {
				if permisos[0] == '2' || permisos[0] == '3' || permisos[0] == '6' || permisos[0] == '7' {
					acceso = true
				}
			} else if ugo == 2 {
				if permisos[1] == '2' || permisos[1] == '3' || permisos[1] == '6' || permisos[1] == '7' {
					acceso = true
				}
			} else {
				if permisos[2] == '2' || permisos[2] == '3' || permisos[2] == '6' || permisos[2] == '7' {
					acceso = true
				}
			}

			if sesion.user == "root" {
				acceso = true
			}

			if acceso {
				if padre {
					crear = true
				} else if posicion == len(path)-1 {
					crear = true
				} else {
					error_padre = true
				}
			} else {
				continuar = false
				error_permisos = true
			}
		}

		//2. Crear la carpeta si no existe
		if crear {
			contador += 1
			bloque_usado := -1
			bloque_intermedio := -1
			cinodo := Inodo{}
			ccarpeta := Bcarpetas{}
			c := byte('-')

			for z := 0; z < 4; z++ {
				copy(ccarpeta.B_content[z].B_name[:], "-")
				copy(ccarpeta.B_content[z].B_inodo[:], []byte(strconv.Itoa(-1)))
			}

			//Buscar un espacio en los bloques directos
			recorrer = ToStringArray(linodo.I_block[:])
			for i := 0; i < 16; i++ {
				if inodoTemporal != -1 {
					break
				}

				if recorrer[i] != -1 {
					posLectura = ToInt(sblock.S_block_start[:]) + ((binary.Size(Bcarpetas{})) * (recorrer[i]))
					archivo.Seek(int64(posLectura), 0)
					binary.Read(archivo, binary.LittleEndian, &lcarpeta)

					for j := 0; j < 4; j++ {
						carpeta := ToString(lcarpeta.B_content[j].B_name[:])

						if carpeta == "-" {

							for a := 0; a < ToInt(sblock.S_inodes_count[:]); a++ {
								posLectura := ToInt(sblock.S_bm_inode_start[:]) + ((a) * (binary.Size(byte(0))))
								archivo.Seek(int64(posLectura), 0)
								binary.Read(archivo, binary.LittleEndian, &c)

								if c == byte(0) {
									inodoTemporal = a
									c = '1'
									archivo.Seek(int64(posLectura), 0)
									binary.Write(archivo, binary.LittleEndian, &c)
									break
								}

								if a == ToInt(sblock.S_inodes_count[:])-1 {
									(*salidas)[0] += "ERROR: No hay inodos disponibles\n."
									return
								}
							}

							//**Pendiente
							lcarpeta.B_content[j].B_name = [12]byte{}
							copy(lcarpeta.B_content[j].B_name[:], path[posicion])
							lcarpeta.B_content[j].B_inodo = [4]byte{}
							copy(lcarpeta.B_content[j].B_inodo[:], strconv.Itoa(inodoTemporal))
							posLectura = ToInt(sblock.S_block_start[:]) + (binary.Size(Bcarpetas{}) * recorrer[i])
							archivo.Seek(int64(posLectura), 0)
							binary.Read(archivo, binary.LittleEndian, &lcarpeta)

							//Actualizar el superbloque
							enteros := ToInt(sblock.S_free_blocks_count[:]) - 1
							sblock.S_free_blocks_count = [40]byte{}
							copy(sblock.S_free_blocks_count[:], strconv.Itoa(enteros))

							enteros = ToInt(sblock.S_free_inodes_count[:]) - 1
							sblock.S_free_inodes_count = [40]byte{}
							copy(sblock.S_free_inodes_count[:], strconv.Itoa(enteros))
							break
						}
					}

				} else {

					for a := 0; a < ToInt(sblock.S_inodes_count[:]); a++ {
						posLectura := ToInt(sblock.S_bm_inode_start[:]) + ((a) * (binary.Size(byte(0))))
						archivo.Seek(int64(posLectura), 0)
						binary.Read(archivo, binary.LittleEndian, &c)

						if c == byte(0) {
							inodoTemporal = a
							c = '1'
							archivo.Seek(int64(posLectura), 0)
							binary.Write(archivo, binary.LittleEndian, &c)
							break
						}

						if a == ToInt(sblock.S_inodes_count[:])-1 {
							(*salidas)[0] += "ERROR: No hay inodos disponibles\n."
							return
						}
					}

					for a := 0; a < ToInt(sblock.S_blocks_count[:]); a++ {
						posLectura := ToInt(sblock.S_bm_block_start[:]) + ((a) * (binary.Size(byte(0))))
						archivo.Seek(int64(posLectura), 0)
						binary.Read(archivo, binary.LittleEndian, &c)

						if c == byte(0) {
							bloque_intermedio = a
							c = 'c'
							archivo.Seek(int64(posLectura), 0)
							binary.Write(archivo, binary.LittleEndian, &c)
							break
						}

						if a == ToInt(sblock.S_blocks_count[:])-1 {
							(*salidas)[0] += "ERROR: No hay bloques disponibles\n."
							return
						}
					}

					for a := 0; a < ToInt(sblock.S_blocks_count[:]); a++ {
						posLectura := ToInt(sblock.S_bm_block_start[:]) + ((a) * (binary.Size(byte(0))))
						archivo.Seek(int64(posLectura), 0)
						binary.Read(archivo, binary.LittleEndian, &c)

						if c == byte(0) {
							bloque_usado = a
							c = 'c'
							archivo.Seek(int64(posLectura), 0)
							binary.Write(archivo, binary.LittleEndian, &c)
							break
						}

						if a == ToInt(sblock.S_blocks_count[:])-1 {
							(*salidas)[0] += "ERROR: No hay bloques disponibles\n."
							return
						}
					}

					//Escribir el nuevo bloque de carpeta
					ccarpeta.B_content[0].B_name = [12]byte{}
					copy(ccarpeta.B_content[0].B_name[:], path[posicion])
					ccarpeta.B_content[0].B_inodo = [4]byte{}
					copy(ccarpeta.B_content[0].B_inodo[:], strconv.Itoa(inodoTemporal))
					posLectura = ToInt(sblock.S_block_start[:]) + (binary.Size(Bcarpetas{}) * bloque_intermedio)
					archivo.Seek(int64(posLectura), 0)
					binary.Write(archivo, binary.LittleEndian, &ccarpeta)

					//Actualizar el inodo
					recorrer[i] = bloque_intermedio
					linodo.I_mtime = [30]byte{}
					copy(linodo.I_mtime[:], []byte(time.Now().String()))
					sliceTemp := ToByteArray(recorrer)
					copy(linodo.I_block[:], sliceTemp)
					posLectura = ToInt(sblock.S_inode_start[:]) + (binary.Size(Inodo{}) * inodo_leido)
					archivo.Seek(int64(posLectura), 0)
					binary.Write(archivo, binary.LittleEndian, &linodo)

					//Actualizar el superbloque
					enteros := ToInt(sblock.S_free_blocks_count[:]) - 2
					sblock.S_free_blocks_count = [40]byte{}
					copy(sblock.S_free_blocks_count[:], strconv.Itoa(enteros))

					enteros = ToInt(sblock.S_free_inodes_count[:]) - 1
					sblock.S_free_inodes_count = [40]byte{}
					copy(sblock.S_free_inodes_count[:], strconv.Itoa(enteros))
					break
				}
			}

			//Crear el bloque e inodo de carpeta
			copy(cinodo.I_uid[:], []byte(sesion.id_user))
			copy(cinodo.I_gid[:], []byte(sesion.id_grp))
			copy(cinodo.I_s[:], strconv.Itoa(0))
			copy(cinodo.I_atime[:], []byte(time.Now().String()))
			copy(cinodo.I_ctime[:], []byte(time.Now().String()))
			copy(cinodo.I_mtime[:], []byte(time.Now().String()))
			var bloques [16]string
			sliceTemp := []byte{}
			for i := 0; i < 16; i++ {
				bloques[i] = "-1"
			}
			bloques[0] = strconv.Itoa(bloque_usado)
			for _, i := range bloques {
				sliceTemp = append(sliceTemp, []byte(i)...)
				sliceTemp = append(sliceTemp, '!')
			}
			copy(cinodo.I_block[:], sliceTemp)
			cinodo.I_type[0] = byte('0')
			copy(cinodo.I_perm[:], "664")
			posLectura = ToInt(sblock.S_inode_start[:])  + (binary.Size(Inodo{}) * inodoTemporal);
			archivo.Seek(int64(posLectura), 0)
			binary.Write(archivo, binary.LittleEndian, cinodo)

			// CREAR E INICIAR EL BLOQUE DE CARPETAS
			ccarpeta = Bcarpetas{}
			for i := 0; i < 4; i++ {
				copy(ccarpeta.B_content[i].B_name[:], "-")
				copy(ccarpeta.B_content[i].B_inodo [:], []byte(strconv.Itoa(-1)))
			}

			// REGISTRAR EL INODO ACTUAL Y EL DEL PADRE
			ccarpeta.B_content[0].B_inodo = [4]byte{}
			copy(ccarpeta.B_content[0].B_name[:], ".")
			copy(ccarpeta.B_content[0].B_inodo[:], []byte(strconv.Itoa(inodoTemporal)))

			ccarpeta.B_content[1].B_inodo = [4]byte{}
			copy(ccarpeta.B_content[1].B_name[:], "..")
			copy(ccarpeta.B_content[1].B_inodo[:], []byte(strconv.Itoa(inodo_leido)))
			
			posLectura = ToInt(sblock.S_block_start[:]) + (binary.Size(Bcarpetas{}) * bloque_usado);
			archivo.Seek(int64(posLectura), 0)
			binary.Write(archivo, binary.LittleEndian, &ccarpeta)

			//Actualizar Variables
            posLectura = ToInt(sblock.S_inode_start[:]) + (binary.Size(Inodo{}) * inodoTemporal);
            archivo.Seek(int64(posLectura), 0)
            binary.Read(archivo, binary.LittleEndian, &linodo)
            inodo_leido = inodoTemporal;
            posicion += 1;

            //Actualizar el superbloque
			archivo.Seek(int64(posInicio), 0)
            binary.Write(archivo, binary.LittleEndian, &sblock)
		}

		//3. Finalizar el proceso si no existe la carpeta y no esta habilitado -p
        if inodoTemporal == -1 && !padre{
            continuar = false;
        }

        //4. Determinar si se finaliza el proceso 
        if linodo.I_type[0] == '1' || posicion == len(path){
            continuar = false;
        }
	}

	if error_padre {
        (*salidas)[0] += "ERROR: Error en la ruta. Una de las carpetas no existe.\n" 
        return
    }else if error_permisos {
        (*salidas)[0] += "ERROR: No posee los permisos para crear la carpeta.\n"
        return
    }else if posicion != len(path) {
        (*salidas)[0] += "ERROR: Ocurrió un error. No se pudo crear la carpeta.\n"
        return
    }

	(*salidas)[0] += "MENSAJE: Carpetas creadas correctamente.\n" 

}
