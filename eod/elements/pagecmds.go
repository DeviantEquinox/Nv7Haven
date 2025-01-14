package elements

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/Nv7-Github/Nv7Haven/eod/util"
)

func (b *Elements) InvCmd(user string, m types.Msg, rsp types.Rsp, sorter string, filter string) {
	rsp.Acknowledge()

	b.lock.RLock()
	dat, exists := b.dat[m.GuildID]
	b.lock.RUnlock()
	if !exists {
		rsp.ErrorMessage("Guild not setup!")
		return
	}

	inv, res := dat.GetInv(user, user == m.Author.ID)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}
	items := make([]string, len(inv.Elements))
	i := 0
	dat.Lock.RLock()
	for k := range inv.Elements {
		el, _ := dat.GetElement(k, true)
		items[i] = el.Name
		i++
	}

	switch filter {
	case "madeby":
		count := 0
		outs := make([]string, len(items))
		for _, val := range items {
			creator := ""
			elem, res := dat.GetElement(val, true)
			if res.Exists {
				creator = elem.Creator
			}
			if creator == user {
				outs[count] = val
				count++
			}
		}
		outs = outs[:count]
		items = outs
	}
	util.SortElemList(items, sorter, dat)
	dat.Lock.RUnlock()

	name := m.Author.Username
	if m.Author.ID != user {
		u, err := b.dg.User(user)
		if rsp.Error(err) {
			return
		}
		name = u.Username
	}
	b.base.NewPageSwitcher(types.PageSwitcher{
		Kind:       types.PageSwitchInv,
		Title:      fmt.Sprintf("%s's Inventory (%d, %s%%)", name, len(items), util.FormatFloat(float32(len(items))/float32(len(dat.Elements))*100, 2)),
		PageGetter: b.base.InvPageGetter,
		Items:      items,
	}, m, rsp)
}

func (b *Elements) LbCmd(m types.Msg, rsp types.Rsp, sorter string, user string) {
	b.lock.RLock()
	dat, exists := b.dat[m.GuildID]
	b.lock.RUnlock()
	if !exists {
		return
	}
	_, res := dat.GetInv(user, user == m.Author.ID) // Check if user exists
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	// Sort invs
	invs := make([]types.Inventory, len(dat.Inventories))
	i := 0
	for _, v := range dat.Inventories {
		invs[i] = v
		i++
	}
	sortFn := func(a, b int) bool {
		return len(invs[a].Elements) > len(invs[b].Elements)
	}
	if sorter == "made" {
		sortFn = func(a, b int) bool {
			return invs[a].MadeCnt > invs[b].MadeCnt
		}
	}
	sort.Slice(invs, sortFn)

	// Convert to right format
	users := make([]string, len(invs))
	cnts := make([]int, len(invs))
	userpos := 0
	for i, v := range invs {
		users[i] = v.User
		if sorter == "count" {
			cnts[i] = len(v.Elements)
		} else {
			cnts[i] = v.MadeCnt
		}
		if v.User == user {
			userpos = i
		}
	}

	b.base.NewPageSwitcher(types.PageSwitcher{
		Kind:       types.PageSwitchLdb,
		Title:      "Top Most Elements",
		PageGetter: b.base.LbPageGetter,

		User:    user,
		Users:   users,
		UserPos: userpos,
		Cnts:    cnts,
	}, m, rsp)
}

func (b *Elements) SearchCmd(search string, sort string, source string, opt string, regex bool, m types.Msg, rsp types.Rsp) {
	b.lock.RLock()
	dat, exists := b.dat[m.GuildID]
	b.lock.RUnlock()
	if !exists {
		return
	}
	rsp.Acknowledge()
	_, res := dat.GetInv(m.Author.ID, true)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	var list map[string]types.Empty
	switch source {
	case "elements":
		list = make(map[string]types.Empty, len(dat.Elements))
		for _, el := range dat.Elements {
			list[el.Name] = types.Empty{}
		}

	case "inventory":
		inv, res := dat.GetInv(opt, m.Author.ID == opt)
		if !res.Exists {
			rsp.ErrorMessage(res.Message)
			return
		}

		list = make(map[string]types.Empty, len(inv.Elements))
		dat.Lock.RLock()
		for el := range inv.Elements {
			elem, res := dat.GetElement(el, true)
			if !res.Exists {
				list[el] = types.Empty{}
				continue
			}
			list[elem.Name] = types.Empty{}
		}
		dat.Lock.RUnlock()

	case "category":
		cat, res := dat.GetCategory(opt)
		if !res.Exists {
			rsp.ErrorMessage(res.Message)
			return
		}
		list = cat.Elements
	}

	items := make(map[string]types.Empty)
	if regex {
		reg, err := regexp.Compile(search)
		if rsp.Error(err) {
			return
		}
		for el := range list {
			m := reg.Find([]byte(el))
			if m != nil {
				items[el] = types.Empty{}
			}
		}
	} else {
		s := strings.ToLower(search)
		for el := range list {
			if strings.Contains(strings.ToLower(el), s) {
				items[el] = types.Empty{}
			}
		}
	}

	txt := make([]string, len(items))
	i := 0
	for k := range items {
		txt[i] = k
		i++
	}
	util.SortElemList(txt, sort, dat)

	if len(txt) == 0 {
		rsp.Message("No results!")
		return
	}

	b.base.NewPageSwitcher(types.PageSwitcher{
		Kind:       types.PageSwitchInv,
		Title:      fmt.Sprintf("Element Search (%d)", len(txt)),
		PageGetter: b.base.InvPageGetter,
		Items:      txt,
		User:       m.Author.ID,
	}, m, rsp)
}
