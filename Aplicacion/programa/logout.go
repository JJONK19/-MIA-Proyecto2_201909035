package programa

import (
	"fmt"
)

func Logout(sesion *Usuario) {
	// VERIFICAR QUE NO EXISTA UNA SESIÓN
	if sesion.user == "" {
		fmt.Println("ERROR: No hay una sesión iniciada.")
		return
	}

	// CERRAR LA SESION
	sesion.user = ""
	sesion.pass = ""
	sesion.disco = ""
	sesion.grupo = ""
	fmt.Println("MENSAJE: Sesión finalizada.")
}
