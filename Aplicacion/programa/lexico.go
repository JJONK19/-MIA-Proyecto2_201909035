package programa

import(
	"unicode"
)

func Analizar(cadena *string, parametros *[]string){

	var estado int = 0
	var temp string = ""

	//RECORRER LA CADENA
	for _, letra := range *cadena{
		switch estado {
			case 0:
				if unicode.IsLetter(letra) {
					estado = 1
					temp += string(letra)
				}else if letra == '>' {
					estado = 2
				} else if letra == '#' {
				estado = 3
				}

			//Palabras reservadas
			case 1: 
				if letra == ' ' {
					*parametros = append(*parametros, temp)
					temp = ""
					estado = 0
				} else {
					temp += string(letra)
				}

			//Parametros
			case 2: 
				if letra == '"' {
					estado = 21
				} else if letra == ' ' {
					*parametros = append(*parametros, temp)
					temp = ""
					estado = 0
				} else {
					temp += string(letra)
				}

			//Reconocer cadenas dentro de parametros
			case 21: 
				if letra == '"' {
					*parametros = append(*parametros, temp)
					temp = ""
					estado = 0
				} else {
					temp += string(letra)
				}
			
			//Comentarios
			case 3: 
			
		}
	}
}