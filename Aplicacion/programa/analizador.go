package programa

import(
	"strings"
	"fmt"
)

func Ejecutar(cadena *string, sesion *Usuario, discos *[]Disco){
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
			//Mkdisk(parametros)
		case "rmdisk":
			//rmdisk(parametros)
		case "fdisk":
			//fdisk(parametros)
		case "mount":
			//mount(parametros, discos)
		case "unmount":
			//unmount(parametros, discos)
		case "mkfs":
			//mkfs(parametros, discos)
		case "login":
			//login(parametros, discos, sesion)
		case "logout":
			//logout(sesion)
		case "mkgrp":
			//mkgrp(parametros, discos, sesion)
		case "rmgrp":
			//rmgrp(parametros, discos, sesion)
		case "mkusr":
			//mkusr(parametros, discos, sesion)
		case "rmusr":
			//rmusr(parametros, discos, sesion)
		case "chmod":
			//chmod(parametros, discos, sesion)
		case "mkfile":
			//mkfile(parametros, discos, sesion)
		case "cat":
			//cat(parametros, discos, sesion)
		case "remove":
			//remove(parametros, discos, sesion)
		case "edit":
			//edit(parametros, discos, sesion)
		case "rename":
			//rename(parametros, discos, sesion)
		case "mkdir":
			//mkdir(parametros, discos, sesion)
		case "copy":
			//copy(parametros, discos, sesion)
		case "move":
			//move(parametros, discos, sesion)
		case "find":
			//find(parametros, discos, sesion)
		case "chown":
			//chown(parametros, discos, sesion)
		case "chgrp":
			//chgrp(parametros, discos, sesion)
		case "pause":
			//pause()
		case "recovery":
			//recovery(parametros, discos)
		case "loss":
			//loss(parametros, discos)
		case "rep":
			//rep(parametros, discos)
		case "execute":
			//execute(parametros, sesion, discos)
		default:
			fmt.Println("ERROR: El comando ingresado no existe.")
	}

}