package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
	"strings"
)
type DeleteFestivalRequest struct{
	ID int `json:"id"`
}
func loadFestivals()([]Festival,error){
	rows,err:=db.Query("SELECT id,name,start_date,end_date,effects FROM festivals ORDER BY start_date,name")
	if err!=nil{
		return nil,err
	}
	defer rows.Close()

	festivals:=[]Festival{}
	for rows.Next(){
		var f Festival
		var effects string
		err:=rows.Scan(&f.ID,&f.Name,&f.Start,&f.End,&effects)
		if err!=nil{
			return nil,err
		}
		err=json.Unmarshal([]byte(effects),&f.Effects)
		if err!=nil{
			return nil,err
		}
		festivals=append(festivals,f)
	}
	return festivals,rows.Err()
}


func insertFestival(f Festival)(int,error){
	effects,err:=json.Marshal(f.Effects)
	if err!=nil{
		return 0,err
	}
	result,err:=db.Exec("INSERT INTO festivals (name,start_date,end_date,effects) VALUES (?,?,?,?)",
  						f.Name, f.Start, f.End, string(effects))
	
	if err!=nil{
		return 0,err
	}
	id,err:=result.LastInsertId()
	return int(id),err
}

func insertDefaultFestivals(){
	var count int
	err:=db.QueryRow("SELECT COUNT(*) FROM festivals").Scan(&count)
	if(err!=nil){
		log.Fatal(err)
	}
	if(count>0){
		return
	}
	for _,f:=range defaultFestivals{
		_,err=insertFestival(f)
		if(err!=nil){
			log.Fatal(err)
		}
	}
}
func readFestival(w http.ResponseWriter, r *http.Request)(Festival,bool){
	var f Festival
	if !readJSON(w,r,&f){
		return f,false
	}
	f.Name=strings.TrimSpace(f.Name)
	if f.Name==""{
		http.Error(w,"Name is required",http.StatusBadRequest)
		return f,false
	}
	_,startErr:=time.Parse("2006-01-02",f.Start)
	_,endErr:=time.Parse("2006-01-02",f.End)
	if startErr!=nil || endErr!=nil{
		http.Error(w,"Start and end must look like 2006-10-14",http.StatusBadRequest)
		return f,false
	}
	if(f.End<f.Start){
		http.Error(w,"End date can not be before start date",http.StatusBadRequest)
		return f,false
	}
	if(len(f.Effects)==0){
		http.Error(w,"Add at least one category with its multiplier",http.StatusBadRequest)
		return f,false
	}

	effects:=map[string]float64{}
	for category,multiplier:=range f.Effects{
		category=strings.TrimSpace(category)
		if(category==""){
			http.Error(w, "Category can not be empty", http.StatusBadRequest)
			return f,false
		}
		if(multiplier<=0 || multiplier>10){
			http.Error(w, "Multiplier must be between 0 and 10", http.StatusBadRequest)
			return f,false
		}
		_,found:=effects[category]
		if(found){
			http.Error(w,category+" is added twice",http.StatusBadRequest)
			return f,false
		}
		effects[category]=multiplier
	}
	f.Effects=effects
	return f,true
}
func listFestivals(w http.ResponseWriter,r *http.Request){
	festivals,err:=loadFestivals()
	if(err!=nil){
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	writeJSON(w,festivals)
}
func addFestival(w http.ResponseWriter,r *http.Request){
	f,ok:=readFestival(w,r)
	if(ok==false){
		return
	}
	id,err:=insertFestival(f)
	if(err!=nil){
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	f.ID=id
	writeJSON(w,f)
}
func updateFestival(w http.ResponseWriter,r *http.Request){
	f,ok:=readFestival(w,r)
	if(ok==false){
		return
	}
	//map---->>>>text conversion using Marshal
	effects,err:=json.Marshal(f.Effects)
	if(err!=nil){
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	result,err:=db.Exec("UPDATE festivals SET name = ?,start_date=?,end_date=?,effects=? WHERE id=?",
				f.Name,f.Start,f.End,string(effects),f.ID)
	if(err!=nil){
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	count,err:=result.RowsAffected()
	if(err!=nil){
		http.Error(w,err.Error(),http.StatusInternalServerError)
		return
	}
	if(count==0){
		http.Error(w,"Festival not found",http.StatusNotFound)
		return
	}
	writeJSON(w,map[string]string{"message":"Festival updated"})

}