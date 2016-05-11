package crawler

type Feed struct {
	url        string
	name       string
	nb_managed int
}

func (feed *Feed) Increase() {
	feed.nb_managed++
}

func (feed *Feed) Decrease() {
	feed.nb_managed++
}

type ByNbManaged []*Feed

func (a ByNbManaged) Len() int           { return len(a) }
func (a ByNbManaged) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNbManaged) Less(i, j int) bool { return a[i].nb_managed < a[j].nb_managed }
