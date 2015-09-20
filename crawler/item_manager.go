package crawler

import (
	"github.com/petar/GoLLRB/llrb"
	"sync"
	"time"
	"math/rand"
	"encoding/json"
	"fmt"
	"bytes"
	"compress/gzip"
)


var (
	max_items 		int = 120
)

type IManager struct {
	sortedItems		*llrb.LLRB
	itemSet			map[string]*Item
	mutex			*sync.RWMutex
	json				string
	json_zipped		[]byte
}

func NewManager() *IManager {
	m := &IManager{}
	m.sortedItems = llrb.New()
	m.itemSet = make(map[string]*Item)
	m.mutex = &sync.RWMutex{}
	return m
}

//this corresponds to an image (a piece of news) shown in the feeds page
type Item struct {
	Img_url 			string		`json:"img_url,omitempty"`
	Img_id			int64		`json:"img_id,omitempty"`
	Title			string		`json:"title,omitempty"`
	Ncomments		int			`json:"n_comments"`
	score			int64		/*`json:"-"`*/	
}

type Comment struct {
	Id		int64		`json:"id,omitempty"`
	Text		string		`json:"text,omitempty"`
	User		string		`json:"user,omitempty"`
}

func (item Item) Less (than llrb.Item) bool {
	it := than.(*Item)
	return item.score < it.score
}


func (m *IManager) IsManaged (item *Item) bool {
	m.mutex.Lock()	
	_, ok := m.itemSet[item.Img_url]
	m.mutex.Unlock()
	return ok
}


func (m *IManager) Insert (item *Item) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.itemSet[item.Img_url]
	if ok {return false}
	//set timestamp partly randomly to shuffle items
	item.score = time.Now().UnixNano()/1000000 + rand.Int63n(1000)
	m.sortedItems.InsertNoReplace(item)
	m.itemSet[item.Img_url] = item
	m.removeTail()
	return true
}


func (m *IManager) RefreshJson(){
	var ary []*Item
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.sortedItems.AscendGreaterOrEqual(&Item{}, func(i llrb.Item) bool {
		ary = append(ary, i.(*Item))
		return true
	})
	b, err := json.Marshal(ary)
    if err != nil {
        fmt.Println(err)
        return
    }
	m.json = string(b)
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	w.Write([]byte(m.json))
	w.Flush()
	m.json_zipped = buf.Bytes()
}

func (m *IManager) removeTail(){
	l := len(m.itemSet)
	if l <= max_items {return}	
	for i := 0; i < l-max_items; i++ {
		min := m.sortedItems.DeleteMin().(*Item)
		delete(m.itemSet, min.Img_url)
	}
	if (m.sortedItems.Len() != len(m.itemSet)){
		panic("manager lists have different lengths!!")
	}
}

func (m *IManager) GetJson() string{
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.json
}

func (m *IManager) GetZippedJson() []byte{
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.json_zipped
}