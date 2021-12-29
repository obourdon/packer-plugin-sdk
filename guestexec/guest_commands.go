package guestexec

import (
	"fmt"
	"log"
	"strings"
)

const UnixOSType = "unix"
const WindowsOSType = "windows"
const DefaultOSType = UnixOSType

type guestOSTypeCommand struct {
	chmod     string
	mkdir     string
	removeDir string
	statPath  string
	mv        string
}

var guestOSTypeCommands = map[string]guestOSTypeCommand{
	UnixOSType: {
		chmod:     "chmod %s '%s'",
		mkdir:     "mkdir -p '%s'",
		removeDir: "rm -rf '%s'",
		statPath:  "stat '%s'",
		mv:        "mv '%s' '%s'",
	},
	WindowsOSType: {
		chmod:     "echo 'skipping chmod %s %s'", // no-op
		mkdir:     "powershell.exe -Command \"New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path %s\"",
		removeDir: "powershell.exe -Command \"rm %s -recurse -force\"",
		statPath:  "powershell.exe -Command { if (test-path %s) { exit 0 } else { exit 1 } }",
		mv:        "powershell.exe -Command \"mv %s %s -force\"",
	},
}

type GuestCommands struct {
	GuestOSType string
	Sudo        bool
}

func NewGuestCommands(osType string, sudo bool) (*GuestCommands, error) {
	_, ok := guestOSTypeCommands[osType]
	if !ok {
		return nil, fmt.Errorf("Invalid osType: \"%s\"", osType)
	}
	return &GuestCommands{GuestOSType: osType, Sudo: sudo}, nil
}

func (g *GuestCommands) Chmod(path string, mode string) string {
	log.Printf("[ERROR] OLIVIER Command Chmod mode [%s] path [%s] escaped [%s]", mode, path, g.escapePath(path))
	log.Printf("[ERROR] OLIVIER Command Chmod [%s]", fmt.Sprintf(g.commands().chmod, mode, g.escapePath(path)))
	log.Printf("[ERROR] OLIVIER Command Chmod sudoed [%s]", g.sudo(fmt.Sprintf(g.commands().chmod, mode, g.escapePath(path))))
	return g.sudo(fmt.Sprintf(g.commands().chmod, mode, g.escapePath(path)))
}

func (g *GuestCommands) CreateDir(path string) string {
	return g.sudo(fmt.Sprintf(g.commands().mkdir, g.escapePath(path)))
}

func (g *GuestCommands) RemoveDir(path string) string {
	return g.sudo(fmt.Sprintf(g.commands().removeDir, g.escapePath(path)))
}

func (g *GuestCommands) commands() guestOSTypeCommand {
	return guestOSTypeCommands[g.GuestOSType]
}

func (g *GuestCommands) escapePath(path string) string {
	if g.GuestOSType == WindowsOSType {
		return strings.Replace(path, " ", "` ", -1)
	}
	return path
}

func (g *GuestCommands) StatPath(path string) string {
	return g.sudo(fmt.Sprintf(g.commands().statPath, g.escapePath(path)))
}

func (g *GuestCommands) MovePath(srcPath string, dstPath string) string {
	return g.sudo(fmt.Sprintf(g.commands().mv, g.escapePath(srcPath), g.escapePath(dstPath)))
}

func (g *GuestCommands) sudo(cmd string) string {
	if g.GuestOSType == UnixOSType && g.Sudo {
		return "sudo " + cmd
	}
	return cmd
}
