package main

import (
    "log"
    "os"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "github.com/emicklei/go-restful"
    "github.com/emicklei/go-restful-swagger12"
)

// This example is functionally the same as the example in restful-user-resource.go
// with the only difference that is served using the restful.DefaultContainer

type User struct {
    Id string        `json:"id"`
    Name string      `json:"name"`
    Password string  `json:"password"`
}

type UserService struct {
    path string
    // normally one would use DAO (data access object)
    users map[string]User
}

func (u UserService) Register() {
    ws := new(restful.WebService)
    ws.
        Path("/users").
        Consumes(restful.MIME_XML, restful.MIME_JSON).
        Produces(restful.MIME_JSON, restful.MIME_XML) // you can specify this per route as well

    ws.Route(ws.GET("/").To(u.findAllUsers).
        // docs
        Doc("get all users").
        Operation("findAllUsers").
        Writes([]User{}).
        Returns(200, "OK", nil))

    ws.Route(ws.GET("/{user-id}").To(u.findUser).
        // docs
        Doc("get a user").
        Operation("findUser").
        Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
        Writes(User{}). // on the response
        Returns(404, "Not Found", nil))

    ws.Route(ws.PUT("/{user-id}").To(u.updateUser).
        // docs
        Doc("update a user").
        Operation("updateUser").
        Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")).
        Reads(User{})) // from the request

    ws.Route(ws.POST("").To(u.createUser).
        // docs
        Doc("create a user").
        Operation("createUser").
        Reads(User{})) // from the request

    ws.Route(ws.DELETE("/{user-id}").To(u.removeUser).
        // docs
        Doc("delete a user").
        Operation("removeUser").
        Param(ws.PathParameter("user-id", "identifier of the user").DataType("string")))

    restful.Add(ws)
}

// GET http://localhost:8080/users
//
func (u UserService) findAllUsers(request *restful.Request, response *restful.Response) {
    list := []User{}
    for _, each := range u.users {
        list = append(list, each)
    }
    response.WriteEntity(list)
}

// GET http://localhost:8080/users/1
//
func (u UserService) findUser(request *restful.Request, response *restful.Response) {
    id := request.PathParameter("user-id")
    usr := u.users[id]
    if len(usr.Id) == 0 {
        response.WriteErrorString(http.StatusNotFound, "User could not be found.")
    } else {
        response.WriteEntity(usr)
    }
}

// PUT http://localhost:8080/users/1
// <User><Id>1</Id><Name>Melissa Raspberry</Name></User>
//
func (u *UserService) updateUser(request *restful.Request, response *restful.Response) {
    usr := new(User)
    err := request.ReadEntity(&usr)
    if err == nil {
        u.users[usr.Id] = *usr
        err = u.Save()
        if err != nil {
            response.WriteError(http.StatusInternalServerError, err)
        } else {
            response.WriteEntity(usr)
        }
    } else {
        response.WriteError(http.StatusInternalServerError, err)
    }
}

// POST http://localhost:8080/users/
// <User><Id>1</Id><Name>Melissa</Name></User>
//
func (u *UserService) createUser(request *restful.Request, response *restful.Response) {
    usr := User{}
    err := request.ReadEntity(&usr)
    if err == nil {
        u.users[usr.Id] = usr
        err = u.Save()
        if err != nil {
            response.WriteError(http.StatusInternalServerError, err)
        } else {
            response.WriteHeaderAndEntity(http.StatusCreated, usr)
        }
    } else {
        response.WriteError(http.StatusInternalServerError, err)
    }
}

// DELETE http://localhost:8080/users/1
//
func (u *UserService) removeUser(request *restful.Request, response *restful.Response) {
    id := request.PathParameter("user-id")
    delete(u.users, id)
    err := u.Save()
    if err != nil {
        response.WriteError(http.StatusInternalServerError, err)
    }
}

func (u *UserService) Load() error {
    content, err := ioutil.ReadFile(u.path)
    if err != nil {
        return err
    }
    err = json.Unmarshal(content, u.users)
    if err != nil {
        return err
    }
    return nil
}

func (u *UserService) Save() error {
    b, err := json.Marshal(u.users)
    if err != nil {
        return err
    }
    err = ioutil.WriteFile(u.path, b, os.ModePerm)
    if err != nil {
        return err
    }
    return nil
}

func main() {
    path := "./data.json"
    u := UserService{path, map[string]User{}}
    err := u.Load()
    if err != nil {
        log.Printf("fail to load %s", path)
        log.Printf("error: %s", err)
    }
    u.Register()

    // Optionally, you can install the Swagger Service which provides a nice Web UI on your REST API
    // You need to download the Swagger HTML5 assets and change the FilePath location in the config below.
    // Open http://localhost:8080/apidocs and enter http://localhost:8080/apidocs.json in the api input field.
    config := swagger.Config{
        WebServices:    restful.RegisteredWebServices(), // you control what services are visible
        WebServicesUrl: "http://localhost:8080",
        ApiPath:        "/apidocs.json",

        // Optionally, specifiy where the UI is located
        SwaggerPath:     "/apidocs/",
        SwaggerFilePath: "/Users/emicklei/Projects/swagger-ui/dist"}
    swagger.InstallSwaggerService(config)

    log.Printf("start listening on localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
