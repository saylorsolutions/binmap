package cli

import (
	"errors"
	"fmt"
)

type subcommands struct {
	cmds    map[string]*Command
	aliases map[string]string
}

func (s *subcommands) initMaps() {
	if s.cmds == nil {
		s.cmds = map[string]*Command{}
	}
	if s.aliases == nil {
		s.aliases = map[string]string{}
	}
}

func (s *subcommands) addSubCommand(name string, cmd *Command, aliases ...string) error {
	s.initMaps()
	_name := cleanseName(name)
	if len(_name) == 0 {
		return errors.New("empty command name")
	}
	_, ok := s.cmds[_name]
	if ok {
		return fmt.Errorf("command '%s' is already registered", _name)
	}
	for i, alias := range aliases {
		_alias := cleanseName(alias)
		if len(_alias) == 0 {
			return fmt.Errorf("invalid alias '%s'", alias)
		}
		aliases[i] = _alias
		bound, ok := s.aliases[_alias]
		if ok {
			return fmt.Errorf("alias '%s' is already bound to command '%s'", _alias, bound)
		}
		s.aliases[_alias] = _name
	}
	s.cmds[_name] = cmd
	cmd.aliases = aliases
	return nil
}

func (s *subcommands) resolveCommand(name string) (*Command, bool) {
	s.initMaps()
	if len(s.cmds) == 0 {
		return nil, false
	}
	name = cleanseName(name)
	if len(name) == 0 {
		return nil, false
	}
	cmd, ok := s.cmds[name]
	if !ok {
		aliased, ok := s.aliases[name]
		if !ok {
			return nil, false
		}
		cmd, ok := s.cmds[aliased]
		return cmd, ok
	}
	return cmd, true
}
