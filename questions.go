package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
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
	UserId string `json:"userid,omitempty"`
	Votes  int    `json:"votes,omitempty"`
	Url    string `json:"Url,omitempty"`
}
type QView struct {
	Total    int
	Question []Question
}
type QAView struct {
	Qid    int
	Ques   string
	Answer []Answer
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
	Url    string `json:"Url,omitempty"`
}
type Qkv struct {
	Key   Question
	Value int
}
type Akv struct {
	Key   Answer
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
		Doc(q).Id(strconv.Itoa(q.Id))
	bResp, bErr := client.Bulk().
		Add(req).
		Do(context.TODO())
	if bErr != nil {
		fmt.Println("Error", bErr)
	}
	fmt.Println(bResp)

}

func GetAllQuestions() QView {
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
	return QView{Total: len(res), Question: res}
}

func GetQuestion(value string) QView {
	client, err := elastic.NewClient()
	if err != nil {

	}
	s := stopwords.CleanString(value, "BCP 47", false)
	values := strings.Fields(s)
	fmt.Println(values)
	var qs []Question
	for _, val := range values {
		wq := elastic.NewWildcardQuery("question", "*"+val+"*")
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
	return QView{Total: len(res), Question: res}
}

func GetQustionById(id int) *QAView {
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
	ans := GetAnswerByQid(id)
	fmt.Println(len(res))
	if len(res) == 0 {
		return nil
	}
	return &QAView{Qid: res[0].Id, Ques: res[0].Que, Answer: ans}
}

func CreateAnswer(a Answer) {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}
	id1 := rand.Intn(10000)
	id2 := rand.Intn(10000)
	a.Id = id1 + id2 + a.QuesId
	req := elastic.NewBulkIndexRequest().
		Index("answers").
		Type("_doc").
		Doc(a).Id(strconv.Itoa(a.Id))
	bResp, bErr := client.Bulk().
		Add(req).
		Do(context.TODO())
	if bErr != nil {
		fmt.Println("Error", err)
	}
	fmt.Println(bResp)
}
func GetAnswerByAID(aid int) []Answer {
	q := elastic.NewTermQuery("ans_id", aid)
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}
	result, err := client.Search().Query(q).Index("answers").Type("_doc").Do(context.TODO())
	if err != nil {
		fmt.Println("Error", err)
	}
	var as []Answer
	for _, hit := range result.Hits.Hits {
		var a Answer
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			fmt.Println("Error", err)
		}
		a.Url = BASEURL + "answers/" + strconv.Itoa(a.Id)
		as = append(as, a)
	}
	var res []Answer
	amap := make(map[Answer]int)
	for _, val := range as {
		amap[val] = val.Votes
	}
	var ss []Akv
	for k, v := range amap {
		ss = append(ss, Akv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	for _, kv := range ss {
		res = append(res, kv.Key)
	}
	return res
}
func GetAnswerByQid(qid int) []Answer {
	q := elastic.NewTermQuery("ques_id", qid)
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}
	result, err := client.Search().Query(q).Index("answers").Type("_doc").Do(context.TODO())
	if err != nil {
		fmt.Println("Error", err)
	}
	var as []Answer
	for _, hit := range result.Hits.Hits {
		var a Answer
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			fmt.Println("Error", err)
		}
		a.Url = BASEURL + "answers/" + strconv.Itoa(a.Id)
		as = append(as, a)
	}
	var res []Answer
	amap := make(map[Answer]int)
	for _, val := range as {
		amap[val] = val.Votes
	}
	var ss []Akv
	for k, v := range amap {
		ss = append(ss, Akv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	for _, kv := range ss {
		res = append(res, kv.Key)
	}
	return res
}

func IncrementQuestionVotes(qid int) {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}

	query := elastic.NewTermQuery("id", qid)
	result, err := client.Search().Query(query).Index("questions").Type("_doc").Do(context.TODO())
	fmt.Println(len(result.Hits.Hits), err)
	var qs []Question
	for _, hit := range result.Hits.Hits {
		var q Question
		err := json.Unmarshal(*hit.Source, &q)
		if err != nil {

		}
		qs = append(qs, q)
	}
	qs[0].Votes += 1
	fmt.Println(qs)
	client.Update().Index("questions").Type("_doc").Id(strconv.Itoa(qid)).Doc(qs[0]).Do(context.TODO())
}
func createQuestion(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var question Question
	// enableCors(&w)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewDecoder(r.Body).Decode(&question)
	CreateQuestion(question)
	json.NewEncoder(w).Encode(question)
}

func getAllQuestions(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	q := GetAllQuestions()
	// enableCors(&w)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(q)
}
func getQuestionById(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	params := mux.Vars(r)
	// enableCors(&w)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	id, _ := strconv.Atoi(params["id"])
	q := GetQustionById(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}

func getQuestions(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	w.Header().Set("Content-Type", "application/json")
	// enableCors(&w)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var question QRequest
	_ = json.NewDecoder(r.Body).Decode(&question)
	s := GetQuestion(question.Que)
	json.NewEncoder(w).Encode(s)
}

func createAnswer(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	w.Header().Set("Content-Type", "application/json")
	// enableCors(&w)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var answer Answer
	_ = json.NewDecoder(r.Body).Decode(&answer)
	CreateAnswer(answer)
	json.NewEncoder(w).Encode(answer)
}

func getAnswerByAID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	// enableCors(&w)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	id, _ := strconv.Atoi(params["id"])
	fmt.Println(id)
	a := GetAnswerByAID(id)
	fmt.Println(a)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/questions", getAllQuestions).Methods("GET")
	r.HandleFunc("/api/questions/{id}", getQuestionById).Methods("GET")
	r.HandleFunc("/api/questions/create", createQuestion).Methods("POST")
	r.HandleFunc("/api/questions/search", getQuestions).Methods("POST")
	r.HandleFunc("/api/answers/create", createAnswer).Methods("POST")
	r.HandleFunc("/api/answers/{id}", getAnswerByAID).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", r))
	//CreateQuestion(Question{Que: "reportix in go lang?", Votes: 5})
	//UpdateQuestionVotes(-7, 15968)
}
func unmarshalJSON(bytes []byte, docType reflect.Type) (interface{}, error) {
	obj := reflect.New(docType)
	doc := obj.Interface()
	if err := json.Unmarshal(bytes, doc); err != nil {
		return nil, fmt.Errorf("error while unmarshalling json document to struct of type %v: %v", docType, err)
	}
	return doc, nil
}
