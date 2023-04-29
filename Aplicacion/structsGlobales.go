package main

type EntradaConsola struct {
	Comandos string `json:"comandos"`
}

type EntradaLogin struct {
	Particion string `json:"particion"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type EntradaFile struct {
	Ruta      string `json:"ruta"`
	Particion string `json:"particion"`
}

type Salida struct {
	Consola string //Contenido a mostrar en la consola
	Login   string //0 si es invalido, 1 si es valido
	File    string //Contenido del reporte file
	Tree    string //Dot del tree
	Sb      string //Dot del superbloque
	Disk    string //Dot del disco
}
