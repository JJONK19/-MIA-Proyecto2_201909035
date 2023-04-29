package programa

import (
	//"fmt"
	//"bufio"
	//"os"
    "strings"
)

func Consola(sesion *Usuario, discos *[]Disco, salida *[6]string, comandos *string){
	//Para ejecutar en consola
    /*
    //VARIABLES
	var continuar bool = true

	//CONSOLA DE COMANDOS
    fmt.Println("****************************************************************************************************")
    fmt.Println()
    fmt.Println("PROYECTO 2 ARCHIVOS - 201909035") 
    fmt.Println()
    fmt.Println("****************************************************************************************************")

	for continuar{
		fmt.Println("----------------------------------------------------------------------------------------------------")
        fmt.Println("INSTRUCCION:")  
        lector := bufio.NewReader(os.Stdin)
		comando, fallo := lector.ReadString('\n')
		if fallo != nil {
			fmt.Println("ERROR: Hubo un problema al leer la entrada.")
			return
		}
        comando = strings.TrimRight(comando, "\n")

        //Salir de la aplicación
        if comando == "EXIT" || comando == "exit"{
            continuar = false
            fmt.Println("----------------------------------------------------------------------------------------------------")          
            continue
        }

        //Ejecutar Instruccion
        fmt.Println("EJECUCIÓN:")
        Ejecutar(&comando, sesion, discos);
        comando = ""  
        fmt.Println()
	}
    */

    //Para ejecutar como API
    (*salida)[0] += "****************************************************************************************************\n"
    (*salida)[0] += "\n"
    (*salida)[0] += "PROYECTO 2 ARCHIVOS - 201909035\n"
    (*salida)[0] += "\n"
    (*salida)[0] += "****************************************************************************************************\n"
    (*salida)[0] += "INSTRUCCION:\n"

    lineas := strings.Split(*comandos, "\n")
    for i := 0; i < len(lineas); i++ {
        (*salida)[0] += lineas[i] + "\n"
        Ejecutar(&lineas[i], sesion, discos, salida)
        (*salida)[0] += "\n"
    }
}