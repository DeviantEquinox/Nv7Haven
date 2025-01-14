package treecmds

import (
	"fmt"

	"github.com/Nv7-Github/Nv7Haven/eod/trees"
	"github.com/Nv7-Github/Nv7Haven/eod/types"
)

func (b *TreeCmds) ElemBreakdownCmd(elem string, calcTree bool, m types.Msg, rsp types.Rsp) {
	rsp.Acknowledge()

	b.lock.RLock()
	dat, exists := b.dat[m.GuildID]
	b.lock.RUnlock()
	if !exists {
		rsp.ErrorMessage("Guild isn't setup yet!")
		return
	}

	el, res := dat.GetElement(elem)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	tree := &trees.BreakDownTree{
		Dat:       dat,
		Breakdown: make(map[string]int),
		Added:     make(map[string]types.Empty),
		Tree:      calcTree,
		Total:     0,
	}
	suc, err := tree.AddElem(el.Name)
	if !suc {
		rsp.ErrorMessage(err)
		return
	}

	b.base.NewPageSwitcher(types.PageSwitcher{
		Kind:       types.PageSwitchInv,
		Title:      fmt.Sprintf("%s Breakdown (%d)", el.Name, tree.Total),
		PageGetter: b.base.InvPageGetter,
		Items:      tree.GetStringArr(),
	}, m, rsp)
}

func (b *TreeCmds) CatBreakdownCmd(catName string, calcTree bool, m types.Msg, rsp types.Rsp) {
	rsp.Acknowledge()

	b.lock.RLock()
	dat, exists := b.dat[m.GuildID]
	b.lock.RUnlock()
	if !exists {
		rsp.ErrorMessage("Guild isn't setup yet!")
		return
	}

	cat, res := dat.GetCategory(catName)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	tree := &trees.BreakDownTree{
		Dat:       dat,
		Breakdown: make(map[string]int),
		Added:     make(map[string]types.Empty),
		Tree:      calcTree,
		Total:     0,
	}

	for elem := range cat.Elements {
		suc, err := tree.AddElem(elem)
		if !suc {
			rsp.ErrorMessage(err)
			return
		}
	}

	b.base.NewPageSwitcher(types.PageSwitcher{
		Kind:       types.PageSwitchInv,
		Title:      fmt.Sprintf("%s Breakdown (%d)", cat.Name, tree.Total),
		PageGetter: b.base.InvPageGetter,
		Items:      tree.GetStringArr(),
	}, m, rsp)
}

func (b *TreeCmds) InvBreakdownCmd(user string, calcTree bool, m types.Msg, rsp types.Rsp) {
	rsp.Acknowledge()

	b.lock.RLock()
	dat, exists := b.dat[m.GuildID]
	b.lock.RUnlock()
	if !exists {
		rsp.ErrorMessage("Guild isn't setup yet!")
		return
	}

	inv, res := dat.GetInv(user, user == m.Author.ID)
	if !res.Exists {
		rsp.ErrorMessage(res.Message)
		return
	}

	tree := &trees.BreakDownTree{
		Dat:       dat,
		Breakdown: make(map[string]int),
		Added:     make(map[string]types.Empty),
		Tree:      calcTree,
		Total:     0,
	}

	for elem := range inv.Elements {
		/*suc, err :=*/ tree.AddElem(elem, true)
		/*	if !suc {
			rsp.ErrorMessage(err)
			return
		}*/
	}

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
		Title:      fmt.Sprintf("%s's Inventory Breakdown (%d)", name, tree.Total),
		PageGetter: b.base.InvPageGetter,
		Items:      tree.GetStringArr(),
	}, m, rsp)
}
