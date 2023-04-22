package programa

import(
	"regexp"
	"strings"
	"fmt"
	"os"
	"encoding/binary"
	"path/filepath"
	"strconv"
	"time"
)
func Mount(parametros *[]string, discos *[]Disco){
	//VARIABLES
	var paramFlag bool = true //Indica si se cumplen con los parametros del comando
	var required bool = true //Indica si vienen los parametros obligatorios
	var archivo *os.File //Sirve para verificar que el archivo exista
	var ruta string = "" //Atributo path
	var nombre string = "" //Atributo name
	var posEBR int = -1 //Posicion del EBR de la particion
	var posMBR int = -1 //Posicion de la particion en el MBR
	var posLogica int = -1 //Posicion de la logica
	var mbr MBR //Auxiliar para leer el MBR
	var ebr EBR //Auxiliar para leer EBRs
	var tamaño int //Tamaño de la particion
	var discName string //Nombre del disco que contiene la particion
	
	//Nuevo
	var extendidaMontada bool = false //Indica si la extendida es la que se va a montar

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
		} else if tag == "name" {
			nombre = value
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
	if ruta == "" || nombre == "" {
		required = false
	}
	
	if !required {
		fmt.Println("ERROR: La instrucción mount carece de todos los parametros obligatorios.")
		return
	}

	//VERIFICAR QUE EL ARCHIVO EXISTA
	archivo, err := os.OpenFile(ruta, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("ERROR: El disco no existe.")
		return
	}
	defer archivo.Close()

	//LEER EL MBR
	archivo.Seek(0, 0)
	binary.Read(archivo, binary.LittleEndian, &mbr)

	// BUSCAR LA PARTICION. PRIMERO EN EL MBR, LUEGO EN LA EXTENDIDA SI NO ESTÁ
	for i := 0; i < 4; i++ {
		temp := mbr.Mbr_partition[i]
		if strings.Trim(string(temp.Part_name[:]), "\x00") == nombre {
			if temp.Part_type[0] != 'e' {
				posMBR = i
				tamaño = ToInt(temp.Part_s[:])
				break
			}else {
				posLogica = ToInt(temp.Part_start[:])
				tamaño = 0
				extendidaMontada = true
			}
		}
	}
	
	if posMBR == -1 && !extendidaMontada {
		// Buscar la extendida
		for i := 0; i < 4; i++ {
			temp := mbr.Mbr_partition[i]
			if temp.Part_type[0] == 'e' {
				posEBR = ToInt(temp.Part_start[:])
				break
			}
		}
	
		// Buscar en la extendida
		if posEBR == -1 {
			fmt.Println("ERROR: La partición no existe.")
			return
		}
	
		// Leer cabecera de la particion
		archivo.Seek(int64(posEBR), 0)
		binary.Read(archivo, binary.LittleEndian, &ebr)
	
		// Determinar si existe la particion
		continuar := true
		existe := false // Indica si existe la particion
		for continuar {
			// Si encuentra el nombre, termina el proceso
			if strings.Trim(string(ebr.Part_name[:]), "\x00")== nombre {
				existe = true
				posLogica = posEBR
				tamaño = ToInt(ebr.Part_s[:])
				break
			}
	
			if ToInt(ebr.Part_next[:]) == -1 {
				continuar = false
			} else {
				posEBR = ToInt(ebr.Part_next[:])
				archivo.Seek(int64(posEBR), 0)
				binary.Read(archivo, binary.LittleEndian, &ebr)
			}
		}
	
		if !existe {
			fmt.Println("ERROR: La partición no existe.")
			archivo.Close()
			return
		}
	
	}

	//EXTRAER DE LA RUTA EL NOMBRE DEL DISCO
	discName = filepath.Base(ruta)

	//MONTAR LA PARTICION
	//1. SI LA LISTA ESTA VACIA, SE AÑADE LA PARTICION Y EL DISCO A LA LISTA
	//2. SI EL DISCO NO EXISTE, SE AÑADE JUNTO A LA PARTICION SIN REVISAR
	//3. SI EL DISCO EXISTE, SE VERIFICA QUE NO ESTE MONTADA LA PARTICION PARA AÑADIRLA

	if len(*discos) != 0 {
		//Buscar el disco
		posDisco := -1
		for i, temp := range *discos {
			if temp.nombre == discName {
				posDisco = i
				break
			}
		}

		//Añadir el disco y la particion si no existe (Caso 2)
		if posDisco == -1 {
			temp := Disco{
				nombre:   discName,
				ruta:     ruta,
				contador: 0,
			}
			nueva := Montada{
				id:      "35" + strconv.Itoa(temp.contador) + string(byte(65 + len(*discos))),
				posEBR:  posLogica,
				posMBR:  posMBR,
				nombre:  nombre,
				tamaño:  tamaño,
			}
			temp.contador++
			temp.particiones = append(temp.particiones, nueva)
			*discos = append(*discos, temp)
		}

		//Buscar si la particion no existe y montar el disco (Caso 3)
		if posDisco != -1 {
			temp := &(*discos)[posDisco]
			for _, t := range temp.particiones {
				if t.nombre == nombre {
					fmt.Printf("ERROR: La partición %s se encuentra montada.\n", nombre)
					return
				}
			}

			nueva := Montada{
				id:      "35" + strconv.Itoa(temp.contador) + string(byte(65 + len(*discos))),
				posEBR:  posLogica,
				posMBR:  posMBR,
				nombre:  nombre,
				tamaño:  tamaño,
			}
			temp.contador++
			temp.particiones = append(temp.particiones, nueva)
		}
	}

	if len(*discos) == 0 {
		//Añadir el disco y la particion (Caso 1)
		var temp Disco
		temp.nombre = discName
		temp.ruta = ruta
	
		var nueva Montada
		nueva.id = "35" + strconv.Itoa(temp.contador) + string(byte(65 + len(*discos)))
		temp.contador++
		nueva.posEBR = posLogica
		nueva.posMBR = posMBR
		nueva.nombre = nombre
		nueva.tamaño = tamaño
		temp.particiones = append(temp.particiones, nueva)
		*discos = append(*discos, temp)
	}
	
	//ESCRIBIR EL EBR/MBR Y EL SUPERBLOQUE
	if posLogica == -1 {
		//Cambiar el estado de la particion en el MBR
		archivo.Seek(0, 0)
		binary.Read(archivo, binary.LittleEndian, &mbr)
		mbr.Mbr_partition[posMBR].Part_status[0] = '1'
	
		archivo.Seek(0, 0)
		binary.Write(archivo, binary.LittleEndian, &mbr)

		//Actualizar el registro del Super Bloque
		var bloque Sbloque
		archivo.Seek(int64(ToInt(mbr.Mbr_partition[posMBR].Part_start[:])), 0)
		binary.Read(archivo, binary.LittleEndian, &bloque)
		
		if bloque.S_filesystem_type[0] == byte('2') || bloque.S_filesystem_type[0] == byte('3') {
			copy(bloque.S_mnt_count[:] , strconv.Itoa(ToInt(bloque.S_mnt_count[:]) + 1))
			copy(bloque.S_mtime[:], []byte(time.Now().String()))

			archivo.Seek(int64(ToInt(mbr.Mbr_partition[posMBR].Part_start[:])), 0)
			binary.Write(archivo, binary.LittleEndian, &bloque)
		}
	
		fmt.Println("MENSAJE: Particion montada correctamente.")
	} else {
		//Cambiar el estado de la particion en el MBR
		archivo.Seek(int64(posLogica), 0)
		binary.Read(archivo, binary.LittleEndian, &ebr)
		ebr.Part_status[0] = byte('1')
	
		archivo.Seek(int64(posLogica), 0)
		binary.Write(archivo, binary.LittleEndian, &ebr)
	
		//Actualizar el registro del Super Bloque
		var bloque Sbloque
		archivo.Seek(int64(ToInt(ebr.Part_s[:])), 0)
		binary.Read(archivo, binary.LittleEndian, &bloque)
		
		if bloque.S_filesystem_type[0] == byte('2') || bloque.S_filesystem_type[0] == byte('3') {
			copy(bloque.S_mnt_count[:] , strconv.Itoa(ToInt(bloque.S_mnt_count[:]) + 1))
			copy(bloque.S_mtime[:], []byte(time.Now().String()))

			archivo.Seek(int64(ToInt(ebr.Part_s[:])), 0)
			binary.Write(archivo, binary.LittleEndian, &bloque)
		}
		fmt.Println("MENSAJE: Particion montada correctamente.")
	}

	//Mostrar el listado de particiones montadas
	fmt.Println()
    fmt.Println("LISTADO DE PARTICIONES MONTADAS")
    fmt.Println()

    if len(*discos) == 0 {
        return
    }

    c_parusadas := 1
    for i := 0; i < len(*discos); i++ {

        if len((*discos)[i].particiones) == 0 {
            continue
        }

        fmt.Println((*discos)[i].nombre, ":")

        for j := 0; j < len((*discos)[i].particiones); j++ {
            imprimir := fmt.Sprintf("%d. %s", c_parusadas, (*discos)[i].particiones[j].nombre)
            fmt.Println(imprimir)
            c_parusadas += 1
        }
    }

}