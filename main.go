package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostName string ="localhost:27017"
	dbName string = "demo_todo"
	collectionName string = "todo"
	port string    		  = ":9000"
)

type (
	todoModel struct{
		ID bson.ObjectId `bson:"_id,omitempty"`
		Title string `bson:"title"`
		Completed bool `bson: "Completed"`
		CreatedAt time.Time `bson:"CreateAt"`
	}

	todo struct{
		ID string `json:"id"`
		Title string `json:title`
		Completed string `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func init(){
	rnd = renderer.New()
	sess,err := mgo.Dial(hostName)
	checkErr(err)
	sess.SetMode(mgo.Monotonic,true)
	db = sess.DB(dbName)
}

func homeHandler(w http.ResponseWriter, r *http.Request){
	err:= rnd.Template(w, http.StatusOK, []string{"static/home.tpl"},nil)
	checkErr()
}
func fetchTodos(w http.ResponseWriter, r *http.Request){
	todos:=[]todoModel{}
		if err := db.C(collectionName).Find(bson.M{}).All(&todos); err !=nil{
			rnd.JSON(w, http.StatusProcessing, renderer.M{
				"message":"Failed to fetch todo",
				"error":err,
			})
			return
		}
	todoList:= []todo{}

	for _,t :=range todos{
		todoList=append(todoList, todo{
			ID:t.ID.Hex(),
			Title: t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data":todoList,
	})
}

func main(){
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	r:= chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/",homeHandler)
	r.Mount("/todo",todoHandlers())

	srv := &http.Server{
		Addr: port,
		Handler: r,
		ReadTimeout: 60*time.Second,
		WriteTimeout: 60*time.Second,
		IdleTimeout: 60*time.Second,
	}
	go func ()  {
		log.Println("Listening on port",port)
		if err:=srv.ListenAndServe(); err != nil{
			log.Printf("Listen:%s\n",err)
		}
	}()
	<-stopChan
	log.Println("Shutting down server....")
	ctx,cancel := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	defer cancel(
		log.Println("server gracefully stopped")
	)
}

func todoHandlers() http.Handler{
	rg:= chi.NewRouter()
	rg.Group(func(r chi.Router){
		r.Get("/",fetchTodos)
		r.Post("/",createTodo)
		r.Put("/{id}",updateTodo)
		r.Delete("/{id}",deleteTodo)
	})
}

func checkErr(err error){
	if err!=nil{
		log.Fatal(err)
	}
}

