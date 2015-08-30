package api

import (
	"github.com/ant0ine/go-json-rest/rest"
	"gopkg.in/mgo.v2/bson"
	"errors"
	"time"
	"github.com/go-realbase/realbase/core"
	"github.com/go-realbase/realbase/utils"
	"net/http"
)

type ApplicationModel struct {
	Id string `json: _id`
	Name string `json: "name"`
}

type ApplicationController struct {
}

func (a *ApplicationController) Path() string {
	return "/applications"
}

func GetAppFromRequest(r *rest.Request) (map[string]interface{}, error) {
	appId := r.PathParam("appId")

	if appId != "" {
		//TODO: cache this
		appDb := realbase.NewApplicationsDbService(r.Env["user"].(string))
		return appDb.FindId(appId, nil)
	} else {
		return nil, errors.New("Invalid app id.")
	}
}

func (a *ApplicationController) CreateApplicationHandler(w rest.ResponseWriter, r *rest.Request) {
	body := &ApplicationModel{}

	if err := r.DecodeJsonPayload(body); err != nil {
		RestGeneralError(w, err)
		return
	}

	if body.Name == "" {
		RestErrorInvalidBody(w)
		return
	}

	db := realbase.NewApplicationsDbService(r.Env["user"].(string))

	username := r.Env["user"].(string)
	doc := bson.M{
		"name": body.Name,
		"owner": username,
		"types": []string{"users"},
		"createdAt": time.Now(),
		"keys": bson.M{ //TODO:
			"Master Key": bson.M{
				"key": utils.GetCleanUUID(),
				"name": "Master Key",
				"permissions": bson.M{
					"types": bson.M{
						"read": true,
						"write": true,
					},
				},
			},
		},
	}

	if err := db.Insert(doc); err != nil {
		RestGeneralError(w, err)
		return
	}

	RespondId(doc["_id"], w)
}

func (a *ApplicationController) GetApplicationsHandler(w rest.ResponseWriter, r *rest.Request) {
	db := realbase.NewApplicationsDbService(r.Env["user"].(string))

	res, err := db.Find(
		bson.M{
			"owner": r.Env["user"],
		},
		bson.M{
			"name": 1,
		},
	)

	if err != nil {
		RestGeneralError(w, err)
		return
	}

	w.WriteJson(res)
}

func (a *ApplicationController) GetApplicationHandler(w rest.ResponseWriter, r *rest.Request) {
	app, err := GetAppFromRequest(r)

	if err != nil {
		RestGeneralError(w, err)
		return
	}

	w.WriteJson(app)
}

func (a *ApplicationController) DeleteApplicationHandler(w rest.ResponseWriter, r *rest.Request) {
	appId := r.PathParam("appId")

	db := realbase.NewApplicationsDbService(r.Env["user"].(string))
	err := db.RemoveId(appId)

	if err != nil {
		RestGeneralError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *ApplicationController) UpdateApplicationHandler(w rest.ResponseWriter, r *rest.Request) {
	appId := r.PathParam("appId")
	db := realbase.NewApplicationsDbService(r.Env["user"].(string))
	doc := utils.WhitelistFields([]string{"name"}, utils.GetBody(r))

	err := db.Update(bson.M{
		"_id": appId,
	}, bson.M{
		"$set": doc,
	})

	if err != nil {
		RestGeneralError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}