package programa

import (
	"fmt"
	"regexp"
	"strings"
    "unicode"
    "os"
    "path/filepath"
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

}

func Disk(discos *[]Disco, posDisco int, ruta *string){

}

func Tree(discos *[]Disco, posDisco int, posParticion int, ruta *string){

}

func Sb(discos *[]Disco, posDisco int, posParticion int, ruta *string){

}

func File(discos *[]Disco, posDisco int, posParticion int, ruta *string, ruta_contenido *string){

}