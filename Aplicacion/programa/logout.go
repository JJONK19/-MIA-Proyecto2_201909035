package programa

import (
	//"fmt"
)

func Logout(sesion *Usuario, salidas *[6]string) {
	// VERIFICAR QUE NO EXISTA UNA SESIÓN
	if sesion.user == "" {
		(*salidas)[0] += "ERROR: No hay una sesión iniciada.\n"
		//fmt.Println("ERROR: No hay una sesión iniciada.")
		return
	}

	// CERRAR LA SESION
	sesion.user = ""
	sesion.pass = ""
	sesion.disco = ""
	sesion.grupo = ""
	(*salidas)[0] += "MENSAJE: Sesión finalizada.\n"
	(*salidas)[1] = "0"
	//fmt.Println("MENSAJE: Sesión finalizada.")
}
