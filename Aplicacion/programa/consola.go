package programa

import (
	"fmt"
	"bufio"
	"os"
    "strings"
)

func Consola(){
	//VARIABLES
	var continuar bool = true
    var sesion Usuario
    var discos []Disco

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
        Ejecutar(&comando, &sesion, &discos);
        comando = ""  
        fmt.Println()
	}
}