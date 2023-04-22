package programa

import (
	"unicode"
)

func Analizar(cadena *string, parametros *[]string){

	var estado int = 0
	var temp string = ""
	*cadena += " " 
	//RECORRER LA CADENA
	for i := 0; i < len(*cadena); i++{
		switch estado {
	
			case 0:
				if unicode.IsLetter(rune((*cadena)[i])) {
					estado = 1
					temp += string((*cadena)[i])
				}else if (*cadena)[i] == '>' {
					estado = 2
				} else if (*cadena)[i] == '#' {
					estado = 3
				}

			//Palabras reservadas
			case 1: 
				if (*cadena)[i] == ' ' {
					*parametros = append(*parametros, temp)
					temp = ""
					estado = 0
				} else {
					temp += string((*cadena)[i])
				}

			//Parametros
			case 2: 
				if (*cadena)[i] == '"' {
					estado = 21
				} else if (*cadena)[i] == ' ' {
					*parametros = append(*parametros, temp)
					temp = ""
					estado = 0
				} else {
					temp += string((*cadena)[i])
				}

			//Reconocer cadenas dentro de parametros
			case 21: 
				if (*cadena)[i] == '"' {
					*parametros = append(*parametros, temp)
					temp = ""
					estado = 0
				} else {
					temp += string((*cadena)[i])
				}
			
			//Comentarios
			case 3: 
		}
	}
}