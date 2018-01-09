// Copyright 2017 Joyent, Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package commands is the parent that loads up the full command set
package commands

import (
	"github.com/joyent/conch-shell/pkg/commands/internal/devices"
	"github.com/joyent/conch-shell/pkg/commands/internal/profile"
	"github.com/joyent/conch-shell/pkg/commands/internal/reports"
	"github.com/joyent/conch-shell/pkg/commands/internal/user"
	"github.com/joyent/conch-shell/pkg/commands/internal/workspaces"
	"gopkg.in/jawher/mow.cli.v1"
)

// Init loads up all the commands
func Init(app *cli.Cli) {
	profile.Init(app)
	workspaces.Init(app)
	devices.Init(app)
	user.Init(app)
	reports.Init(app)
}
