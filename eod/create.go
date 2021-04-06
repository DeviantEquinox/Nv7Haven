package eod

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const newText = "🆕"

var datafile *os.File

func (b *EoD) elemCreate(name string, parents []string, creator string, guild string) {
	lock.RLock()
	dat, exists := b.dat[guild]
	lock.RUnlock()
	if !exists {
		return
	}

	data := elems2txt(parents)
	row := b.db.QueryRow("SELECT COUNT(1) FROM eod_combos WHERE guild=? AND elems=?", guild, data)
	var count int
	err := row.Scan(&count)
	if err != nil {
		fmt.Println(103, err)
		return
	}
	if count != 0 {
		return
	}

	row = b.db.QueryRow("SELECT COUNT(1) FROM eod_elements WHERE name=? AND guild=?", name, guild)
	err = row.Scan(&count)
	if err != nil {
		fmt.Println(23, err)
		return
	}
	text := "Combination"
	if count == 0 {
		diff := -1
		compl := -1
		areUnique := false
		for _, val := range parents {
			elem := dat.elemCache[strings.ToLower(val)]
			if elem.Difficulty > diff {
				diff = elem.Difficulty
			}
			if elem.Complexity > compl {
				compl = elem.Complexity
			}
			if !strings.EqualFold(parents[0], val) {
				areUnique = true
			}
		}
		compl++
		if areUnique {
			diff++
		}
		elem := element{
			Name:       name,
			Categories: make(map[string]empty),
			Guild:      guild,
			Comment:    "None",
			Creator:    creator,
			CreatedOn:  time.Now(),
			Parents:    parents,
			Complexity: compl,
			Difficulty: diff,
		}
		dat.elemCache[strings.ToLower(elem.Name)] = elem
		dat.invCache[creator][strings.ToLower(elem.Name)] = empty{}
		lock.Lock()
		b.dat[guild] = dat
		lock.Unlock()
		cats, err := json.Marshal(elem.Categories)
		if err != nil {
			fmt.Println(65, err)
			return
		}

		pars := make(map[string]empty, len(parents))
		for _, val := range parents {
			pars[val] = empty{}
		}
		dat, err := json.Marshal(pars)
		if err != nil {
			fmt.Println(75, err)
			return
		}
		_, err = b.db.Exec("INSERT INTO eod_elements VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )", elem.Name, string(cats), elem.Image, elem.Guild, elem.Comment, elem.Creator, int(elem.CreatedOn.Unix()), string(dat), elem.Complexity, elem.Difficulty, 0)
		if err != nil {
			fmt.Println(80, err)
			return
		}
		text = "Element"

		b.saveInv(guild, creator)
	} else {
		el, exists := dat.elemCache[strings.ToLower(name)]
		if !exists {
			return
		}
		name = el.Name

		dat.invCache[creator][strings.ToLower(name)] = empty{}
		lock.Lock()
		b.dat[guild] = dat
		lock.Unlock()
		b.saveInv(guild, creator)
	}
	b.db.Exec("INSERT INTO eod_combos VALUES ( ?, ?, ? )", guild, data, name)

	params := make(map[string]empty)
	for _, val := range parents {
		params[val] = empty{}
	}
	for k := range params {
		b.db.Exec("UPDATE eod_elements SET usedin=usedin+1 WHERE name=? AND guild=?", k, guild)
		el := dat.elemCache[strings.ToLower(k)]
		el.UsedIn++
		dat.elemCache[strings.ToLower(k)] = el
	}
	lock.Lock()
	b.dat[guild] = dat
	lock.Unlock()
	b.dg.ChannelMessageSend(dat.newsChannel, newText+" "+text+" - **"+name+"** (By <@"+creator+">)")
	if guild == "819077688371314718" {
		datafile.Write([]byte(fmt.Sprintf("%s %s\n", name, parents)))
	}
}
