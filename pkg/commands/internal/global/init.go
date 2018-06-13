// Copyright Joyent, Inc.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package global contains commands that operate on strucutres in the global
// domain, rather than a workspace. API "global admin" access level is required
// for these commands.
package global

import (
	"github.com/jawher/mow.cli"
	"github.com/joyent/conch-shell/pkg/util"
	uuid "gopkg.in/satori/go.uuid.v1"
)

// GdcUUID is the UUID of the global datacenter provided by the user
var GdcUUID uuid.UUID

// GRoomUUID is the UUID of the global datacenter provided by the user
var GRoomUUID uuid.UUID

// Init loads up the commands
func Init(app *cli.Cli) {
	app.Command(
		"global",
		"Execute commands against objects in the global space. Admin access is required.",
		func(cmd *cli.Cmd) {
			cmd.Before = util.BuildAPIAndVerifyLogin

			cmd.Command(
				"datacenters dcs",
				"Operate on all datacenters",
				func(dcs *cli.Cmd) {
					dcs.Command(
						"get",
						"Get all datacenters",
						dcGetAll,
					)

					dcs.Command(
						"create",
						"Create a datacenter",
						dcCreate,
					)
				},
			)

			cmd.Command(
				"datacenter dc",
				"Operate on individual datacenters",
				func(dc *cli.Cmd) {
					var gdcIDStr = dc.StringArg("ID", "", "The UUID of the datacenter")

					dc.Spec = "ID"
					dc.Before = func() {
						id, err := uuid.FromString(*gdcIDStr)
						if err != nil {
							util.Bail(err)
						}
						GdcUUID = id
					}

					dc.Command(
						"get",
						"Get a datacenter",
						dcGet,
					)

					dc.Command(
						"delete rm",
						"Delete a datacenter",
						dcDelete,
					)

					dc.Command(
						"update",
						"Update a datacenter",
						dcUpdate,
					)

					dc.Command(
						"rooms",
						"Get all rooms assigned to a datacenter",
						dcGetAllRooms,
					)
				},
			)
			/////////////////////////////////
			cmd.Command(
				"rooms rs",
				"Operate on all rooms",
				func(rs *cli.Cmd) {
					rs.Command(
						"get",
						"Get all rooms",
						roomGetAll,
					)

					rs.Command(
						"create",
						"Create a room",
						roomCreate,
					)
				},
			)

			cmd.Command(
				"room r",
				"Operate on individual rooms",
				func(r *cli.Cmd) {
					var roomIDStr = r.StringArg("ID", "", "The UUID of the room")

					r.Spec = "ID"
					r.Before = func() {
						id, err := uuid.FromString(*roomIDStr)
						if err != nil {
							util.Bail(err)
						}
						GRoomUUID = id
					}

					r.Command(
						"get",
						"Get a room",
						roomGet,
					)

					r.Command(
						"delete rm",
						"Delete a room",
						roomDelete,
					)

					r.Command(
						"update",
						"Update a room",
						roomUpdate,
					)
				},
			)
		},
	)
}
