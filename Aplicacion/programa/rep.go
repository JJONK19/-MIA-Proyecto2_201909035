package programa

import (
	"fmt"
	"regexp"
	"strings"
    "unicode"
    "os"
    "path/filepath"
    "encoding/binary"
    "os/exec"
    "sort"
    "strconv"
    "math"
)

func Rep(parametros *[]string, discos *[]Disco) {
    var paramFlag bool = true // Indica si se cumplen con los parametros del comando
	var required bool = true // Indica si vienen los parametros obligatorios
	var ruta string = "" // Atributo path
	var nombre string= "" // Atributo name
	var id string = "" // Atributo ID
	var rutaS string = "" // Atributo ruta
	var diskName string // Nombre del disco sin los numeros del ID
	var posDisco int = -1 // Posicion del disco en la lista
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
            fmt.Printf("ERROR: El parametro %s no es valido.\n", tag)
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
        fmt.Println("ERROR: La instrucción rep carece de todos los parametros obligatorios.")
        return
    }

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

    //BUSCAR LA PARTICION DENTRO DEL DISCO MONTADO
    tempD := (*discos)[posDisco]
    for i, temp := range tempD.particiones {
        if temp.id == id {
            posParticion = i
            break
        }
    }

    if posParticion == -1 {
        fmt.Println("ERROR: No existe una partición montada con ese ID.")
        return
    }

    // CREAR DIRECTORIOS EN CASO NO EXISTAN
    if err := os.MkdirAll(filepath.Dir(ruta), os.ModePerm); err != nil {
        fmt.Println("Error creando directorios:", err)
        return
    }

    // BORRAR EL ARCHIVO EN CASO YA EXISTA
    if err := os.Remove(ruta); err != nil && !os.IsNotExist(err) {
        fmt.Println("Error borrando archivo:", err)
        return
    }

    //SEPARAR TIPO DE INSTRUCIION Y EJECUTARLA
    nombre = strings.ToLower(nombre)

    switch nombre {
    case "mbr":
        Mbr(discos, posDisco, &ruta)
    case "disk":
        Disk(discos, posDisco, &ruta)
    case "tree":
        Tree(discos, posDisco, posParticion, &ruta)
    case "sb":
        Sb(discos, posDisco, posParticion, &ruta)
    case "file":
        File(discos, posDisco, posParticion, &ruta, &rutaS)
    default:
        fmt.Println("ERROR: Tipo de reporte invalido.")
    }
}

func Mbr(discos *[]Disco, posDisco int, ruta *string){
    var codigo string //Contenedor del codigo del dot
    uso := (*discos)[posDisco] //Disco en uso
    var mbr MBR //Para leer el mbr
    var posExtendida int //Posicion para leer la extendida
    var ebr EBR //Para leer los ebr de las particiones logicas
    var comando string //Instruccion a mandar a la consola para generar el comando

    //VERIFICAR QUE EXISTA EL ARCHIVO
    archivo, err := os.OpenFile(uso.ruta, os.O_RDWR, 0644)
    if err != nil {
        fmt.Println("ERROR: No se encontro el disco.")
        return
    }
    defer archivo.Close()

    //LEER EL MBR
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
        if len(strings.Trim(string(mbr.Mbr_partition[i].Part_name[:]), "\x00")) == 0 {
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
            }else{
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
    fmt.Println("MENSAJE: Reporte MBR creado correctamente.")
}

func Disk(discos *[]Disco, posDisco int, ruta *string){
    //VARIABLES
    var codigo string = "" //Contenedor del código del dot
    uso := (*discos)[posDisco] //Disco en uso
    var mbr MBR //Para leer el mbr
    var ebr EBR //Para leer los ebr de las particiones lógicas
    var comando string //Instrucción a mandar a la consola para generar el comando
    var size float64 //Tamaño del disco
    finExtendida := -1
    posEBR := -1
    var posiciones OrdenarPosicion
    var porcentaje int //Maneja los porcentajes a escribir en el reporte

    //VERIFICAR QUE EXISTA EL ARCHIVO
    archivo, err := os.OpenFile(uso.ruta, os.O_RDWR, 0644)
    if err != nil {
        fmt.Println("ERROR: No se encontro el disco.")
        return
    }
    defer archivo.Close()

    //LEER EL MBR
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
        if len(strings.Trim(string(mbr.Mbr_partition[i].Part_name[:]), "\x00")) == 0 {
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
    }else{
        for i := 0; i < len(posiciones); i++ {
            x := &posiciones[i]
            free := 0
        
            if i == 0 && i != (len(posiciones)-1) {
                free = x.inicio - int(binary.Size(MBR{})) + 1
        
                if free > 0 {
                    codigo += "<TD ROWSPAN='3' WIDTH='100' BGCOLOR='#3FA796'>LIBRE<BR/>"
                    porcentaje = int(math.Round(float64(free) / size * 100))
                    codigo += strconv.Itoa(porcentaje)
                    codigo += "% del disco"
                    codigo += "</TD>"
                }
        
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
            }else if i == 0 && i == (len(posiciones)-1) {
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
            }else if i != len(posiciones)-1 {
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
            }else{
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
        continuar := true         //Sirve para salir del while
        var free int

        for continuar{
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
            }else{
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
    fmt.Println("MENSAJE: Reporte DISKS creado correctamente.")
}

func Tree(discos *[]Disco, posDisco int, posParticion int, ruta *string){
    
}

func Sb(discos *[]Disco, posDisco int, posParticion int, ruta *string){
    //VARIABLES
    var codigo string //Contenedor del codigo del dot
    disco_uso := (*discos)[posDisco] //Disco en uso
    part_uso := &disco_uso.particiones[posParticion] //Particion Montada
    var archivo *os.File //Para leer el archivo
    var mbr MBR //Para leer el mbr
    var ebr EBR //Para leer los ebr de las particiones logicas
    var comando string //Instruccion a mandar a la consola para generar el comando
    var posInicio int //Posicion donde inicia la particion
    var sblock Sbloque //Para leer el superbloque

    //VERIFICAR QUE EXISTA EL ARCHIVO
    archivo, err := os.OpenFile(disco_uso.ruta, os.O_RDWR, 0644)
    if err != nil {
    fmt.Println("ERROR: No se encontro el disco.")
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
    fmt.Println("MENSAJE: Reporte DISKS creado correctamente.")
}

func File(discos *[]Disco, posDisco int, posParticion int, ruta *string, ruta_contenido *string){

}