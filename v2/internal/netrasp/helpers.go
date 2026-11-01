package netrasp

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// enableAnyPrompt matches either the "Password:" prompt shown when 'enable'
// requires a secret, or an already-privileged prompt (hostname#). IOS reports
// privilege level 15 with a trailing '#' regardless of the configured hostname,
// so a '#' means the device is already enabled and no secret is required.
var enableAnyPrompt = regexp.MustCompile(`^(?:[Pp]assword:|[a-zA-Z0-9_-]+#)`)

// enablePrivileges performs the 'enable' dialog. If the device asks for a
// password it sends the configured enable secret; devices already at privilege
// level 15 do not prompt for a password, so in that case nothing is sent and
// the connection is considered already privileged.
func enablePrivileges(ctx context.Context, c connection, enableSecret string) error {
	if err := c.Send(ctx, "enable"); err != nil {
		return fmt.Errorf("unable to send command to device: %w", err)
	}

	reader := c.Recv(ctx)
	output, err := readUntilPrompt(ctx, reader, enableAnyPrompt)
	if err != nil {
		return err
	}

	// No "Password:" prompt means the device was already privileged (or enable
	// requires no secret): there is nothing to send.
	if !strings.Contains(output, "Password:") {
		return nil
	}

	if err := c.Send(ctx, enableSecret); err != nil {
		return fmt.Errorf("unable to send command to device: %w", err)
	}
	reader = c.Recv(ctx)
	_, err = readUntilPrompt(ctx, reader, generalPrompt)

	return err
}

func establishConnection(ctx context.Context, p Platform, c connection, prompt *regexp.Regexp, preparationCommands []string) error {
	err := c.Dial(ctx)
	if err != nil {
		return fmt.Errorf("unable to open connection: %w", err)
	}

	reader := c.Recv(ctx)
	// Make sure that we find the initial prompt to clear the buffer before we continue
	_, err = readUntilPrompt(ctx, reader, prompt)
	if err != nil {
		return fmt.Errorf("unable to find the initial prompt: %w", err)
	}

	for _, command := range preparationCommands {
		output, err := p.Run(ctx, command)
		if err != nil {
			return fmt.Errorf("unable to prepare session using the command '%s': %w", output, err)
		}
	}

	return nil
}
