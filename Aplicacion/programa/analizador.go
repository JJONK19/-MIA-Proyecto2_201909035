package programa

import (
	//"fmt"
	"strings"
)

func Ejecutar(cadena *string, sesion *Usuario, discos *[]Disco, salida *[6]string) {
	//Variables
	var parametros []string

	//ANALIZAR LA CADENA
	Analizar(cadena, &parametros)

	//IGNORAR COMENTARIOS
	if len(parametros) == 0 {
		return
	}

	//EJECUTAR INSTRUCCION
	tipo := parametros[0]

	tipo = strings.ToLower(tipo)

	switch tipo {
	case "mkdisk":
		Mkdisk(&parametros, salida)
	case "rmdisk":
		Rmdisk(&parametros, salida)
	case "fdisk":
		Fdisk(&parametros, salida)
	case "mount":
		Mount(&parametros, discos, salida)
	case "mkfs":
		Mkfs(&parametros, discos, salida)
	case "login":
		Login(&parametros, discos, sesion, salida)
	case "logout":
		Logout(sesion, salida)
	case "mkgrp":
		//mkgrp(parametros, discos, sesion)
	case "rmgrp":
		//rmgrp(parametros, discos, sesion)
	case "mkusr":
		//mkusr(parametros, discos, sesion)
	case "rmusr":
		//rmusr(parametros, discos, sesion)
	case "mkfile":
		//mkfile(parametros, discos, sesion)
	case "mkdir":
		//mkdir(parametros, discos, sesion)
	case "rep":
		Rep(&parametros, discos, salida)
	default:
		//fmt.Println("ERROR: El comando ingresado no existe.")
		(*salida)[0] += "ERROR: El comando ingresado no existe."
	}

}
