package models

import (
	"time"

	"github.com/uptrace/bun"
)

type LinuxUserIDType string

const (
	LinuxUserIDTypeUID   LinuxUserIDType = "uid"
	LinuxUserIDTypeEUID  LinuxUserIDType = "euid"
	LinuxUserIDTypeSUID  LinuxUserIDType = "suid"
	LinuxUserIDTypeFSUID LinuxUserIDType = "fsuid"
	LinuxUserIDTypeGID   LinuxUserIDType = "gid"
	LinuxUserIDTypeEGID  LinuxUserIDType = "egid"
	LinuxUserIDTypeSGID  LinuxUserIDType = "sgid"
	LinuxUserIDTypeFSGID LinuxUserIDType = "fsgid"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	UID      string `bun:"uid,notnull,unique:machine_user" json:"uid,omitempty" jsonschema:"description=identifier of the user,example=1000,example=S-1-5-21-3623811015-3361044348-30300820-1013"`
	GID      string `bun:"gid" json:"gid,omitempty" jsonschema:"description=primary group identifier of the user,example=1000,example=S-1-5-21-3623811015-3361044348-30300820-513"`
	Name     string `bun:"name" json:"name,omitempty" jsonschema:"description=user's real or display name,example=jdoe,example=administrator,example=root"`
	Username string `bun:"username" json:"username,omitempty" jsonschema:"description=the login name"`
	Domain   string `bun:"domain" json:"domain,omitempty" jsonschema:"description=domain of the user (windows),example=WORKGROUP,example=CORP"`

	MachineID int64    `bun:"machine_id,notnull,unique:machine_user"`
	Machine   *Machine `bun:"rel:belongs-to,join:machine_id=id,on_delete:cascade"`
}

type UserApplication struct {
	bun.BaseModel `bun:"table:user_applications"`

	ID        int64     `bun:"id,pk,autoincrement"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp"`

	UserID        int64        `bun:"user_id,notnull,unique:user_application"`
	User          *User        `bun:"rel:belongs-to,join:user_id=id,on_delete:cascade"`
	ApplicationID int64        `bun:"application_id,notnull,unique:user_application"`
	Application   *Application `bun:"rel:belongs-to,join:application_id=id,on_delete:cascade"`

	Linux string `bun:"linux" json:"linux,omitempty" jsonschema:"description=Linux-specific data for this user-application relation,example=suid:0"`
}
