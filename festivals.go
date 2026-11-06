package main
import(
	"encoding/json"
	"log"
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
