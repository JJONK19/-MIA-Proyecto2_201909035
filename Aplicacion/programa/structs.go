package programa

//********** PARTICION ********** 
type Particion struct {
    part_status [1]byte     // 0 para desactivada, 1 para activa
    part_type   [1]byte     // Partición Primaria o Extendida (p ó e)
    part_fit    [1]byte     // Fit: Best(b), First(f), Worst(w)
    part_start  [40]byte    // Posicion inicial de la partición
    part_s      [40]byte    // Tamaño de la partición (bytes)
    part_name   [30]byte    // Nombre de la partición
}

//********** MBR **********
type MBR struct{
    mbr_tamano [40]byte                         //Tamaño en bytes del disco
    mbr_fecha_creacion [30]byte              	//Fecha y Hora 
    mbr_dsk_signature [40]byte                  //Int random que identifica el disco
    dsk_fit	[1]byte                            	//Fit: Best(b), First(f), Worst(w)
    mbr_partition [4]Particion                  //Particiones en Array (para más facilidad de acceso)
}

//********** EBR **********
type EBR struct{
    part_status [1]byte                       //0 para desactivada, 1 para activa
    part_fit [1]byte                            //Fit: Best(b), First(f), Worst(w)
    part_start [40]byte                         //Posicion inicial de la la partición
    part_s [40]byte                             //Tamaño de la partición(bytes)
    part_next [40]byte                          //Es -1  si no hay otro EBR
    part_name [40]byte                         //Nombre de la partición logica

}

//********** POSICIONES DE LAS PARTICIONES **********
//Indica un resumen de donde estan las particiones para evitar tener que leer constantemente
//el disco. Se usa para encontrar los espacios vacíos.

type Position struct {
    inicio  int32
    fin     int32
    tipo    byte
    nombre  string
    tamaño  int
}

type OrdenarPosicion []Position

func (p OrdenarPosicion) Less(i, j int) bool {
    return p[i].inicio < p[j].inicio
}

//********** ESPACIOS VACIOS **********
//Indica el inicio y fin de los espacios vacíos (particiones primarias y extendida)
type Libre struct{
    inicio int32
    tamaño int32
}

type OrdenarLibre []Libre

func (p OrdenarLibre) Less(i, j int) bool {
    return p[i].tamaño < p[j].tamaño
}

//Indica el espacio vacío dentro de la partición expandida
type LibreL struct{
    inicioEBR int32              //Para leer el EBR de la partición
    finLogica int32              //Donde termina la particion logica
    tamaño int32                 //Espacio Libre
    cabecera bool              //Para diferenciar si es la cabecera de la extendida
}

type OrdenarLibreL []LibreL

func (p OrdenarLibreL) Less(i, j int) bool {
    return p[i].tamaño < p[j].tamaño
}

//********** MONTAR DISCOS **********
//Maneja la posición de la partición dentro del disco
type Montada struct {
    id       string
    posEBR   int32   // default:  -1
    posMBR   int32   // default:  -1
    nombre   string
    tamaño   int
}

//Maneja los datos del disco. Necesario para leer las particiones. 
type Disco struct {
    ruta        string
    nombre      string
    contador    int32    // default: 1
    particiones []Montada
}

//********** SUPER BLOQUE **********

type Sbloque struct {
    s_filesystem_type    [1]byte
    s_inodes_count       [40]byte
    s_blocks_count       [40]byte
    s_free_blocks_count  [40]byte
    s_free_inodes_count  [40]byte
    s_mtime              [30]byte
    s_umtime             [30]byte
    s_mnt_count          [40]byte
    s_magic              [10]byte
    s_inode_s            [40]byte
    s_block_s            [40]byte
    s_firts_ino          [40]byte
    s_first_blo          [40]byte
    s_bm_inode_start     [40]byte
    s_bm_block_start     [40]byte
    s_inode_start        [40]byte
    s_block_start        [40]byte
}


//********** INODOS **********
type Inodo struct {
    i_uid       [10]byte
    i_gid       [10]byte
    i_s         [40]byte
    i_atime     [30]byte
    i_ctime     [30]byte
    i_mtime     [30]byte
    i_block     [16]byte
    i_type      [1]byte 
    i_perm      [3]byte
}

//********** BLOQUES **********
type Content struct {
    b_name [12]byte
    b_inodo [4]byte
}

type Bcarpetas struct {
    b_content [4]Content
}


type Barchivos struct {
    b_content [64]byte
}

type Bapuntadores struct {
    b_pointers [16]int
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


