package xorm

import (
	"context"
	"fmt"

	"github.com/faisalhardin/medilink/internal/config"
	"github.com/go-xorm/xorm"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"xorm.io/core"
)

type DBConnect struct {
	MasterDB *xorm.Engine
	SlaveDB  *xorm.Engine
}

type dbContext string

// Context keys
var dbContextSession = dbContext("session")

func NewDBConnection(cfg *config.Config) (dbConnection *DBConnect, err error) {

	masterDB, err := generateXormEngineInstance(cfg.Vault.DBMaster.DSN)
	if err != nil {
		return nil, errors.New("failed to make connection to master db")
	}

	slaveDB, err := generateXormEngineInstance(cfg.Vault.DBSlave.DSN)
	if err != nil {
		return nil, errors.New("failed to make connection to slave db")
	}

	return &DBConnect{
		SlaveDB:  slaveDB,
		MasterDB: masterDB,
	}, nil
}

func (conn *DBConnect) CloseDBConnection() error {
	if conn.MasterDB != nil {
		err := conn.MasterDB.Close()
		if err != nil {
			return errors.Wrap(err, "failed to close master db engine")
		}
	}

	if conn.SlaveDB != nil {
		err := conn.SlaveDB.Close()
		if err != nil {
			return errors.Wrap(err, "failed to close slave db engine")
		}
	}
	return nil
}

func generateXormEngineInstance(dsn string) (*xorm.Engine, error) {

	engine, err := xorm.NewEngine("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create engine: %v", err)
	}

	engine.ShowSQL(true)

	// Ping the database to verify the connection
	if err := engine.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	engine.SetTableMapper(core.GonicMapper{})
	engine.SetColumnMapper(core.GonicMapper{})

	return engine, nil

}

// SetDBSession sets existing DB session into context.
func SetDBSession(ctx context.Context, session *xorm.Session) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, dbContextSession, session)
}

// GetDBSession retrieves session from context. Will return nil if session doesn't exist.
func GetDBSession(ctx context.Context) *xorm.Session {
	if ctx == nil {
		return nil
	}
	session, ok := ctx.Value(dbContextSession).(*xorm.Session)
	if !ok {
		return nil
	}
	return session
}

// DBTransactionInterface models a contract for implementing a DB transaction.
type DBTransactionInterface interface {
	Begin(ctx context.Context) (*xorm.Session, error)
	Finish(session *xorm.Session, err *error)
}

// DBTransaction models a wrapper for executing transactions
// on database
type DBTransaction struct {
	*DBConnect
}

// NewTransaction constructs a new DB with transaction
// methods.
func NewTransaction(c *DBConnect) *DBTransaction {
	return &DBTransaction{
		DBConnect: c,
	}
}

// Begin starts a new starting point of a transaction.
func (d *DBTransaction) Begin(ctx context.Context) (*xorm.Session, error) {
	sess := d.MasterDB.NewSession().Context(ctx)
	if err := sess.Begin(); err != nil {
		return nil, errors.Wrap(err, "Begin Transaction")
	}
	return sess, nil
}

func (d *DBTransaction) commit(session *xorm.Session) error {
	if err := session.Commit(); err != nil {
		return errors.Wrap(err, "Commit Transaction")
	}
	return nil
}

func (d *DBTransaction) rollback(session *xorm.Session) error {
	if err := session.Rollback(); err != nil {
		return errors.Wrap(err, "Rollback Transaction")
	}
	return nil
}

// Finish closes the transaction and does a rollback if error exists,
// or commits data if no error found.
func (d *DBTransaction) Finish(session *xorm.Session, err *error) {
	var errOrigin error

	if p := recover(); p != nil {
		_ = d.rollback(session)
		panic(p)
	}
	if err != nil {
		errOrigin = *err
	}
	if errOrigin != nil {
		_ = d.rollback(session)
	} else {
		_ = d.commit(session)
	}
	if session != nil {
		session.Close()
	}
}
