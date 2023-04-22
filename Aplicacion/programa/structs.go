package programa

import (
	"strconv"
    "strings"
)

//********** PARTICION **********
type Particion struct {
    Part_status [1]byte     // 0 para desactivada, 1 para activa
    Part_type   [1]byte     // Partición Primaria o Extendida (p ó e)
    Part_fit    [1]byte     // Fit: Best(b), First(f), Worst(w)
    Part_start  [40]byte    // Posicion inicial de la partición
    Part_s      [40]byte    // Tamaño de la partición (bytes)
    Part_name   [30]byte    // Nombre de la partición
}

//********** MBR **********
type MBR struct{
    Mbr_tamano [40]byte                         //Tamaño en bytes del disco
    Mbr_fecha_creacion [30]byte              	//Fecha y Hora 
    Mbr_dsk_signature [40]byte                  //Int random que identifica el disco
    Dsk_fit	[1]byte                            	//Fit: Best(b), First(f), Worst(w)
    Mbr_partition [4]Particion                  //Particiones en Array (para más facilidad de acceso)
}

//********** EBR **********
type EBR struct{
    Part_status [1]byte                       //0 para desactivada, 1 para activa
    Part_fit [1]byte                            //Fit: Best(b), First(f), Worst(w)
    Part_start [40]byte                         //Posicion inicial de la la partición
    Part_s [40]byte                             //Tamaño de la partición(bytes)
    Part_next [40]byte                          //Es -1  si no hay otro EBR
    Part_name [40]byte                         //Nombre de la partición logica

}

//********** POSICIONES DE LAS PARTICIONES **********
//Indica un resumen de donde estan las particiones para evitar tener que leer constantemente
//el disco. Se usa para encontrar los espacios vacíos.

type Position struct {
    inicio  int
    fin     int
    tipo    byte
    nombre  string
    tamaño  int
}

type OrdenarPosicion []Position

func (p OrdenarPosicion) Less(i, j int) bool {
    return p[i].inicio < p[j].inicio
}

func (o OrdenarPosicion) Len() int {
    return len(o)
}

func (o OrdenarPosicion) Swap(i, j int) {
    o[i], o[j] = o[j], o[i]
}

//********** ESPACIOS VACIOS **********
//Indica el inicio y fin de los espacios vacíos (particiones primarias y extendida)
type Libre struct{
    inicio int
    tamaño int
}

type OrdenarLibre []Libre

func (p OrdenarLibre) Less(i, j int) bool {
    return p[i].tamaño < p[j].tamaño
}

func (p OrdenarLibre) Len() int {
    return len(p)
}

func (p OrdenarLibre) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}


//Indica el espacio vacío dentro de la partición expandida
type LibreL struct{
    inicioEBR int              //Para leer el EBR de la partición
    finLogica int              //Donde termina la particion logica
    tamaño int                 //Espacio Libre
    cabecera bool              //Para diferenciar si es la cabecera de la extendida
}

type OrdenarLibreL []LibreL

func (p OrdenarLibreL) Less(i, j int) bool {
    return p[i].tamaño < p[j].tamaño
}

func (p OrdenarLibreL) Len() int {
    return len(p)
}

func (p OrdenarLibreL) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}


//********** MONTAR DISCOS **********
//Maneja la posición de la partición dentro del disco
type Montada struct {
    id       string
    posEBR   int   // default:  -1
    posMBR   int   // default:  -1
    nombre   string
    tamaño   int
}

//Maneja los datos del disco. Necesario para leer las particiones. 
type Disco struct {
    ruta        string
    nombre      string
    contador    int    // default: 1
    particiones []Montada
}

//********** SUPER BLOQUE **********

type Sbloque struct {
    S_filesystem_type    [1]byte
    S_inodes_count       [40]byte
    S_blocks_count       [40]byte
    S_free_blocks_count  [40]byte
    S_free_inodes_count  [40]byte
    S_mtime              [30]byte
    S_umtime             [30]byte
    S_mnt_count          [40]byte
    S_magic              [10]byte
    S_inode_s            [40]byte
    S_block_s            [40]byte
    S_firts_ino          [40]byte
    S_first_blo          [40]byte
    S_bm_inode_start     [40]byte
    S_bm_block_start     [40]byte
    S_inode_start        [40]byte
    S_block_start        [40]byte
}


//********** INODOS **********
type Inodo struct {
    I_uid       [10]byte
    I_gid       [10]byte
    I_s         [40]byte
    I_atime     [30]byte
    I_ctime     [30]byte
    I_mtime     [30]byte
    I_block     [16]byte
    I_type      [1]byte 
    I_perm      [3]byte
}

//********** BLOQUES **********
type Content struct {
    B_name [12]byte
    B_inodo [4]byte
}

type Bcarpetas struct {
    B_content [4]Content
}


type Barchivos struct {
    B_content [64]byte
}

type Bapuntadores struct {
    B_pointers [16]int
}

//********** LOGIN **********
type Usuario struct {
    user     string // ID del usuario
    pass     string // Contraseña del usuario
    disco    string // ID de la particion en la que esta trabajando
    grupo    string // Grupo al que pertenece el usuario
    id_user  string // Numero de usuario
    id_grp   string // Numero de grupo
}

//*********MANEJO DE BYTES EN EL ARCHIVO*********

//Pasar bytes a int. Se estan manejando como cadenas, entonces se debe de castear de bytes a cadenas y luego a int
func ToInt(numero []byte) int {
	str := strings.Trim(string(numero[:]), "\x00") 
	salida, err := strconv.Atoi(str) 
	if err != nil {
	
	}

	return salida
}

