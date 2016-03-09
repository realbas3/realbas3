package db

import (
	"github.com/neutrinoapp/neutrino/src/common/config"
	"github.com/neutrinoapp/neutrino/src/common/log"
	"github.com/neutrinoapp/neutrino/src/common/models"

	"strings"

	r "github.com/dancannon/gorethink"
	"github.com/neutrinoapp/neutrino/src/common/utils"
)

var session *r.Session

type DbService interface {
	GetSession() *r.Session
	GetDb() r.Term
	GetTable() r.Term
	Insert(doc map[string]interface{}) error
	Update(q, u map[string]interface{}) error
	FindId(id interface{}) (map[string]interface{}, error)
	Find(query interface{}) ([]map[string]interface{}, error)
	FindOne(query interface{}) (map[string]interface{}, error)
	RemoveId(id interface{}) error
	UpdateId(id, u interface{}) error
	ReplaceId(id, u interface{}) error
	Changes(filter, channel interface{}) error
}

type dbService struct {
	address, dbName, tableName string
}

func NewDbService(dbName, tableName string) DbService {
	address := config.Get(config.KEY_RETHINK_ADDR)
	tableName = strings.Replace(strings.Replace(tableName, ".", "_", -1), "@", "_", -1)
	d := dbService{address, dbName, tableName}
	return &d
}

func NewAppsMapDbService() DbService {
	return NewDbService(Constants.DatabaseName(), Constants.AppsMapCollection())
}

func NewUsersDbService() DbService {
	return NewDbService(Constants.DatabaseName(), Constants.UsersCollection())
}

func NewTypeDbService(appId, typeName string) DbService {
	return NewDbService(Constants.DatabaseName(), appId+"."+typeName)
}

func NewAppsDbService(user string) DbService {
	return NewDbService(Constants.DatabaseName(), user+"."+Constants.ApplicationsCollection())
}

func NewAppUsersDbService(appId string) DbService {
	return NewDbService(Constants.DatabaseName(), appId+"."+"users")
}

func NewSystemDbService() DbService {
	return NewDbService(Constants.DatabaseName(), Constants.SystemCollection())
}

func (d *dbService) GetSession() *r.Session {
	if session == nil {
		s, err := r.Connect(r.ConnectOpts{
			Address: d.address,
		})

		if err != nil {
			log.Error(err)
			panic(err)
		}

		session = s
	}

	return session
}

func (d *dbService) GetDb() r.Term {
	return r.DB(d.dbName)
}

func (d *dbService) GetTable() r.Term {
	db := d.GetDb()

	_, err := db.TableList().Contains(d.tableName).Do(func(tableExists r.Term) r.Term {
		return r.Branch(
			tableExists,
			models.JSON{"created": 0},
			db.TableCreate(d.tableName),
		)
	}).Run(d.GetSession())

	if err != nil {
		log.Error(err)
	}

	return db.Table(d.tableName)
}

func (d *dbService) Insert(doc map[string]interface{}) error {
	if doc["id"] == nil {
		doc["id"] = utils.GetCleanUUID()
	}

	_, err := d.GetTable().Insert(doc).RunWrite(d.GetSession())
	return err
}

func (d *dbService) Update(q, u map[string]interface{}) error {
	_, err := d.GetTable().Filter(q).Update(u).RunWrite(d.GetSession())

	return err
}

func (d *dbService) FindId(id interface{}) (map[string]interface{}, error) {
	c, err := d.GetTable().Get(id).Run(d.GetSession())
	defer c.Close()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = c.One(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *dbService) Find(query interface{}) ([]map[string]interface{}, error) {
	c, err := d.GetTable().Filter(query).Run(d.GetSession())
	defer c.Close()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	err = c.All(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (d *dbService) FindOne(query interface{}) (map[string]interface{}, error) {
	c, err := d.GetTable().Filter(query).Limit(1).Run(d.GetSession())
	defer c.Close()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = c.All(&result)
	if err != nil {
		return nil, err
	}

	return result, err
}

func (d *dbService) RemoveId(id interface{}) error {
	_, err := d.GetTable().Get(id).Delete().RunWrite(d.GetSession())
	return err
}

func (d *dbService) UpdateId(id, u interface{}) error {
	_, err := d.GetTable().Get(id).Update(u).RunWrite(d.GetSession())
	return err
}

func (d *dbService) ReplaceId(id, u interface{}) error {
	_, err := d.GetTable().Get(id).Replace(u).RunWrite(d.GetSession())
	return err
}

func (d *dbService) Changes(filter, channel interface{}) error {
	c, err := d.GetTable().Filter(filter).Changes().Run(d.GetSession())
	if err != nil {
		return err
	}

	c.Listen(channel)
	return nil
}