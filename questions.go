package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/bbalet/stopwords"
	mux "github.com/gorilla/mux"
	elastic "gopkg.in/olivere/elastic.v6"
)

type Question struct {
	Id     int    `json:"id,omitempty"`
	Que    string `json:"question,omitempty"`
	UserId int    `json:"userid,omitempty"`
	Votes  int    `json:"votes,omitempty"`
	Url    string `json:"Url,omitempty"`
}
type QView struct {
	Total    int
	Question []Question
}
type QRequest struct {
	Que string `json:"query,omitempty"`
}
type Answer struct {
	Id     int    `json:"ans_id"`
	QuesId int    `json:"ques_id"`
	Ans    string `json:"answer"`
	Votes  int    `json:"votes"`
	UserId string `json:"user_id"`
}
type Qkv struct {
	Key   Question
	Value int
}
type Akv struct {
	Key   Question
	Value int
}

const BASEURL = "http://localhost:8000/api/"

func CreateQuestion(q Question) {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}
	id1 := rand.Intn(10000)
	id2 := rand.Intn(10000)
	q.Id = id1 + id2
	fmt.Println(q.Id)
	req := elastic.NewBulkIndexRequest().
		Index("questions").
		Type("_doc").
		Doc(q)
	bResp, bErr := client.Bulk().
		Add(req).
		Do(context.TODO())
	if bErr != nil {
		fmt.Println("Error", bErr)
	}
	fmt.Println(bResp)

}

func GetAllQuestions() string {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}
	var qs []Question
	result, err := client.Search().Index("questions").Type("_doc").Do(context.TODO())
	if err != nil {
		fmt.Println("Error", err)
	}
	for _, hit := range result.Hits.Hits {
		var q Question
		err := json.Unmarshal(*hit.Source, &q)
		if err != nil {

		}
		q.Url = BASEURL + "questions/" + strconv.Itoa(q.Id)
		qs = append(qs, q)
	}
	var res []Question
	qmap := make(map[Question]int)
	for _, val := range qs {
		qmap[val] = val.Votes
	}
	var ss []Qkv
	for k, v := range qmap {
		ss = append(ss, Qkv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for _, kv := range ss {
		res = append(res, kv.Key)
	}
	b, err := json.Marshal(QView{Total: len(res), Question: res})
	if err != nil {
		fmt.Println("Error", err)
	}
	return string(b)
}

func GetQuestion(value string) string {

	client, err := elastic.NewClient()
	if err != nil {

	}
	s := stopwords.CleanString(value, "BCP 47", false)
	values := strings.Fields(s)
	fmt.Println(values)
	var qs []Question
	for _, val := range values {
		wq := elastic.NewWildcardQuery("question", "*"+val+"*")
		//q := elastic.NewTermQuery("question", val)
		result, err := client.Search().Query(wq).Index("questions").Type("_doc").Do(context.TODO())
		if err != nil {

		}
		fmt.Println(len(result.Hits.Hits))
		for _, hit := range result.Hits.Hits {
			var q Question
			err := json.Unmarshal(*hit.Source, &q)
			if err != nil {

			}
			q.Url = BASEURL + "questions/" + strconv.Itoa(q.Id)
			qs = append(qs, q)
		}
	}
	encountered := map[string]bool{}
	result := []Question{}
	for v := range qs {
		if encountered[qs[v].Que] != true {
			encountered[qs[v].Que] = true
			result = append(result, qs[v])
		}
	}
	var res []Question
	qmap := make(map[Question]int)
	for _, val := range qs {
		qmap[val] = val.Votes
	}
	var ss []Qkv
	for k, v := range qmap {
		ss = append(ss, Qkv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	for _, kv := range ss {
		res = append(res, kv.Key)
	}
	b, err := json.Marshal(QView{Total: len(res), Question: res})
	if err != nil {
		fmt.Println("Error", err)
	}
	return string(b)
}

func GetQustionById(id int) []Question {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}
	var qs []Question
	q := elastic.NewTermQuery("Id", id)
	result, err := client.Search().Query(q).Index("questions").Type("_doc").Do(context.TODO())
	if err != nil {
		fmt.Println("Error", err)
	}
	for _, hit := range result.Hits.Hits {
		var q Question
		err := json.Unmarshal(*hit.Source, &q)
		if err != nil {

		}
		q.Url = BASEURL + "questions/" + strconv.Itoa(q.Id)
		qs = append(qs, q)
	}
	var res []Question
	qmap := make(map[Question]int)
	for _, val := range qs {
		qmap[val] = val.Votes
	}
	var ss []Qkv
	for k, v := range qmap {
		ss = append(ss, Qkv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	for _, kv := range ss {
		res = append(res, kv.Key)
	}
	return res
}

func CreateAnswer(value, uid string, qid int) {
	client, err := elastic.NewClient()
	if err != nil {

	}
	q := Answer{Ans: value, UserId: uid, Votes: -1, QuesId: qid}
	req := elastic.NewBulkIndexRequest().
		Index("answers").
		Type("_doc").
		Doc(q)
	bResp, bErr := client.Bulk().
		Add(req).
		Do(context.TODO())
	if bErr != nil {

	}
	fmt.Println(bResp)
}

func GetAnswer(qid int) {
	q := elastic.NewTermQuery("ques_id", qid)
	client, err := elastic.NewClient()
	if err != nil {

	}
	result, err := client.Search().Query(q).Index("answers").Type("_doc").Do(context.TODO())
	if err != nil {

	}
	as := make(map[string]int)
	for _, hit := range result.Hits.Hits {
		var a Answer
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {

		}
		as[a.Ans] = a.Votes
	}
	//var ss []kv
	//for k, v := range as {
	//	ss = append(ss, kv{k, v})
	//}
	//sort.Slice(ss, func(i, j int) bool {
	//	return ss[i].Value > ss[j].Value
	//})
	//
	//for _, kv := range ss {
	//	fmt.Printf("%s %d\n", kv.Key, kv.Value)
	//}
}
func getAllQuestions(w http.ResponseWriter, r *http.Request) {
	q := GetAllQuestions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}
func getQuestionById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	q := GetQustionById(id)
	fmt.Println(q)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)

	fmt.Println(params)
}

func createQuestion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var question Question
	_ = json.NewDecoder(r.Body).Decode(&question)
	fmt.Println(question.Id, question.Que, question.Votes)
	CreateQuestion(question)
	json.NewEncoder(w).Encode(question)
}
func getQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var question QRequest
	_ = json.NewDecoder(r.Body).Decode(&question)
	fmt.Println(question.Que)
	s := GetQuestion(question.Que)
	fmt.Println(s)
	json.NewEncoder(w).Encode(question)
}

func main() {
	//GetAnswer(1)
	r := mux.NewRouter()
	r.HandleFunc("/api/questions", getAllQuestions).Methods("GET")
	r.HandleFunc("/api/questions/{id}", getQuestionById).Methods("GET")
	r.HandleFunc("/api/questions/create", createQuestion).Methods("POST")
	r.HandleFunc("/api/questions/search", getQuestions).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", r))
	//CreateAnswer("no idea ","u1",1)
	//GetQuestion("how are you Ram")
}
