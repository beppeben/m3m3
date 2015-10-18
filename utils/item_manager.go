package utils

import (
	"github.com/petar/GoLLRB/llrb"
	"sync"
	"time"
	"math/rand"
	"encoding/json"
	//"fmt"
	"bytes"
	"compress/gzip"
	"math"
	"log"
)


var (
	max_items 		int = 150
	max_show			int = 100
	mill_hour		float64 = 1000*60*60
	mill_day			float64 = mill_hour*24
	max_int			int64 = 9223372036854775807
)

type IManager struct {
	sortedItems		*llrb.LLRB
	itemsByTid		map[string]*Item
	itemsById		map[int64]*Item
	mutex			*sync.RWMutex
	json				string
	json_zipped		[]byte
}

//this corresponds to an image (a piece of news) shown in the feeds page
type Item struct {
	Tid	 			string		`json:"item_tid,omitempty"`
	Id				int64		`json:"item_id,omitempty"`
	Title			string		`json:"title,omitempty"`
	Source			string 		`json:"source,omitempty"`
	Ncomments		int			`json:"-"`
	BestComment		*Comment		`json:"b_comm,omitempty"`
	Url				string		`json:"img_url,omitempty"`
	time				int64		/*`json:"-"`*/	
	score 			int64
}

func NewManager() *IManager {
	m := &IManager{}
	m.sortedItems = llrb.New()
	m.itemsByTid = make(map[string]*Item)
	m.itemsById = make(map[int64]*Item)
	m.mutex = &sync.RWMutex{}
	return m
}

func (item Item) Less (than llrb.Item) bool {
	it := than.(*Item)
	return item.score < it.score
}

func (item *Item) updateScore() {
	result := item.time + int64(math.Sqrt(float64(item.Ncomments))*mill_day/2)
	if item.BestComment != nil {
		result += int64(math.Sqrt(float64(item.BestComment.Likes))*mill_day/2)
	}
	item.score = result
}

func (m *IManager) NotifyComment (comment *Comment) {
	//could add a check here to see if item is managed only with a read lock
	m.mutex.Lock()
	defer m.mutex.Unlock()	
	item, ok := m.itemsById[comment.Item_id]
	if !ok {
		//the item is no longer managed, so don't bother updating the list
		return
	}
	//only update the best comment if it is null or has zero likes
	m.sortedItems.Delete(item)
	prev_comm := item.BestComment
	if prev_comm == nil || comment.Likes >= prev_comm.Likes {
		item.BestComment = comment
	}
	if comment.Likes == 0 {
		item.Ncomments = item.Ncomments + 1
	}	
	item.updateScore()
	m.sortedItems.InsertNoReplace(item)
	m.refreshJson()	
}

func (m *IManager) NotifyItemId (tid string, id int64) {
	m.mutex.Lock()	
	defer m.mutex.Unlock()
	item, ok := m.itemsByTid[tid]
	if !ok {
		return
	}
	item.Id = id
	m.itemsById[id] = item
}


func (m *IManager) IsManaged (item *Item) bool {
	m.mutex.RLock()	
	defer m.mutex.RUnlock()
	_, ok := m.itemsByTid[item.Tid]
	if !ok {
		_, ok = m.itemsById[item.Id]
	}
	return ok
}

func (m *IManager) GetItemByTid (tid string) (*Item, bool) {
	m.mutex.RLock()	
	defer m.mutex.RUnlock()
	item, ok := m.itemsByTid[tid]
	return item, ok
}

func (m *IManager) GetItemById (id int64) (*Item, bool) {
	m.mutex.RLock()	
	defer m.mutex.RUnlock()
	item, ok := m.itemsById[id]
	return item, ok
}


func (m *IManager) Insert (item *Item) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, ok := m.itemsByTid[item.Tid]
	if ok {return false}
	//set timestamp partly randomly to shuffle items
	item.time = time.Now().UnixNano()/1000000 + rand.Int63n(1000)
	item.updateScore()
	m.sortedItems.InsertNoReplace(item)
	m.itemsByTid[item.Tid] = item
	m.removeTail()
	return true
}

func (m *IManager) RefreshJson(){	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.refreshJson()
}

func (m *IManager) refreshJson(){	
	l := len(m.itemsByTid)
	if l > max_show {
		l = max_show
	}
	var ary = make([]*Item, l)	
	count := 0
	m.sortedItems.DescendLessOrEqual(&Item{score: max_int}, func(i llrb.Item) bool {
		if count >= l {
			return true
		}
		ary[count] = i.(*Item)
		//if _, ok := m.itemsByTid[ary[count].Tid]; !ok {
		//	log.Printf("WARNING!!!! Item %s is in the tree but not in the index!", ary[count].Tid)
		//}		
		count++
		return true
	})
	b, err := json.Marshal(ary)
    if err != nil {
        return
    }
	m.json = string(b)
	buf := new(bytes.Buffer)
	w := gzip.NewWriter(buf)
	w.Write([]byte(m.json))
	w.Flush()
	m.json_zipped = buf.Bytes()
}

//to be called inside a lock
func (m *IManager) removeTail(){
	l := len(m.itemsByTid)
	if l <= max_items {return}	
	for i := 0; i < l-max_items; i++ {
		min := m.sortedItems.DeleteMin().(*Item)
		delete(m.itemsByTid, min.Tid)
		delete(m.itemsById, min.Id)
		err := DeleteTempImage(min.Tid)
		if err != nil {
			log.Printf("[CRAW] Error: %v", err)
		}
	}
	if (m.sortedItems.Len() != len(m.itemsByTid)){
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