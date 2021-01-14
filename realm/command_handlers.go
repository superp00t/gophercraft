package realm

// CommandPrivileges a bitmask of privileges to be compared with Command.Requires.
// Any player with sys.Tier_Admin tier can bypass this check.
type CommandPrivileges uint8

// No bits. This means only people who can bypass privilege checks can use this command.
const Admin CommandPrivileges = 0

const (
	Player CommandPrivileges = 1 << iota
	GameMaster
	PhaseBuilder
	PhaseOwner
)

type Command struct {
	Requires  CommandPrivileges
	Signature string
	// Can be multiline
	Description string
	Function    interface{}
}

func (ws *Server) Cmd(req CommandPrivileges, sig, description string, function interface{}) {
	ws.CommandHandlers = append(ws.CommandHandlers, Command{req, sig, description, function})
}

const (
	HelpColor = "FF50c41a"
	DemoColor = "FF99CCFF"
)
