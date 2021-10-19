package elements

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/eod/types"
	"github.com/Nv7-Github/Nv7Haven/eod/util"
	"github.com/bwmarrin/discordgo"
)

var ideaCmp = discordgo.ActionsRow{
	Components: []discordgo.MessageComponent{
		discordgo.Button{
			Label:    "New Idea",
			Style:    discordgo.SuccessButton,
			CustomID: "idea",
		},
	},
}

type ideaComponent struct {
	catName  string
	hasCat   bool
	elemName string
	hasEl    bool
	count    int
	b        *Elements
}

func (c *ideaComponent) Handler(_ *discordgo.Session, i *discordgo.InteractionCreate) {
	res, suc := c.b.genIdea(c.count, c.catName, c.hasCat, c.elemName, c.hasEl, i.GuildID, i.Member.User.ID)
	if !suc {
		res += " " + types.RedCircle
	}
	err := c.b.dg.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    res,
			Components: []discordgo.MessageComponent{ideaCmp},
		},
	})
	if err != nil {
		fmt.Println("Failed to send message:", err)
	}
}

func (b *Elements) genIdea(count int, catName string, hasCat bool, elemName string, hasEl bool, guild string, author string) (string, bool) {
	if count > types.MaxComboLength {
		return fmt.Sprintf("You can only combine up to %d elements!", types.MaxComboLength), false
	}

	if count < 2 {
		return "You must combine at least 2 elements!", false
	}

	b.lock.RLock()
	dat, exists := b.dat[guild]
	b.lock.RUnlock()
	if !exists {
		return "Guild not found", false
	}

	inv, res := dat.GetInv(author, true)
	if !res.Exists {
		return res.Message, false
	}

	if hasEl {
		elName := strings.ToLower(elemName)

		el, res := dat.GetElement(elName)
		if !res.Exists {
			return res.Message, false
		} else {
			elemName = elName
			count--
		}

		exists = inv.Elements.Contains(elemName)
		if !exists {
			return fmt.Sprintf("Element **%s** is not in your inventory!", el.Name), false
		}
	}

	els := inv.Elements
	if hasCat {
		cat, res := dat.GetCategory(catName)
		if !res.Exists {
			return res.Message, false
		}
		els = make(map[string]types.Empty)

		for el := range cat.Elements {
			exists := inv.Elements.Contains(el)
			if exists {
				els[strings.ToLower(el)] = types.Empty{}
			}
		}

		if len(els) == 0 {
			return fmt.Sprintf("You don't have any elements in category **%s**!", cat.Name), false
		}
	}

	res = types.GetResponse{Exists: true}
	var elems []string
	tries := 0
	for res.Exists {
		elems = make([]string, count)
		for i := range elems {
			cnt := rand.Intn(len(els))
			j := 0
			for k := range els {
				if j == cnt {
					elems[i] = k
					break
				}
				j++
			}
		}
		if hasEl {
			elems = append([]string{elemName}, elems...)
		}

		_, res = dat.GetCombo(util.Elems2Txt(elems))
		tries++

		if tries > 21 {
			return "Couldn't find a random unused combination, maybe try again later?", false
		}
	}

	text := ""
	for i, el := range elems {
		el, _ := dat.GetElement(el)
		text += el.Name
		if i != len(elems)-1 {
			text += " + "
		}
	}

	dat.SetComb(author, types.Comb{
		Elems: elems,
		Elem3: "",
	})

	b.lock.Lock()
	b.dat[guild] = dat
	b.lock.Unlock()

	return fmt.Sprintf("Your random unused combination is... **%s**\n 	Suggest it by typing **/suggest**", text), true
}
func (b *Elements) IdeaCmd(count int, catName string, hasCat bool, elemName string, hasEl bool, m types.Msg, rsp types.Rsp) {
	res, suc := b.genIdea(count, catName, hasCat, elemName, hasEl, m.GuildID, m.Author.ID)
	if !suc {
		rsp.ErrorMessage(res)
		return
	}
	rsp.Acknowledge()

	b.lock.Lock()
	dat, exists := b.dat[m.GuildID]
	b.lock.Unlock()
	if !exists {
		rsp.ErrorMessage("Guild not found")
		return
	}

	id := rsp.Message(res, ideaCmp)

	dat.AddComponentMsg(id, &ideaComponent{
		catName:  catName,
		count:    count,
		hasCat:   hasCat,
		elemName: elemName,
		hasEl:    hasEl,
		b:        b,
	})

	b.lock.Lock()
	b.dat[m.GuildID] = dat
	b.lock.Unlock()
}