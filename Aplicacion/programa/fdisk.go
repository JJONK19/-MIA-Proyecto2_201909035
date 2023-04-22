package programa

import (
	"sort"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func Fdisk(parametros *[]string) {
    //VARIABLES
    paramFlag := true 			//Indica si se cumplen con los parametros del comando
    required := true  			//Indica si vienen los parametros obligatorios
    valid := true     			//Verifica que los valores de los parametros sean correctos
    var tamaño int = 0   		//Atrubuto >size
    var fit string = ""    		//Atributo >fit
    var fit_char byte = '0' 	//El fit se maneja como caracter
    var unidad string = "" 		//Atributo >unit
    var ruta string = ""   		//Atributo path
    var tipo string = ""    	//Atributo >type
    var tipo_char byte = '0' 	//El tipo se maneja como char
    var nombre string = ""  	//Atributo name
    var comando string = "" 	//Indica si se crea/borra/expande al ejecutar
    var comando_cont int = 0	//Indica cuántas instrucciones se ingresaron en el comando

    //COMPROBACIÓN DE PARAMETROS
	for i := 1; i < len(*parametros); i++ {
        temp := (*parametros)[i]
        salida := regexp.MustCompile(`=`).Split(temp, -1)

        tag := salida[0]
        value := salida[1]

        // Pasar a minusculas
        tag = strings.ToLower(tag)

        if tag == "size" {
            var err error
            tamaño, err = strconv.Atoi(value)
            if err != nil {
                fmt.Println("ERROR: El tamaño debe de ser un valor númerico.")
                return
            }

			if comando == "" {
                comando = "create"
                comando_cont += 1
            }

        } else if tag == "fit" {
            fit = value
        } else if tag == "unit" {
            unidad = value
        } else if tag == "path" {
            ruta = value
        } else if tag == "type" {
            tipo = value
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
    if nombre == "" || ruta == "" {
        required = false
    }

    if !required {
        fmt.Println("ERROR: La instrucción fdisk carece de todos los parametros obligatorios.")
        return
    }

    //VALIDACION DE PARAMETROS
	fit = strings.ToLower(fit)
	tipo = strings.ToLower(tipo)
	unidad = strings.ToLower(unidad)

	if comando == "create" && tamaño <= 0 {
        fmt.Println("ERROR: El tamaño de la nueva partición debe de ser mayor que 0.")
        valid = false
    }

    if fit == "bf" || fit == "ff" || fit == "wf" || fit == "" {
    } else {
        fmt.Println("ERROR: Tipo de Fit Invalido.")
        valid = false
    }

    if tipo == "p" || tipo == "e" || tipo == "l" || tipo == "" {
    } else {
        fmt.Println("ERROR: Tipo de Particion Invalido.")
        valid = false
    }

    if unidad == "k" || unidad == "m" || unidad == "b"|| unidad == "" {
    } else {
        fmt.Println("ERROR: Tipo de Unidad Invalido.")
        valid = false
    }

    if comando == "" {
        fmt.Println("ERROR: La instrucción carece de una tarea (añadir, borrar o modificar).")
        valid = false
    }

    if comando_cont != 1 {
        fmt.Println("ERROR: La instrucción posee más de una tarea (añadir, borrar o modificar).")
        valid = false
    }

	if !valid {
        return;
    }

	//PREPARACIÓN DE PARAMETROS - Determinar el alias del fit y pasar a bytes el tamaño
	if fit == "ff" {
		fit_char = 'f'
	} else if fit == "bf" {
		fit_char = 'b'
	} else if fit == "wf" || fit == "" {
		fit_char = 'w'
	}

	if tipo == "" || tipo == "p" {
		tipo_char = 'p'
	} else if tipo == "e" {
		tipo_char = 'e'
	} else if tipo == "l" {
		tipo_char = 'l'
	}

	if unidad == "m" {
		tamaño = tamaño * 1024 * 1024
	} else if unidad == "" || unidad == "k" {
		tamaño = tamaño * 1024
	}

	// VERIFICAR QUE EL ARCHIVO EXISTA
	archivo, err := os.OpenFile(ruta, os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("ERROR: El disco no existe.")
		return
	} else {
		archivo.Close()
	}

	//SEPARAR TIPO DE OPERACIÓN Y EJECUTARLA
	if comando == "create" {
		crear_particion(&tamaño, &tipo_char, &ruta, &nombre, &fit_char)
	}
}

func crear_particion(tamaño *int, tipo *byte, ruta, nombre *string, fit *byte) {
	//VARIABLES GENERALES
	archivo, err := os.OpenFile(*ruta, os.O_RDWR, 0666) //Disco con el que se va a trabajar
	if err != nil {
		fmt.Println("ERROR: El disco no existe.")
		return
	}
	defer archivo.Close()

	mbr := MBR{} //Para leer el mbr del disco
	posicion := -1 //Posicion de la particion en el mbr

	//COLOCAR EL PUNTERO DE LECTURA EN LA CABECERA Y LEER MBR
	archivo.Seek(0, 0)
	binary.Read(archivo, binary.LittleEndian, &mbr)

    //=============== PARTICIONES LÓGICAS ===============
    if *tipo == 'l' {
        //VARIABLES
        var posExtendida int //Posicion de la cabecera de la extendida
        var finExtendida int //Posicion donde acaba la extendida
        ebr := EBR{} //Auxiliar para leer los ebr
        var espacios OrdenarLibreL //Posiciones de los espacios libres
        var cabecera_visitada bool = false //Indica si es la cabecera la revisada
        var continuar bool = true //Salir del while de busqueda de espacios
        var existe bool = false //Indica si existe la particion
        var contador int = 0 //Numero de particiones en la extendida
        var posEspacio int = -1 //Posicion del espacio a usar dentro del vector de espacios

        //BUSCAR LA PARTICIÓN EXTENDIDA
        for i := 0; i < 4; i++ {
            if mbr.Mbr_partition[i].Part_type[0] == byte('e') {
                posicion = i
                break
            }
        }

        if posicion == -1 {
            fmt.Println("ERROR: No existe una partición extendida. Partición no creada.")
            return
        }

        // MOVER EL PUNTERO A LA EXTENDIDA Y LEER EL EBR CABECERA
        posExtendida =  ToInt(mbr.Mbr_partition[posicion].Part_start[:])
        finExtendida = posExtendida + ToInt(mbr.Mbr_partition[posicion].Part_s[:])
        archivo.Seek(int64(posExtendida), 0)
        binary.Read(archivo, binary.LittleEndian, &ebr)

        // BUSCAR SI ESTA REPETIDA LA PARTICION Y ENCONTRAR ESPACIOS VACIOS
        for continuar {
            if strings.Trim(string(ebr.Part_name[:]), "\x00") == *nombre {
                existe = true
                break
            }

            if !cabecera_visitada {
                contador++
                if ToInt(ebr.Part_s[:]) == 0 {
                    temp := LibreL {
                        cabecera: true,
                        inicioEBR: posExtendida,
                        finLogica: ToInt(ebr.Part_start[:]),
                    }
                    if ToInt(ebr.Part_next[:]) == -1 {
                        temp.tamaño = finExtendida - temp.finLogica
                    } else {
                        temp.tamaño = ToInt(ebr.Part_next[:]) - temp.finLogica
                    }
        
                    if temp.tamaño > 0 {
                        espacios = append(espacios, temp)
                    }
                } else {
                    temp := LibreL {
                        cabecera: false,
                        inicioEBR: posExtendida,
                        finLogica: ToInt(ebr.Part_start[:]) + ToInt(ebr.Part_s[:]) - 1,
                    }
                    if ToInt(ebr.Part_next[:]) == -1 {
                        temp.tamaño = finExtendida - (temp.finLogica + 1)
                    } else {
                        temp.tamaño = ToInt(ebr.Part_next[:]) - (temp.finLogica + 1)
                    }
        
                    if temp.tamaño > 0 {
                        espacios = append(espacios, temp)
                    }
                }
        
                cabecera_visitada = true
                if ToInt(ebr.Part_next[:]) == -1 {
                    continuar = false
                } else {
                    posExtendida = ToInt(ebr.Part_next[:])
                    archivo.Seek(int64(posExtendida), 0)
                    binary.Read(archivo, binary.LittleEndian, &ebr)
                }
            } else {
                contador++
                temp := LibreL {
                    cabecera: false,
                    inicioEBR: posExtendida,
                    finLogica: ToInt(ebr.Part_start[:]) + ToInt(ebr.Part_s[:]) - 1,
                }
                if ToInt(ebr.Part_next[:]) == -1 {
                    temp.tamaño = finExtendida - (temp.finLogica + 1)
                } else {
                    temp.tamaño = ToInt(ebr.Part_next[:]) - (temp.finLogica + 1)
                }
        
                if temp.tamaño > 0 {
                    espacios = append(espacios, temp)
                }
                if ToInt(ebr.Part_next[:]) == -1 {
                    continuar = false
                } else {
                    posExtendida = ToInt(ebr.Part_next[:])
                    archivo.Seek(int64(posExtendida), 0)
                    binary.Read(archivo, binary.LittleEndian, &ebr)
                }
            }
        }   

        if existe {
            fmt.Println("ERROR: Ya existe una partición lógica con ese nombre.")
            return
        }
        
        // Verificar que no sobrepase el límite
        if contador >= 24 {
            fmt.Println("ERROR: Se ha alcanzado el máximo de particiones lógicas (24 particiones).")
            return
        }

        // Comprobar que hay espacios vacíos en el disco
        if len(espacios) == 0 {
            fmt.Println("ERROR: La partición expandida se encuentra totalmente ocupada.")
            return
        }

        // BUSCAR ESPACIO PARA INSERTAR - FIRST FIT
        if *fit == byte('f') {
            for i := 0; i < len(espacios); i++ {
                temp := espacios[i]
                if (*tamaño + int(binary.Size(EBR{}))) <= temp.tamaño {
                    posEspacio = i
                    break
                }
            }

            if posEspacio == -1 {
                fmt.Println("ERROR: No hay espacio disponible en la partición extendida.")
                return
            }
        }

        // BUSCAR ESPACIO PARA INSERTAR - BEST FIT
        if *fit == byte('b') {
            sort.Sort(OrdenarLibreL(espacios))
            for i := 0; i < len(espacios); i++ {
                temp := espacios[i]
                if (*tamaño + int(binary.Size(EBR{}))) <= temp.tamaño {
                    posEspacio = i
                    break
                }
            }
            
            if posEspacio == -1 {
                fmt.Println("ERROR: No hay espacio disponible en la partición extendida.")
                return
            }
        }

        // BUSCAR ESPACIO PARA INSERTAR - WORST FIT
        if *fit == byte('w') {
            sort.Sort(OrdenarLibreL(espacios))
            if *tamaño <= espacios[len(espacios)-1].tamaño {
                posEspacio = len(espacios) - 1
            }

            if posEspacio == -1 {
                fmt.Println("ERROR: No hay espacio disponible en la partición extendida.")
                return
            }
        }

        //CREAR LA PARTICION
        if espacios[posEspacio].cabecera{
            //Leer y reescribir EBR padre
            archivo.Seek(int64(espacios[posEspacio].inicioEBR), 0)
            binary.Read(archivo, binary.LittleEndian, &ebr)
            copy(ebr.Part_name[:], []byte(*nombre))
            ebr.Part_fit[0] = *fit
            copy(ebr.Part_s[:], strconv.Itoa(*tamaño))
            ebr.Part_status[0] = '0'
            archivo.Seek(int64(espacios[posEspacio].inicioEBR), 0)
            binary.Write(archivo, binary.LittleEndian, &ebr)
            fmt.Println("MENSAJE: Particion lógica creada correctamente.")
        }else{
            posEbr := espacios[posEspacio].finLogica + 1
            //Leer y reescribir el EBR padre
            archivo.Seek(int64(espacios[posEspacio].inicioEBR), 0)
            binary.Read(archivo, binary.LittleEndian, &ebr)
            copy(ebr.Part_next[:], strconv.Itoa(posEbr))
            archivo.Seek(int64(espacios[posEspacio].inicioEBR), 0)
            binary.Write(archivo, binary.LittleEndian, &ebr)
            
            fmt.Println(posEbr)
            fmt.Println(espacios[posEspacio].inicioEBR)
            //Crear nueva particion (EBR)
            ebr.Part_fit[0] = *fit
            copy(ebr.Part_name[:], []byte(*nombre))
            ebr.Part_status[0] = '0'
            copy(ebr.Part_s[:], strconv.Itoa(*tamaño))
            copy(ebr.Part_start[:], strconv.Itoa(posEbr + int(binary.Size(EBR{}))))
            copy(ebr.Part_next[:], strconv.Itoa(-1))
            archivo.Seek(int64(posEbr), 0)
            binary.Write(archivo, binary.LittleEndian, &ebr)
            fmt.Println("MENSAJE: Particion lógica creada correctamente.")
        }
    }

     //=============== PARTICIONES PRIMARIAS / EXPANDIDA ===============
     if *tipo == byte('p') || *tipo == byte('e') {
        // VARIABLES
        var extendedExist, existe bool = false, false // Verifica que no exista una extendida / Indica si existe la particion
        var posiciones OrdenarPosicion // Posiciones de las particiones
        var espacios OrdenarLibre // Espacios vacios entre particiones
        var posEspacio int = -1 // Posicion en la lista de espacios libres
    
        // BUSCAR SI HAY ESPACIO EN EL MBR Y DE PASO VER SI EXISTE LA EXTENDIDA
        for i := 0; i < 4; i++ {
            if len(strings.Trim(string(mbr.Mbr_partition[i].Part_name[:]), "\x00")) == 0 {
                posicion = i
            }
    
            if mbr.Mbr_partition[i].Part_type[0] == byte('e') {
                extendedExist = true
            }
    
            if posicion != -1 {
                break
            }
        }
    
        if extendedExist && *tipo == 'e' {
            fmt.Println("ERROR: Solo puede existir una partición extendida a la vez.")
            return
        }
    
        if posicion == -1 {
            fmt.Println("ERROR: Limite de particiones alcanzado (4). Elimine una particion para continuar.")
            return
        }

        //DETERMINAR SI EXISTE LA PARTICION Y MARCAR LAS POSICIONES DE CADA PARTICION
        for i := 0; i < 4; i++ {
            if len(strings.Trim(string(mbr.Mbr_partition[i].Part_name[:]), "\x00")) != 0 {
                var temp Position
                temp.inicio = ToInt(mbr.Mbr_partition[i].Part_start[:])
                temp.fin = ToInt(mbr.Mbr_partition[i].Part_start[:]) + ToInt(mbr.Mbr_partition[i].Part_s[:]) - 1
                temp.nombre = string(mbr.Mbr_partition[i].Part_name[:])
                posiciones = append(posiciones, temp)

                if *nombre == strings.Trim(string(mbr.Mbr_partition[i].Part_name[:]), "\x00"){
                    existe = true
                }
            }
        }

        if existe {
            fmt.Println("ERROR: Ya existe una particion con el nombre indicado.")
            return
        }

        //ORDENAR LAS PARTICIONES
        sort.Sort(OrdenarPosicion(posiciones))

        // CREAR LA LISTA DE ESPACIOS VACIOS
        if len(posiciones) == 0 {
            var temp Libre
            temp.inicio = int(binary.Size(MBR{})) + 1
            temp.tamaño = ToInt(mbr.Mbr_tamano[:]) - int(binary.Size(MBR{})) + 1
            espacios = append(espacios, temp)
        }else{
            for i := 0; i < len(posiciones); i++ {
                var temp Libre
                x := &posiciones[i]
                free := 0
                if i == 0 && i != (len(posiciones)-1) {
                    // Espacio entre el inicio y la primera particion
                    free = x.inicio - int(binary.Size(MBR{})) + 1
                    if free > 0 {
                        temp.inicio = int(binary.Size(MBR{})) + 1
                        temp.tamaño = free
                        espacios = append(espacios, temp)
                    }
            
                    // Espacio entre la primera particion y la siguiente
                    y := &posiciones[i+1]
                    free = y.inicio - (x.fin + 1)
                    if free > 0 {
                        temp.inicio = x.fin + 1
                        temp.tamaño = free
                        espacios = append(espacios, temp)
                    }
                } else if i == 0 && i == (len(posiciones)-1) {
                    // Espacio entre el inicio y la primera particion
                    free = x.inicio - int(binary.Size(MBR{})) + 1
                    if free > 0 {
                        temp.inicio = int(binary.Size(MBR{})) + 1
                        temp.tamaño = free
                        espacios = append(espacios, temp)
                    }
            
                    // Espacio entre la primera particion y el fin
                    free = ToInt(mbr.Mbr_tamano[:]) - (x.fin + 1)
                    if free > 0 {
                        temp.inicio = x.fin + 1
                        temp.tamaño = free
                        espacios = append(espacios, temp)
                    }
                } else if i != (len(posiciones)-1) {
                    y := &posiciones[i+1]
                    free = y.inicio - (x.fin + 1)
                    if free > 0 {
                        temp.inicio = x.fin + 1
                        temp.tamaño = free
                        espacios = append(espacios, temp)
                    }
                } else {
                    free = ToInt(mbr.Mbr_tamano[:]) - (x.fin + 1)
                    if free > 0 {
                        temp.inicio = x.fin + 1
                        temp.tamaño = free
                        espacios = append(espacios, temp)
                    }
                }
            }            
        } 

        if len(espacios) == 0 {
            fmt.Println("ERROR: No hay espacio disponible en el disco (1).")
            return
        }
        
        // BUSCAR ESPACIO PARA INSERTAR - FIRST FIT
        if *fit == byte('f') {
            for i := 0; i < len(espacios); i++ {
                temp := espacios[i]
                if *tamaño <= temp.tamaño {
                    posEspacio = i
                    break
                }
            }
        
            if posEspacio == -1 {
                fmt.Println("ERROR: No hay espacio disponible en el disco (2).")
                archivo.Close()
                return
            }
        }

        // BUSCAR ESPACIO PARA INSERTAR - BEST FIT
        if *fit == byte('b') {
            sort.Sort(OrdenarLibre(espacios))
            for i := 0; i < len(espacios); i++ {
                temp := espacios[i]
                if *tamaño <= temp.tamaño {
                    posEspacio = i
                    break
                }
            }
    
            if posEspacio == -1 {
                fmt.Println("ERROR: No hay espacio disponible en el disco. (3)")
                return
            }
        }

        // BUSCAR ESPACIO PARA INSERTAR - WORST FIT
        if *fit == byte('w') {
            sort.Sort(OrdenarLibre(espacios))
            
            if *tamaño <= espacios[len(espacios)-1].tamaño {
                posEspacio = len(espacios) - 1
            }
        
            if posEspacio == -1 {
                fmt.Println("ERROR: No hay espacio disponible en el disco. (4)")
                return
            }
        }

        //CREAR PARTICION
        mbr.Mbr_partition[posicion].Part_fit[0] = *fit
        copy(mbr.Mbr_partition[posicion].Part_name[:], []byte(*nombre))
        mbr.Mbr_partition[posicion].Part_status[0] = '0'
        copy(mbr.Mbr_partition[posicion].Part_s[:], strconv.Itoa(*tamaño))
        mbr.Mbr_partition[posicion].Part_type[0] = *tipo
        copy(mbr.Mbr_partition[posicion].Part_start[:], strconv.Itoa(espacios[posEspacio].inicio))   

        archivo.Seek(0,0)                   
        binary.Write(archivo, binary.LittleEndian, &mbr)

        //CREAR EL EBR INICIAL EN CASO DE SER EXTENDIDA
        if *tipo == byte('e'){
            ebr := EBR{}
            copy(ebr.Part_name[:], []byte(""))
            copy(ebr.Part_next[:], []byte(strconv.Itoa(-1)))
            copy(ebr.Part_start[:], strconv.Itoa(espacios[posEspacio].inicio + 1 + int(binary.Size(EBR{}))))
            copy(ebr.Part_s[:], strconv.Itoa(0))
            ebr.Part_status[0] = '0'
            archivo.Seek(int64(espacios[posEspacio].inicio), 0)
            binary.Write(archivo, binary.LittleEndian, &ebr)
        }
        fmt.Println("MENSAJE: Particion creada correctamente.")
    }

}
