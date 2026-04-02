package console

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/8op-org/gl1tch/internal/busd"
	"github.com/8op-org/gl1tch/internal/busd/topics"
	"github.com/8op-org/gl1tch/internal/executor"
	"github.com/8op-org/gl1tch/internal/game"
	"github.com/8op-org/gl1tch/internal/pipeline"
	"github.com/8op-org/gl1tch/internal/store"
)

// SignalHandlerRegistry maps handler names to dispatch functions.
// Plugins reference handlers by name in their signals block.
type SignalHandlerRegistry map[string]func(topic, payload string)

// BuildSignalHandlerRegistry constructs a registry with the built-in handlers.
// narrationCh receives companion narration strings; st is used by the score handler.
// pack is the active game world pack, used for MUD XP event weights.
func BuildSignalHandlerRegistry(narrationCh chan<- string, st *store.Store, pack game.GameWorldPack) SignalHandlerRegistry {
	eng := game.NewGameEngine()
	reg := SignalHandlerRegistry{
		"companion":                 companionHandler(eng, narrationCh),
		"score":                     scoreHandler(st),
		"log":                       logHandler(),
		"npc-memory":                npcMemoryHandler(st),
		"npc-narrate":               npcNarrateHandler(narrationCh, st),
		"game-achievement-unlocked": achievementUnlockedHandler(narrationCh),
		"game-ice-encountered":      iceEncounteredHandler(st, pack),
		"game-quest-event":          logHandler(), // re-use log handler for quest events
		"game-bounty-completed":     bountyCompletedHandler(st, pack),
	}
	// Register MUD XP bridge handlers for each topic in the pack's mud_xp_events map.
	for topic, xp := range pack.Weights.MUDXPEvents {
		topic, xp := topic, xp // capture loop vars
		reg["mud-xp-"+topic] = mudXPBridgeHandler(st, topic, xp)
	}
	return reg
}

// achievementUnlockedHandler pushes a styled notification to the narration channel.
func achievementUnlockedHandler(ch chan<- string) func(topic, payload string) {
	return func(topic, payload string) {
		var p struct {
			AchievementID string `json:"achievement_id"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil || p.AchievementID == "" {
			return
		}
		msg := fmt.Sprintf("\x1b[95m[ACHIEVEMENT UNLOCKED]\x1b[0m %s", p.AchievementID)
		if ch != nil {
			select {
			case ch <- msg:
			default:
			}
		}
	}
}

// iceEncounteredHandler inserts an ICE encounter row and logs the event.
func iceEncounteredHandler(st *store.Store, pack game.GameWorldPack) func(topic, payload string) {
	return func(topic, payload string) {
		var p struct {
			ICEClass string `json:"ice_class"`
			RunID    string `json:"run_id"`
		}
		if err := json.Unmarshal([]byte(payload), &p); err != nil || p.ICEClass == "" {
			return
		}
		if st != nil {
			timeoutHours := pack.ICEEncounter.TimeoutHours
			if timeoutHours <= 0 {
				timeoutHours = 24
			}
			deadline := time.Now().Add(time.Duration(timeoutHours) * time.Hour)
			id := fmt.Sprintf("ice-%d", time.Now().UnixNano())
			if err := st.InsertICEEncounter(id, p.ICEClass, p.RunID, deadline); err != nil {
				log.Printf("[WARN] ice-encountered: insert encounter: %v", err)
			}
		}
		// Also log to signal log.
		logHandler()(topic, payload)
	}
}

// mudXPBridgePayload is the expected shape of MUD signal payloads that carry a signal ID.
type mudXPBridgePayload struct {
	SignalID string `json:"signal_id"`
	RoomID   string `json:"room_id"`
}

// mudXPBridgeHandler awards a fixed XP amount for a MUD event, deduped by signal_id.
func mudXPBridgeHandler(st *store.Store, topic string, xp int) func(topic, payload string) {
	return func(t, payload string) {
		if st == nil || xp <= 0 {
			return
		}
		var p mudXPBridgePayload
		_ = json.Unmarshal([]byte(payload), &p)
		runID := p.SignalID
		if runID == "" {
			runID = fmt.Sprintf("mud-%s-%d", topic, time.Now().UnixNano())
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ev := store.ScoreEvent{
			XP:        int64(xp),
			Provider:  "mud",
			Model:     topic,
			CreatedAt: time.Now().UnixMilli(),
		}
		if err := st.RecordScoreEvent(ctx, ev); err != nil {
			log.Printf("[DEBUG] mud-xp-bridge: record score event for %s: %v", topic, err)
		}
		if err := st.InsertOrUpdatePersonalBest("highest_xp", float64(xp), runID); err != nil {
			log.Printf("[DEBUG] mud-xp-bridge: update personal best: %v", err)
		}
	}
}

// bountyCompletedPayload is the expected shape for game.bounty.completed signals.
type bountyCompletedPayload struct {
	ContractID string `json:"contract_id"`
}

// bountyCompletedHandler validates and awards XP for a completed bounty contract.
func bountyCompletedHandler(st *store.Store, pack game.GameWorldPack) func(topic, payload string) {
	return func(topic, payload string) {
		if st == nil {
			return
		}
		var p bountyCompletedPayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil || p.ContractID == "" {
			log.Printf("[WARN] bounty-completed: cannot parse payload: %v", err)
			return
		}
		// Find the active contract.
		var contract *game.BountyContract
		for i := range pack.BountyContracts {
			c := pack.BountyContracts[i]
			if c.ID == p.ContractID {
				contract = &c
				break
			}
		}
		if contract == nil {
			log.Printf("[WARN] bounty-completed: contract %q not found", p.ContractID)
			return
		}
		if !contract.ValidUntil.IsZero() && contract.ValidUntil.Before(time.Now()) {
			log.Printf("[WARN] bounty-completed: contract %q expired", p.ContractID)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		runID := fmt.Sprintf("bounty-%s-%d", p.ContractID, time.Now().UnixNano())
		ev := store.ScoreEvent{
			XP:        int64(contract.XPReward),
			Provider:  "bounty",
			Model:     p.ContractID,
			CreatedAt: time.Now().UnixMilli(),
		}
		if err := st.RecordScoreEvent(ctx, ev); err != nil {
			log.Printf("[DEBUG] bounty-completed: record score event: %v", err)
		}
		if err := st.InsertOrUpdatePersonalBest("highest_xp", float64(contract.XPReward), runID); err != nil {
			log.Printf("[DEBUG] bounty-completed: update personal best: %v", err)
		}
		// Emit signal to notify consumers (e.g. publish back on bus).
		if sockPath, err := busd.SocketPath(); err == nil {
			_ = busd.PublishEvent(sockPath, topics.GameBountyCompleted, map[string]any{
				"contract_id": p.ContractID,
				"xp_reward":   contract.XPReward,
			})
		}
	}
}

// Dispatch looks up the handler for name and calls it with topic and payload.
// Unknown handlers emit a debug log and drop the event.
func (r SignalHandlerRegistry) Dispatch(name, topic, payload string) {
	h, ok := r[name]
	if !ok {
		log.Printf("[DEBUG] signal_handlers: unknown handler %q for topic %s — event dropped", name, topic)
		return
	}
	h(topic, payload)
}

const pluginCompanionPrompt = `You are gl1tch, a cynical AI companion watching a plugin event.
React to what just happened in 2-4 lines. Terse. Dry. Occasionally helpful. Never cheerful.
Reference the event naturally — don't just repeat the JSON. No markdown. No bullet points.`

func companionHandler(eng *game.GameEngine, ch chan<- string) func(topic, payload string) {
	return func(topic, payload string) {
		if ch == nil {
			return
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			userMsg := fmt.Sprintf("Event: %s\nPayload: %s", topic, payload)
			result := eng.Respond(ctx, pluginCompanionPrompt, userMsg)
			if result != "" && ch != nil {
				ch <- result
			}
		}()
	}
}

// tokenUsagePayload is the expected JSON shape for the score handler.
type tokenUsagePayload struct {
	Input  int64  `json:"input"`
	Output int64  `json:"output"`
	Model  string `json:"model"`
}

func scoreHandler(st *store.Store) func(topic, payload string) {
	return func(topic, payload string) {
		if st == nil {
			return
		}
		var usage tokenUsagePayload
		if err := json.Unmarshal([]byte(payload), &usage); err != nil {
			log.Printf("[DEBUG] signal_handlers: score handler: cannot parse payload for %s: %v", topic, err)
			return
		}
		xpResult := game.ComputeXP(game.TokenUsage{
			InputTokens:  usage.Input,
			OutputTokens: usage.Output,
		}, 0, game.DefaultPackWeights())
		ev := store.ScoreEvent{
			XP:           xpResult.Final,
			InputTokens:  usage.Input,
			OutputTokens: usage.Output,
			Provider:     topic,
			Model:        usage.Model,
			CreatedAt:    time.Now().Unix(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := st.RecordScoreEvent(ctx, ev); err != nil {
			log.Printf("[DEBUG] signal_handlers: score handler: record event: %v", err)
		}
	}
}

// npcMemoryPayload is the expected JSON shape for the npc-memory handler.
type npcMemoryPayload struct {
	NPCID        string `json:"npc_id"`
	NPCName      string `json:"npc_name"`
	Trigger      string `json:"trigger"`
	Text         string `json:"text"`
	StealthLevel int    `json:"stealth_level"`
}

func npcMemoryHandler(st *store.Store) func(topic, payload string) {
	return func(topic, payload string) {
		if st == nil {
			return
		}
		var p npcMemoryPayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			fmt.Fprintf(os.Stderr, "[npc-memory] cannot parse payload for %s: %v\n", topic, err)
			return
		}
		if p.NPCID == "" || p.NPCName == "" {
			fmt.Fprintf(os.Stderr, "[npc-memory] missing required fields (npc_id, npc_name) for %s\n", topic)
			return
		}
		body := fmt.Sprintf(
			"Player triggered %q with NPC %s (%s). NPC said: %q. Stealth: %d.",
			p.Trigger, p.NPCName, p.NPCID, p.Text, p.StealthLevel,
		)
		note := store.BrainNote{
			RunID:     0,
			StepID:    "npc-" + p.NPCID,
			CreatedAt: time.Now().Unix(),
			Tags:      "mud,npc-" + p.NPCID,
			Body:      body,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := st.InsertBrainNote(ctx, note); err != nil {
			fmt.Fprintf(os.Stderr, "[npc-memory] failed to write brain note: %v\n", err)
		}
	}
}

// npcNarrateHandler runs the mud-npc-narrate pipeline with brain injection so
// the Ollama narration has access to prior interaction notes for this NPC.
func npcNarrateHandler(narrationCh chan<- string, st *store.Store) func(topic, payload string) {
	pipelinePath := filepath.Join(os.Getenv("HOME"), "Projects", "gl1tch-mud", "pipelines", "mud-npc-narrate.pipeline.yaml")
	wrappersDir := filepath.Join(os.Getenv("HOME"), ".config", "glitch", "wrappers")

	return func(topic, payload string) {
		if narrationCh == nil {
			return
		}
		var p npcMemoryPayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil || p.NPCID == "" {
			return
		}
		go func() {
			f, err := os.Open(pipelinePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[npc-narrate] pipeline not found: %v\n", err)
				return
			}
			defer f.Close()

			pipe, err := pipeline.Load(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[npc-narrate] pipeline load error: %v\n", err)
				return
			}

			mgr := executor.NewManager()
			if errs := mgr.LoadWrappersFromDir(wrappersDir); len(errs) > 0 {
				for _, e := range errs {
					fmt.Fprintf(os.Stderr, "[npc-narrate] sidecar warn: %v\n", e)
				}
			}

			userInput := fmt.Sprintf("NPC: %s\nWhat they said: %q\nTrigger: %s\nPlayer stealth: %d",
				p.NPCName, p.Text, p.Trigger, p.StealthLevel)

			var opts []pipeline.RunOption
			opts = append(opts, pipeline.WithNoClarification())
			if st != nil {
				opts = append(opts, pipeline.WithRunStore(st))
				opts = append(opts, pipeline.WithBrainInjector(pipeline.NewStoreBrainInjector(st)))
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			result, err := pipeline.Run(ctx, pipe, mgr, userInput, opts...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[npc-narrate] run error: %v\n", err)
				return
			}
			if result != "" {
				narrationCh <- result
			}
		}()
	}
}

func logHandler() func(topic, payload string) {
	logDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "glitch")
	logPath := filepath.Join(logDir, "plugin-signals.log")
	return func(topic, payload string) {
		if err := os.MkdirAll(logDir, 0o755); err != nil {
			log.Printf("[WARN] signal_handlers: log handler: mkdir %s: %v", logDir, err)
			return
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			log.Printf("[WARN] signal_handlers: log handler: open %s: %v", logPath, err)
			return
		}
		defer f.Close()
		line := fmt.Sprintf("%s %s %s\n", time.Now().UTC().Format(time.RFC3339), topic, payload)
		if _, err := f.WriteString(line); err != nil {
			log.Printf("[WARN] signal_handlers: log handler: write %s: %v", logPath, err)
		}
	}
}
