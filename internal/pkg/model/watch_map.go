package model

import "container/list"

type watchKey struct {
	wc   *watchClients
	elem *list.Element
}

type watchClients struct {
	key     string
	clients *list.List
}

type watchMap map[string]*watchClients

func newWatchMap() watchMap {
	return make(watchMap)
}

func (w watchMap) Put(cli *Client, key string) *watchKey {
	wc, ok := w[key]
	if !ok {
		wc = &watchClients{key, list.New()}
		w[key] = wc
	}
	return &watchKey{wc, wc.clients.PushBack(cli)}
}

func (w watchMap) Touch(key string) {
	wc, ok := w[key]
	if !ok {
		return
	}
	for e := wc.clients.Front(); e != nil; e = e.Next() {
		cli := e.Value.(*Client)
		if cli.Multi.State {
			cli.Multi.Dirty = true
		}
	}
}

func (w watchMap) Remove(wk *watchKey) {
	wk.wc.clients.Remove(wk.elem)
	if wk.wc.clients.Len() == 0 {
		delete(w, wk.wc.key)
	}
}
