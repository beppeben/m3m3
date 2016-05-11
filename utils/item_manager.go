package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/beppeben/m3m3/domain"
	"github.com/petar/GoLLRB/llrb"
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	max_int   int64   = 9223372036854775807
	mill_hour float64 = 1000 * 60 * 60
	mill_day  float64 = mill_hour * 24
)

type IManager struct {
	sortedItems  *llrb.LLRB
	itemsByTid   map[string]*Item
	itemsById    map[int64]*Item
	itemsByTitle map[string]*Item
	mutex        *sync.RWMutex
	json         string
	json_zipped  []byte
	max_items    int
	max_show     int
	max_on_top   int
	max_comments int
	n_comments   int
	h            MHelper
}

type MHelper interface {
	DeleteTempImage(tid string) error
}

func NewManager(helper MHelper) *IManager {
	m := &IManager{}
	m.sortedItems = llrb.New()
	m.itemsByTid = make(map[string]*Item)
	m.itemsById = make(map[int64]*Item)
	m.itemsByTitle = make(map[string]*Item)
	m.mutex = &sync.RWMutex{}
	m.max_items = 120
	m.max_show = 100
	m.max_on_top = 5
	m.max_comments = 20
	m.h = helper
	return m
}

type Item struct {
	domain.Item
}

func (x Item) Less(than llrb.Item) bool {
	return x.Score < than.(*Item).Score
}

func (item *Item) UpdateScore() {
	result := item.Time + int64(math.Sqrt(float64(item.Ncomments))*mill_day/2)
	if item.BestComment != nil {
		result += int64(math.Sqrt(float64(item.BestComment.Likes)) * mill_day / 2)
	}
	item.Score = result
}

func (m *IManager) MaxShowItems() int {
	return m.max_show
}

func (m *IManager) Count() int {
	return len(m.itemsByTid)
}

func (m *IManager) NotifyComment(comment *domain.Comment) {
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
	if prev_comm == nil {
		m.n_comments++
	}
	if comment.Likes == 0 {
		item.Ncomments++
	}
	item.UpdateScore()
	m.sortedItems.InsertNoReplace(item)
	m.removeTail()
	m.adjustScores()
	m.refreshJson()
}

func (m *IManager) NotifyItemId(tid string, id int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	item, ok := m.itemsByTid[tid]
	if !ok {
		return
	}
	item.Id = id
	m.itemsById[id] = item
}

func (m *IManager) IsManaged(item *domain.Item) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isManaged(&Item{*item})
}

func (m *IManager) isManaged(item *Item) bool {
	_, ok := m.itemsByTid[item.Tid]
	if !ok {
		_, ok = m.itemsById[item.Id]
	}
	if !ok {
		_, ok = m.itemsByTitle[item.Title]
	}
	return ok
}

func (m *IManager) GetItemByTid(tid string) (*domain.Item, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	item, ok := m.itemsByTid[tid]
	if ok {
		return &item.Item, true
	} else {
		return nil, false
	}
}

func (m *IManager) GetItemById(id int64) (*domain.Item, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	item, ok := m.itemsById[id]
	if ok {
		return &item.Item, true
	} else {
		return nil, false
	}
}

func (m *IManager) Insert(it *domain.Item) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	item := &Item{*it}
	if m.isManaged(item) {
		return false
	}
	//set timestamp partly randomly to shuffle items
	item.Time = time.Now().UnixNano()/1000000 + rand.Int63n(1000)
	item.UpdateScore()
	m.sortedItems.InsertNoReplace(item)
	m.itemsByTid[item.Tid] = item
	m.itemsByTitle[item.Title] = item
	item.Src.Increase()
	m.removeTail()
	return true
}

func (m *IManager) Remove(it *domain.Item) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	item := &Item{*it}
	if !m.isManaged(item) {
		return false
	}
	m.sortedItems.Delete(item)
	m.cleanItem(item)
	//it.Ncomments = item.Ncomments
	m.refreshJson()
	return true
}

func (m *IManager) RefreshJson() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.refreshJson()
}

//sets score of (n+1)th item to the "natural" one only based on time, so it will be moved by new feeds
func (m *IManager) adjustScores() {
	temp := m.max_on_top
	var item *Item
	m.sortedItems.DescendLessOrEqual(&Item{domain.Item{Score: max_int}}, func(i llrb.Item) bool {
		it := i.(*Item)
		if it.BestComment == nil {
			return false
		}
		if temp > 0 {
			temp--
			return true
		} else if temp == 0 {
			item = it
			return true
		} else {
			return false
		}
	})
	if item != nil {
		m.sortedItems.Delete(item)
		item.Score = time.Now().UnixNano() / 1000000
		m.sortedItems.InsertNoReplace(item)
	}
}

func (m *IManager) refreshJson() {
	l := len(m.itemsByTid)
	if l > m.max_show {
		l = m.max_show
	}
	var ary = make([]*Item, l)
	count := 0
	m.sortedItems.DescendLessOrEqual(&Item{domain.Item{Score: max_int}}, func(i llrb.Item) bool {
		if count >= l {
			return false
		}
		ary[count] = i.(*Item)
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
func (m *IManager) removeTail() {
	//delete item with comment with least score if there are too many
	if m.n_comments > m.max_comments {
		var it *Item
		m.sortedItems.AscendGreaterOrEqual(&Item{domain.Item{Score: 0}}, func(i llrb.Item) bool {
			cand := i.(*Item)
			if cand.BestComment != nil {
				it = cand
				return false
			}
			return true
		})
		if it != nil {
			m.sortedItems.Delete(it)
			m.cleanItem(it)
		}
	}
	l := len(m.itemsByTid)
	if l <= m.max_items {
		return
	}
	for i := 0; i < l-m.max_items; i++ {
		min := m.sortedItems.DeleteMin().(*Item)
		m.cleanItem(min)
	}
	if m.sortedItems.Len() != len(m.itemsByTid) {
		panic("Manager lists have different lengths!!")
	}
}

func (m *IManager) cleanItem(item *Item) {
	delete(m.itemsByTid, item.Tid)
	delete(m.itemsById, item.Id)
	delete(m.itemsByTitle, item.Title)
	item.Src.Decrease()
	if item.BestComment != nil {
		m.n_comments--
		item.Ncomments--
	}
	err := m.h.DeleteTempImage(item.Tid)
	if err != nil {
		log.Warnf("Manager Error: %v", err)
	}
}

func (m *IManager) GetJson() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.json
}

func (m *IManager) GetZippedJson() []byte {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.json_zipped
}
