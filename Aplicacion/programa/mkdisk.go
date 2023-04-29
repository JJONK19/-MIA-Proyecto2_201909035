package programa

import (
	"encoding/binary"
	//"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)
func Mkdisk(parametros *[]string, salidas *[6]string) {
    // VARIABLES
    var paramFlag bool = true                     // Indica si se cumplen con los parametros del comando
    var required bool = true                      // Indica si vienen los parametros obligatorios
    var valid bool = true                         // Verifica que los valores de los parametros sean correctos
    var vacio byte = '0'
    var tamaño int = 0                            // Atributo >size
    var fit string = ""                           // Atributo >fit
    var fitChar byte = '0'                        // El fit se maneja como caracter
    var unidad string = ""                        // Atributo >unit
    var ruta string = ""                          // Atributo path
    mbr := MBR{}                                   // Para manejar el MBR

    // COMPROBACIÓN DE PARAMETROS
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
                (*salidas)[0] += "ERROR: El tamaño debe de ser un valor númerico.\n"
                //fmt.Println("ERROR: El tamaño debe de ser un valor númerico.")
                return
            }
        } else if tag == "fit" {
            fit = value
        } else if tag == "unit" {
            unidad = value
        } else if tag == "path" {
            ruta = value
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

    // COMPROBAR PARAMETROS OBLIGATORIOS
    if tamaño == 0 || ruta == "" {
        required = false
    }

    if !required {
        (*salidas)[0] += "ERROR: La instrucción mkdisk carece de todos los parametros obligatorios.\n"
        //fmt.Println("ERROR: La instrucción mkdisk carece de todos los parametros obligatorios.")
    }

    // VALIDACION DE PARAMETROS
    fit = strings.ToLower(fit)
    unidad = strings.ToLower(unidad)

    if tamaño <= 0 {
        (*salidas)[0] += "ERROR: El tamaño debe de ser mayor que 0.\n"
        //fmt.Println("ERROR: El tamaño debe de ser mayor que 0.")
        valid = false
    }

    if fit == "bf" || fit == "ff" || fit == "wf" || fit == "" {
    } else {
        (*salidas)[0] += "ERROR: Tipo de Fit Invalido.\n" 
        //fmt.Println("ERROR: Tipo de Fit Invalido.")
        valid = false
    }

    if unidad == "k" || unidad == "m" || unidad == "" {
    } else {
        (*salidas)[0] += "ERROR: Tipo de Unidad Invalido.\n"
        //fmt.Println("ERROR: Tipo de Unidad Invalido.")
        valid = false
    }

    if !valid {
        return
    }

    // PREPARACIÓN DE PARAMETROS - Determinar el alias del fit y pasar a bytes el tamaño
    if fit == "" || fit == "ff" {
        fitChar = 'f'
    } else if fit == "bf" {
        fitChar = 'b'
    } else if fit == "wf" {
        fitChar = 'w'
    }

    if unidad == "" || unidad == "m" {
        tamaño = tamaño * 1024 * 1024
    } else {
        tamaño = tamaño * 1024
	}

    // VERIFICAR QUE EL ARCHIVO NO EXISTA
    if _, err := os.Stat(ruta); !os.IsNotExist(err) {
        (*salidas)[0] += "ERROR: El archivo que desea crear ya existe.\n"
        //fmt.Println("ERROR: El archivo que desea crear ya existe.")
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

    // CREAR EL ARCHIVO BINARIO (DISCO) Y LLENARLO DE 0s
    archivo, err := os.Create(ruta)
    if err != nil {
        (*salidas)[0] += "Error creando archivo.\n"
        //fmt.Println("Error creando archivo:", err)
        return
    }
    defer archivo.Close()

    kb := make([]byte, 1024)
    for i := 0; i < 1024; i++ {
		kb[i] = 0
	}
    for i := 0; i < tamaño/1024; i++ {
        if _, err := archivo.Write(kb); err != nil {
            (*salidas)[0] += "Error escribiendo en archivo.\n"
            //fmt.Println("Error escribiendo en archivo:", err)
            return
        }
    }

    //CREAR EL MBR Y LLENARLO DE VALORES DEFAULT
    copy(mbr.Mbr_tamano[:], strconv.Itoa(tamaño))
    copy(mbr.Mbr_dsk_signature[:], strconv.Itoa(rand.Intn(9999)))
    copy(mbr.Mbr_fecha_creacion[:], []byte(time.Now().String()))
    mbr.Dsk_fit[0] = fitChar

    for i := 0; i < 4; i++ {
        copy(mbr.Mbr_partition[i].Part_name[:], []byte(""))
        mbr.Mbr_partition[i].Part_status[0] = vacio
        copy(mbr.Mbr_partition[i].Part_s[:], strconv.Itoa(0))
        mbr.Mbr_partition[i].Part_fit[0] = fitChar
        copy(mbr.Mbr_partition[i].Part_start[:], strconv.Itoa(-1))
    }

    //ESCRIBIR EL STRUCT EN EL DISCO
    archivo.Seek(0, 0)                   
    binary.Write(archivo, binary.LittleEndian, &mbr)
    (*salidas)[0] += "MENSAJE: Archivo creado correctamente.\n"
    //fmt.Println("MENSAJE: Archivo creado correctamente.")
}