package main
import(
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
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

	festivals:=[]Festivals{}
	for rows.Next(){
		var f Festival
		var effects string
		err:=rows.Scan(&f.ID,&f,Name,&f.Start,&f.End,&effects)
		if err!=nil{
			
		}
	}
}