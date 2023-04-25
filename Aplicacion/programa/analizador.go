package programa

import (
	"fmt"
	"strings"
)

func Ejecutar(cadena *string, sesion *Usuario, discos *[]Disco) {
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
		Mkdisk(&parametros)
	case "rmdisk":
		Rmdisk(&parametros)
	case "fdisk":
		Fdisk(&parametros)
	case "mount":
		Mount(&parametros, discos)
	case "mkfs":
		Mkfs(&parametros, discos)
	case "login":
		Login(&parametros, discos, sesion)
	case "logout":
		Logout(sesion)
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
	case "pause":
		//pause()
	case "rep":
		Rep(&parametros, discos)
	default:
		fmt.Println("ERROR: El comando ingresado no existe.")
	}

}
