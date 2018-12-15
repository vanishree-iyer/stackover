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
	"time"

	"github.com/bbalet/stopwords"
	mux "github.com/gorilla/mux"
	elastic "gopkg.in/olivere/elastic.v6"
)

type Question struct {
	Id        int       `json:"id,omitempty"`
	Title     string    `json:"title,omitempty"`
	Body      string    `json:"body,omitempty"`
	UserId    string    `json:"userid,omitempty"`
	Votes     int       `json:"votes,omitempty"`
	TimeStamp time.Time `json:timestamp,omitempty`
	Url       string    `json:"Url,omitempty"`
}
type QView struct {
	Total    int
	Question []Question
}
type QAView struct {
	Qid    int
	Title  string
	Body   string
	Answer []Answer
}
type QRequest struct {
	Que string `json:"query,omitempty"`
}
type Answer struct {
	Id        int       `json:"ans_id"`
	QuesId    int       `json:"ques_id"`
	Body      string    `json:"answer"`
	Votes     int       `json:"votes"`
	UserId    string    `json:"user_id"`
	TimeStamp time.Time `json:timestamp,omitempty`
	Url       string    `json:"Url,omitempty"`
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
	q.TimeStamp = time.Now()
	rand.Seed(time.Now().UTC().UnixNano())
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
		wq := elastic.NewWildcardQuery("title", "*"+val+"*")
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
	for _, val := range values {
		wq := elastic.NewWildcardQuery("body", "*"+val+"*")
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
		if encountered[qs[v].Body] != true {
			encountered[qs[v].Body] = true
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
	q := elastic.NewTermQuery("id", id)
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
	fmt.Println(qs)
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
	fmt.Println(ans)
	fmt.Println(len(res))
	if len(res) == 0 {
		return nil
	}
	return &QAView{Qid: res[0].Id, Body: res[0].Body, Title: res[0].Title, Answer: ans}
}

func CreateAnswer(a Answer) {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	id1 := rand.Intn(10000)
	id2 := rand.Intn(10000)
	a.Id = id1 + id2 + a.QuesId
	a.TimeStamp = time.Now()
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
			fmt.Println("Error", err)
		}
		qs = append(qs, q)
	}
	if len(qs) < 0 {
		qs[0].Votes += 1
		client.Update().Index("questions").Type("_doc").Id(strconv.Itoa(qid)).Doc(qs[0]).Do(context.TODO())
	}
}

func DecrementQuestionVotes(qid int) {
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
			fmt.Println("Error", err)
		}
		qs = append(qs, q)
	}
	if len(qs) > 0 {
		qs[0].Votes -= 1
		client.Update().Index("questions").Type("_doc").Id(strconv.Itoa(qid)).Doc(qs[0]).Do(context.TODO())
	}
}
func IncrementAnswerVotes(qid int) {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}

	query := elastic.NewTermQuery("id", qid)
	result, err := client.Search().Query(query).Index("questions").Type("_doc").Do(context.TODO())
	fmt.Println(len(result.Hits.Hits), err)
	var as []Answer
	for _, hit := range result.Hits.Hits {
		var a Answer
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			fmt.Println("Error", err)
		}
		as = append(as, a)
	}
	if len(as) > 0 {
		as[0].Votes += 1
		client.Update().Index("questions").Type("_doc").Id(strconv.Itoa(qid)).Doc(as[0]).Do(context.TODO())
	}
}

func DecrementAnswerVotes(qid int) {
	client, err := elastic.NewClient()
	if err != nil {
		fmt.Println("Error", err)
	}

	query := elastic.NewTermQuery("id", qid)
	result, err := client.Search().Query(query).Index("questions").Type("_doc").Do(context.TODO())
	fmt.Println(len(result.Hits.Hits), err)
	var as []Answer
	for _, hit := range result.Hits.Hits {
		var a Answer
		err := json.Unmarshal(*hit.Source, &a)
		if err != nil {
			fmt.Println("Error", err)
		}
		as = append(as, a)
	}
	if len(as) > 0 {
		as[0].Votes -= 1
		client.Update().Index("questions").Type("_doc").Id(strconv.Itoa(qid)).Doc(as[0]).Do(context.TODO())
	}
}
func createQuestion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var question Question
	_ = json.NewDecoder(r.Body).Decode(&question)
	CreateQuestion(question)
	json.NewEncoder(w).Encode(question)
}

func getAllQuestions(w http.ResponseWriter, r *http.Request) {
	q := GetAllQuestions()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(q)
}
func getQuestionById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	q := GetQustionById(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(q)
}

func getQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var question QRequest
	_ = json.NewDecoder(r.Body).Decode(&question)
	s := GetQuestion(question.Que)
	json.NewEncoder(w).Encode(s)
}

func createAnswer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var answer Answer
	_ = json.NewDecoder(r.Body).Decode(&answer)
	CreateAnswer(answer)
	json.NewEncoder(w).Encode(answer)
}

func getAnswerByAID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	fmt.Println(id)
	a := GetAnswerByAID(id)
	fmt.Println(a)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(a)
}

func questionUpVote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	IncrementQuestionVotes(id)
	fmt.Println("vote updated")
}
func questionDownVote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	DecrementQuestionVotes(id)
	fmt.Println("vote updated")
}
func answerUpVote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	IncrementAnswerVotes(id)
	fmt.Println("vote updated")
}
func answerDownVote(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	DecrementAnswerVotes(id)
	fmt.Println("vote updated")
}
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/questions", getAllQuestions).Methods("GET")
	r.HandleFunc("/api/questions/{id}", getQuestionById).Methods("GET")
	r.HandleFunc("/api/questions/create", createQuestion).Methods("POST")
	r.HandleFunc("/api/questions/search", getQuestions).Methods("POST")
	r.HandleFunc("/api/answers/create", createAnswer).Methods("POST")
	r.HandleFunc("/api/answers/{id}", getAnswerByAID).Methods("GET")
	r.HandleFunc("/api/questions/upvote{id}", questionUpVote).Methods("GET")
	r.HandleFunc("/api/questions/downvote{id}", questionDownVote).Methods("GET")
	r.HandleFunc("/api/answers/upvote{id}", answerUpVote).Methods("GET")
	r.HandleFunc("/api/answers/downvote{id}", answerDownVote).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", r))

}
func unmarshalJSON(bytes []byte, docType reflect.Type) (interface{}, error) {
	obj := reflect.New(docType)
	doc := obj.Interface()
	if err := json.Unmarshal(bytes, doc); err != nil {
		return nil, fmt.Errorf("error while unmarshalling json document to struct of type %v: %v", docType, err)
	}
	return doc, nil
}
