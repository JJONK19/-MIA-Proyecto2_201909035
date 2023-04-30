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
		Mkgrp(&parametros, discos, sesion, salida)
	case "rmgrp":
		Rmgrp(&parametros, discos, sesion, salida)
	case "mkusr":
		Mkusr(&parametros, discos, sesion, salida)
	case "rmusr":
		Rmusr(&parametros, discos, sesion, salida)
	case "mkfile":
		Mkfile(&parametros, discos, sesion, salida)
	case "mkdir":
		Mkdir(&parametros, discos, sesion, salida)
	case "rep":
		Rep(&parametros, discos, salida)
	default:
		//fmt.Println("ERROR: El comando ingresado no existe.")
		(*salida)[0] += "ERROR: El comando ingresado no existe."
	}

}
