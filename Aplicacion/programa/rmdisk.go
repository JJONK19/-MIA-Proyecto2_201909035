package programa

import (
	"os"
	"regexp"
	"strings"
	"fmt"
)

func Rmdisk(parametros *[]string) {
	//VARIABLES
	paramFlag := true //Indica si se cumplen con los parametros del comando
	required := true //Indica si vienen los parametros obligatorios
	ruta := "" //Atributo path

	// COMPROBACIÓN DE PARAMETROS
    for i := 1; i < len(*parametros); i++ {
        temp := (*parametros)[i]
        salida := regexp.MustCompile(`=`).Split(temp, -1)

        tag := salida[0]
        value := salida[1]

        // Pasar a minusculas
        tag = strings.ToLower(tag)

        if tag == "path" {
            ruta = value
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
	if ruta == "" {
		required = false
	}
	
	if !required {
		fmt.Println("ERROR: La instrucción rmdisk carece de todos los parametros obligatorios.")
	}
	
	//VERIFICAR QUE EL ARCHIVO EXISTA
	_, err := os.Stat(ruta)
	
	if os.IsNotExist(err) {
		fmt.Println("ERROR: El disco que desea eliminar no existe.")
		return
	}
	
	//BORRAR EL DISCO
	err = os.Remove(ruta)
	if err != nil {
		fmt.Println("ERROR: Ocurrió un error al tratar de eliminar el disco.")
	}
	
	fmt.Println("MENSAJE: Archivo eliminado correctamente.")
	
	}