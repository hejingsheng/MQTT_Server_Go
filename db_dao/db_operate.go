package db_dao

type DBOperate interface {

	OpenDatabase() (conn interface{}, err error)

	CloseDatabase(conn interface{}) int

}
