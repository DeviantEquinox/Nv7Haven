package eod

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/Nv7-Github/Nv7Haven/dgutil"
	"github.com/Nv7-Github/Nv7Haven/eod/logs"
	"github.com/bwmarrin/discordgo"
)

const guild = "" // 819077688371314718 for testing

func (b *EoD) initHandlers() {
	// Debugging
	discordgo.Logger = func(msgL, caller int, format string, a ...interface{}) {
		// This code is a slightly modified version of https://github.com/bwmarrin/discordgo/blob/577e7dd4f6ccf1beb10acdb1871300c7638b84c4/logging.go#L46
		pc, file, line, _ := runtime.Caller(caller)

		files := strings.Split(file, "/")
		file = files[len(files)-1]

		name := runtime.FuncForPC(pc).Name()
		fns := strings.Split(name, ".")
		name = fns[len(fns)-1]

		msg := fmt.Sprintf(format, a...)

		log.SetOutput(logs.DiscordLogs)
		log.Printf("[DG%d] %s:%d:%s() %s\n", msgL, file, line, name, msg)
	}

	dgutil.UpdateBotCommands(b.dg, clientID, guild, commands)

	b.dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Member == nil {
			return
		}

		// Command
		if i.Type == discordgo.InteractionApplicationCommand {
			rsp := b.newRespSlash(i)
			canRun, msg := b.canRunCmd(i)
			if !canRun {
				rsp.ErrorMessage(msg)
				return
			}

			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
			return
		}

		// Button
		if i.Type == discordgo.InteractionMessageComponent {
			lock.Lock()
			dat, exists := b.dat[i.GuildID]
			if !exists {
				return
			}
			lock.Unlock()

			// Check if page switch handler or component handler
			_, exists = dat.PageSwitchers[i.Message.ID]
			if exists {
				b.base.PageSwitchHandler(s, i)
				return
			}

			compMsg, exists := dat.ComponentMsgs[i.Message.ID]
			if exists {
				compMsg.Handler(s, i)
				return
			}
			return
		}
	})
	b.dg.AddHandler(b.cmdHandler)
	b.dg.AddHandler(b.polls.ReactionHandler)
	b.dg.AddHandler(b.polls.UnReactionHandler)
}
